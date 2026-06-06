package op

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/llmname"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"gorm.io/gorm"
)

const defaultUpstreamRefreshIntervalSecs = 12 * 60 * 60

func normalizeUpstreamRefreshInterval(value int) int {
	if value <= 0 {
		return defaultUpstreamRefreshIntervalSecs
	}
	if value < 3600 {
		return 3600
	}
	return value
}

func encryptUpstreamSecret(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	secret, err := SettingGetString(model.SettingKeyAuthTokenSecret)
	if err != nil {
		return "", err
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	payload := append(nonce, gcm.Seal(nil, nonce, []byte(value), nil)...)
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decryptUpstreamSecret(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	payload, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	secret, err := SettingGetString(model.SettingKeyAuthTokenSecret)
	if err != nil {
		return "", err
	}
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(payload) < gcm.NonceSize() {
		return "", fmt.Errorf("invalid upstream credential")
	}
	nonce, ciphertext := payload[:gcm.NonceSize()], payload[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func UpstreamSiteList(ctx context.Context) ([]model.UpstreamSite, error) {
	var sites []model.UpstreamSite
	if err := db.GetDB().WithContext(ctx).Order("updated_at desc, id desc").Find(&sites).Error; err != nil {
		return nil, err
	}
	return sites, nil
}

func UpstreamSiteDetailGet(ctx context.Context, id int) (model.UpstreamSiteDetail, error) {
	var site model.UpstreamSite
	if err := db.GetDB().WithContext(ctx).First(&site, id).Error; err != nil {
		return model.UpstreamSiteDetail{}, fmt.Errorf("upstream site not found")
	}
	detail := model.UpstreamSiteDetail{Site: site}
	if err := db.GetDB().WithContext(ctx).Where("upstream_site_id = ?", id).Order("id asc").Find(&detail.Credentials).Error; err != nil {
		return detail, err
	}
	redactUpstreamCredentials(detail.Credentials)
	if err := db.GetDB().WithContext(ctx).Where("upstream_site_id = ?", id).Order("id asc").Find(&detail.Keys).Error; err != nil {
		return detail, err
	}
	if err := db.GetDB().WithContext(ctx).Where("upstream_site_id = ?", id).Order("name asc").Find(&detail.Groups).Error; err != nil {
		return detail, err
	}
	if err := db.GetDB().WithContext(ctx).Where("upstream_site_id = ?", id).Order("model_name asc").Find(&detail.Prices).Error; err != nil {
		return detail, err
	}
	for i := range detail.Prices {
		supported, _, _ := model.InferCacheSupport(model.LLMInfo{
			Name:          detail.Prices[i].ModelName,
			CanonicalName: detail.Prices[i].CanonicalName,
			LLMPrice:      model.LLMPrice{Input: detail.Prices[i].Input, Output: detail.Prices[i].Output, CacheRead: detail.Prices[i].CacheRead, CacheWrite: detail.Prices[i].CacheWrite},
			CachePolicy:   model.CachePolicy(detail.Prices[i].CachePolicy),
			CacheReason:   detail.Prices[i].CacheReason,
		})
		detail.Prices[i].CacheSupported = &supported
	}
	if site.LinkedChannelID > 0 {
		if channel, err := ChannelGet(site.LinkedChannelID, ctx); err == nil {
			summary := appliedChannelSummary(*channel)
			detail.LinkedChannel = &summary
		}
	}
	return detail, nil
}

func redactUpstreamCredentials(credentials []model.UpstreamCredential) {
	for i := range credentials {
		credentials[i].EncryptedValue = ""
	}
}

func UpstreamSiteCreate(ctx context.Context, req model.UpstreamSiteCreateRequest) (model.UpstreamSiteDetail, error) {
	inspectReq := upstreamInspectRequestFromCreate(req)
	inspect, err := InspectUpstreamGateway(ctx, inspectReq)
	if err != nil {
		return model.UpstreamSiteDetail{}, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = defaultUpstreamChannelName(inspect)
	}
	site := model.UpstreamSite{
		Name:                name,
		ProviderType:        inspect.ProviderType,
		BaseURL:             inspect.BaseURL,
		APIBaseURL:          inspect.APIBaseURL,
		AuthMode:            inspect.AuthMode,
		Enabled:             true,
		AutoRefresh:         req.AutoRefresh,
		RefreshIntervalSecs: normalizeUpstreamRefreshInterval(req.RefreshIntervalSecs),
		SyncToChannel:       req.SyncToChannel,
		SourceLabel:         inspect.SourceLabel,
	}
	if err := db.GetDB().WithContext(ctx).Create(&site).Error; err != nil {
		return model.UpstreamSiteDetail{}, err
	}
	if err := saveUpstreamCredentials(ctx, site.ID, inspect, inspectReq); err != nil {
		return model.UpstreamSiteDetail{}, err
	}
	if req.SyncToChannel || req.TargetChannelID > 0 {
		applied, err := ApplyUpstreamGateway(ctx, model.UpstreamApplyRequest{
			Inspect:         inspectReq,
			TargetChannelID: req.TargetChannelID,
			UpstreamSiteID:  site.ID,
			ChannelName:     firstNonEmptyUpstreamValue(req.ChannelName, site.Name),
			AppendKeys:      upstreamBoolPtr(true),
			OverwriteModels: upstreamBoolPtr(true),
			EnableChannel:   upstreamBoolPtr(true),
		})
		if err != nil {
			return model.UpstreamSiteDetail{}, err
		}
		site.LinkedChannelID = applied.Summary.ID
		if err := db.GetDB().WithContext(ctx).Model(&model.UpstreamSite{}).Where("id = ?", site.ID).Updates(map[string]any{
			"linked_channel_id": site.LinkedChannelID,
		}).Error; err != nil {
			return model.UpstreamSiteDetail{}, err
		}
	}
	if err := replaceUpstreamSnapshots(ctx, site, inspect, model.UpstreamRefreshManual, ""); err != nil {
		return model.UpstreamSiteDetail{}, err
	}
	return UpstreamSiteDetailGet(ctx, site.ID)
}

func UpstreamSiteUpdate(ctx context.Context, req model.UpstreamSiteUpdateRequest) (model.UpstreamSite, error) {
	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.AutoRefresh != nil {
		updates["auto_refresh"] = *req.AutoRefresh
	}
	if req.RefreshIntervalSecs != nil {
		updates["refresh_interval_secs"] = normalizeUpstreamRefreshInterval(*req.RefreshIntervalSecs)
	}
	if req.SyncToChannel != nil {
		updates["sync_to_channel"] = *req.SyncToChannel
	}
	if req.LinkedChannelID != nil {
		updates["linked_channel_id"] = *req.LinkedChannelID
	}
	if len(updates) == 0 {
		var site model.UpstreamSite
		if err := db.GetDB().WithContext(ctx).First(&site, req.ID).Error; err != nil {
			return site, fmt.Errorf("upstream site not found")
		}
		return site, nil
	}
	if err := db.GetDB().WithContext(ctx).Model(&model.UpstreamSite{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
		return model.UpstreamSite{}, err
	}
	var site model.UpstreamSite
	if err := db.GetDB().WithContext(ctx).First(&site, req.ID).Error; err != nil {
		return site, err
	}
	return site, nil
}

func UpstreamSiteDelete(ctx context.Context, id int) error {
	return db.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, table := range []any{
			&model.UpstreamCredential{},
			&model.UpstreamKeySnapshot{},
			&model.UpstreamGroupSnapshot{},
			&model.UpstreamModelPrice{},
		} {
			if err := tx.Where("upstream_site_id = ?", id).Delete(table).Error; err != nil {
				return err
			}
		}
		if err := tx.Delete(&model.UpstreamSite{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

func UpstreamInspect(ctx context.Context, req model.UpstreamInspectRequest) (model.UpstreamInspectResult, error) {
	return InspectUpstreamGateway(ctx, req)
}

func UpstreamSiteRefresh(ctx context.Context, req model.UpstreamRefreshRequest) (model.UpstreamSiteDetail, error) {
	var site model.UpstreamSite
	if err := db.GetDB().WithContext(ctx).First(&site, req.ID).Error; err != nil {
		return model.UpstreamSiteDetail{}, fmt.Errorf("upstream site not found")
	}
	inspectReq, err := inspectRequestFromStoredUpstream(ctx, site)
	if err != nil {
		return model.UpstreamSiteDetail{}, err
	}
	inspect, err := InspectUpstreamGateway(ctx, inspectReq)
	if err != nil {
		_ = markUpstreamRefreshFailed(ctx, site.ID, err.Error())
		return model.UpstreamSiteDetail{}, err
	}
	if (req.ApplyChannel || site.SyncToChannel) && site.Enabled {
		applied, err := ApplyUpstreamGateway(ctx, model.UpstreamApplyRequest{
			Inspect:         inspectReq,
			TargetChannelID: site.LinkedChannelID,
			UpstreamSiteID:  site.ID,
			ChannelName:     site.Name,
			AppendKeys:      upstreamBoolPtr(true),
			OverwriteModels: upstreamBoolPtr(false),
			EnableChannel:   upstreamBoolPtr(true),
		})
		if err != nil {
			_ = markUpstreamRefreshFailed(ctx, site.ID, err.Error())
			return model.UpstreamSiteDetail{}, err
		}
		site.LinkedChannelID = applied.Summary.ID
	}
	mode := model.UpstreamRefreshScheduled
	if req.Manual {
		mode = model.UpstreamRefreshManual
	}
	if err := replaceUpstreamSnapshots(ctx, site, inspect, mode, ""); err != nil {
		return model.UpstreamSiteDetail{}, err
	}
	return UpstreamSiteDetailGet(ctx, site.ID)
}

func UpstreamSiteApply(ctx context.Context, id int, targetChannelID int) (model.UpstreamApplyResult, error) {
	var site model.UpstreamSite
	if err := db.GetDB().WithContext(ctx).First(&site, id).Error; err != nil {
		return model.UpstreamApplyResult{}, fmt.Errorf("upstream site not found")
	}
	inspectReq, err := inspectRequestFromStoredUpstream(ctx, site)
	if err != nil {
		return model.UpstreamApplyResult{}, err
	}
	if targetChannelID <= 0 {
		targetChannelID = site.LinkedChannelID
	}
	applied, err := ApplyUpstreamGateway(ctx, model.UpstreamApplyRequest{
		Inspect:         inspectReq,
		TargetChannelID: targetChannelID,
		UpstreamSiteID:  site.ID,
		ChannelName:     site.Name,
		AppendKeys:      upstreamBoolPtr(true),
		OverwriteModels: upstreamBoolPtr(true),
		EnableChannel:   upstreamBoolPtr(true),
	})
	if err != nil {
		return model.UpstreamApplyResult{}, err
	}
	if err := db.GetDB().WithContext(ctx).Model(&model.UpstreamSite{}).Where("id = ?", id).Updates(map[string]any{
		"linked_channel_id": applied.Summary.ID,
		"sync_to_channel":   true,
	}).Error; err != nil {
		return model.UpstreamApplyResult{}, err
	}
	return applied, nil
}

func upstreamInspectRequestFromCreate(req model.UpstreamSiteCreateRequest) model.UpstreamInspectRequest {
	return model.UpstreamInspectRequest{
		ProviderType: req.ProviderType,
		BaseURL:      req.BaseURL,
		AuthMode:     req.AuthMode,
		Token:        req.Token,
		AccessKey:    req.AccessKey,
		UserID:       req.UserID,
		Username:     req.Username,
		Password:     req.Password,
	}
}

func inspectRequestFromStoredUpstream(ctx context.Context, site model.UpstreamSite) (model.UpstreamInspectRequest, error) {
	var credentials []model.UpstreamCredential
	if err := db.GetDB().WithContext(ctx).Where("upstream_site_id = ?", site.ID).Order("credential_type asc, id asc").Find(&credentials).Error; err != nil {
		return model.UpstreamInspectRequest{}, err
	}
	req := model.UpstreamInspectRequest{
		ProviderType: site.ProviderType,
		BaseURL:      site.BaseURL,
		AuthMode:     model.UpstreamAuthModeToken,
		UserID:       firstStoredUserID(credentials),
	}
	for _, credential := range credentials {
		secret, err := decryptUpstreamSecret(credential.EncryptedValue)
		if err != nil {
			return model.UpstreamInspectRequest{}, err
		}
		switch credential.CredentialType {
		case model.UpstreamCredentialManagementToken:
			req.Token = secret
			req.AuthMode = model.UpstreamAuthModeToken
			return req, nil
		case model.UpstreamCredentialAccessKey:
			if req.AccessKey == "" {
				req.AccessKey = secret
				req.AuthMode = model.UpstreamAuthModeAccessKey
			}
		}
	}
	if req.AccessKey != "" {
		return req, nil
	}
	return model.UpstreamInspectRequest{}, fmt.Errorf("upstream site has no saved token or access key")
}

func firstStoredUserID(credentials []model.UpstreamCredential) string {
	for _, credential := range credentials {
		if strings.TrimSpace(credential.UserID) != "" {
			return strings.TrimSpace(credential.UserID)
		}
	}
	return ""
}

func normalizeUpstreamModelList(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		name := strings.ToLower(strings.TrimSpace(value))
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func normalizeUpstreamRequestCapabilities(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		capability := model.NormalizeRequestCapability(value)
		if capability == "" {
			continue
		}
		if _, ok := seen[capability]; ok {
			continue
		}
		seen[capability] = struct{}{}
		out = append(out, capability)
	}
	sort.Strings(out)
	return out
}

func saveUpstreamCredentials(ctx context.Context, siteID int, inspect model.UpstreamInspectResult, req model.UpstreamInspectRequest) error {
	credentials := make([]model.UpstreamCredential, 0, 2)
	add := func(kind, authMode, displayName, raw, userID string, importable bool) error {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil
		}
		encrypted, err := encryptUpstreamSecret(raw)
		if err != nil {
			return err
		}
		credentials = append(credentials, model.UpstreamCredential{
			UpstreamSiteID:  siteID,
			CredentialType:  kind,
			AuthMode:        authMode,
			DisplayName:     displayName,
			MaskedValue:     maskSecret(raw),
			EncryptedValue:  encrypted,
			UserID:          strings.TrimSpace(userID),
			Importable:      importable,
			LastValidatedAt: time.Now(),
		})
		return nil
	}
	if err := add(model.UpstreamCredentialManagementToken, model.UpstreamAuthModeToken, "管理 Token", firstNonEmptyUpstreamValue(inspect.ManagementToken, req.Token), req.UserID, false); err != nil {
		return err
	}
	if err := add(model.UpstreamCredentialAccessKey, model.UpstreamAuthModeAccessKey, "网关 Key", firstNonEmptyUpstreamValue(inspect.GatewayAccessKey, req.AccessKey), req.UserID, true); err != nil {
		return err
	}
	if len(credentials) == 0 {
		return fmt.Errorf("upstream site has no credential to save")
	}
	return db.GetDB().WithContext(ctx).Create(&credentials).Error
}

func replaceUpstreamSnapshots(ctx context.Context, site model.UpstreamSite, inspect model.UpstreamInspectResult, refreshMode, message string) error {
	return db.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, table := range []any{&model.UpstreamKeySnapshot{}, &model.UpstreamGroupSnapshot{}, &model.UpstreamModelPrice{}} {
			if err := tx.Where("upstream_site_id = ?", site.ID).Delete(table).Error; err != nil {
				return err
			}
		}
		keyRows := make([]model.UpstreamKeySnapshot, 0, len(inspect.Keys))
		for _, key := range inspect.Keys {
			keyRows = append(keyRows, model.UpstreamKeySnapshot{
				UpstreamSiteID:      site.ID,
				Name:                key.Name,
				MaskedKey:           key.MaskedKey,
				AllowedModels:       strings.Join(normalizeUpstreamModelList(key.AllowedModels), ","),
				RequestCapabilities: strings.Join(normalizeUpstreamRequestCapabilities(key.RequestCapabilities), ","),
				Groups:              strings.Join(compactStrings(key.Groups), ","),
				Status:              key.Status,
				Quota:               key.Quota,
				QuotaUsed:           key.QuotaUsed,
				ExpiresAt:           key.ExpiresAt,
				Importable:          key.Importable,
				SourceType:          key.SourceType,
			})
		}
		if len(keyRows) > 0 {
			if err := tx.Create(&keyRows).Error; err != nil {
				return err
			}
		}
		groupRows := make([]model.UpstreamGroupSnapshot, 0, len(inspect.Groups))
		for _, group := range inspect.Groups {
			groupRows = append(groupRows, model.UpstreamGroupSnapshot{
				UpstreamSiteID:      site.ID,
				ExternalID:          group.ID,
				Name:                group.Name,
				Description:         group.Description,
				Platform:            group.Platform,
				Status:              group.Status,
				RateMultiplier:      group.RateMultiplier,
				Models:              strings.Join(normalizeUpstreamModelList(group.Models), ","),
				RequestCapabilities: strings.Join(normalizeUpstreamRequestCapabilities(group.RequestCapabilities), ","),
				Source:              group.Source,
			})
		}
		if len(groupRows) > 0 {
			if err := tx.Create(&groupRows).Error; err != nil {
				return err
			}
		}
		priceRows := make([]model.UpstreamModelPrice, 0, len(inspect.PriceCandidates))
		for _, candidate := range inspect.PriceCandidates {
			if candidate.LLMPrice.IsZero() {
				continue
			}
			priceRows = append(priceRows, model.UpstreamModelPrice{
				UpstreamSiteID:  site.ID,
				ChannelID:       site.LinkedChannelID,
				ModelName:       strings.ToLower(strings.TrimSpace(candidate.Name)),
				CanonicalName:   candidate.CanonicalName,
				PriceSource:     firstNonEmptyUpstreamValue(candidate.PriceSource, model.UpstreamPriceSourceGateway),
				PriceMatchedKey: candidate.PriceMatchedKey,
				SourceLabel:     inspect.SourceLabel,
				CachePolicy:     candidate.CachePolicy,
				CacheReason:     candidate.CacheReason,
				Input:           candidate.Input,
				Output:          candidate.Output,
				CacheRead:       candidate.CacheRead,
				CacheWrite:      candidate.CacheWrite,
			})
		}
		if len(priceRows) > 0 {
			if err := tx.Create(&priceRows).Error; err != nil {
				return err
			}
		}
		statusMessage := firstNonEmptyUpstreamValue(message, refreshMode)
		if err := tx.Model(&model.UpstreamSite{}).Where("id = ?", site.ID).Updates(map[string]any{
			"api_base_url":         inspect.APIBaseURL,
			"auth_mode":            inspect.AuthMode,
			"source_label":         inspect.SourceLabel,
			"linked_channel_id":    site.LinkedChannelID,
			"last_refresh_at":      time.Now(),
			"last_refresh_status":  "success",
			"last_refresh_message": statusMessage,
			"model_count":          inspect.ModelCount,
			"key_count":            len(inspect.Keys),
			"group_count":          len(inspect.Groups),
			"price_count":          len(priceRows),
			"subscription_count":   len(inspect.Subscriptions),
			"balance_available":    inspect.TokenUsage.Available,
			"balance_used":         inspect.TokenUsage.UsedQuota,
			"balance_remain":       inspect.TokenUsage.RemainQuota,
			"balance_unlimited":    inspect.TokenUsage.Unlimited,
		}).Error; err != nil {
			return err
		}
		return nil
	})
}

func markUpstreamRefreshFailed(ctx context.Context, id int, message string) error {
	return db.GetDB().WithContext(ctx).Model(&model.UpstreamSite{}).Where("id = ?", id).Updates(map[string]any{
		"last_refresh_at":      time.Now(),
		"last_refresh_status":  "failed",
		"last_refresh_message": message,
	}).Error
}

func upstreamBoolPtr(value bool) *bool {
	return &value
}

func UpstreamModelPriceList(ctx context.Context) ([]model.UpstreamModelPrice, error) {
	var prices []model.UpstreamModelPrice
	if err := db.GetDB().WithContext(ctx).Order("model_name asc, updated_at desc").Find(&prices).Error; err != nil {
		return nil, err
	}
	return prices, nil
}

func UpstreamPriceSummaries(ctx context.Context) ([]model.UpstreamPriceSummary, error) {
	infos, err := LLMList(ctx)
	if err != nil {
		return nil, err
	}
	prices, err := UpstreamModelPriceList(ctx)
	if err != nil {
		return nil, err
	}
	byModel := make(map[string][]model.UpstreamModelPrice, len(prices))
	for _, priceRow := range prices {
		key := strings.ToLower(strings.TrimSpace(priceRow.ModelName))
		if key == "" {
			key = strings.ToLower(strings.TrimSpace(priceRow.CanonicalName))
		}
		byModel[key] = append(byModel[key], priceRow)
		canonical := strings.ToLower(strings.TrimSpace(priceRow.CanonicalName))
		if canonical != "" && canonical != key {
			byModel[canonical] = append(byModel[canonical], priceRow)
		}
	}
	out := make([]model.UpstreamPriceSummary, 0, len(infos))
	for _, info := range infos {
		keys := llmname.CandidateModelKeys(info.Name)
		var gatewayPrices []model.UpstreamModelPrice
		seen := map[int]struct{}{}
		for _, key := range keys {
			for _, row := range byModel[strings.ToLower(strings.TrimSpace(key))] {
				if _, ok := seen[row.ID]; ok {
					continue
				}
				seen[row.ID] = struct{}{}
				gatewayPrices = append(gatewayPrices, row)
			}
		}
		sort.Slice(gatewayPrices, func(i, j int) bool {
			if gatewayPrices[i].UpdatedAt.Equal(gatewayPrices[j].UpdatedAt) {
				return gatewayPrices[i].ID > gatewayPrices[j].ID
			}
			return gatewayPrices[i].UpdatedAt.After(gatewayPrices[j].UpdatedAt)
		})
		summary := model.UpstreamPriceSummary{
			ModelName:     info.Name,
			OfficialPrice: info.OfficialLLMPrice,
			GatewayPrices: gatewayPrices,
		}
		if len(gatewayPrices) > 0 {
			summary.EffectiveGateway = &gatewayPrices[0]
		}
		out = append(out, summary)
	}
	return out, nil
}

func ResolveGatewayLLMPrice(modelName string, channelID int) (model.LLMPrice, bool) {
	keys := llmname.CandidateModelKeys(modelName)
	if len(keys) == 0 {
		return model.LLMPrice{}, false
	}
	rows := make([]model.UpstreamModelPrice, 0)
	query := db.GetDB().Where("lower(model_name) IN ? OR lower(canonical_name) IN ?", keys, keys)
	if channelID > 0 {
		query = query.Where("channel_id = ? OR channel_id = 0", channelID)
	}
	if err := query.Order("channel_id desc, updated_at desc, id desc").Limit(1).Find(&rows).Error; err != nil || len(rows) == 0 {
		return model.LLMPrice{}, false
	}
	row := rows[0]
	price := model.LLMPrice{Input: row.Input, Output: row.Output, CacheRead: row.CacheRead, CacheWrite: row.CacheWrite}
	return price, !price.IsZero()
}
