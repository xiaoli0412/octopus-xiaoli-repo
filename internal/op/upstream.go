package op

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/db"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/llmname"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"gorm.io/gorm"
)

const defaultUpstreamRefreshIntervalSecs = 12 * 60 * 60
const defaultUpstreamCheckinIntervalSecs = 24 * 60 * 60
const maxUpstreamCheckinLogEntries = 50

const upstreamHealthMaxRefreshAge = 24 * time.Hour
const upstreamHealthErrorRateWindow = 1 * time.Hour
const upstreamHealthErrorRateThreshold = 0.5

func normalizeUpstreamRefreshInterval(value int) int {
	if value <= 0 {
		return defaultUpstreamRefreshIntervalSecs
	}
	if value < 3600 {
		return 3600
	}
	return value
}

func normalizeUpstreamCheckinInterval(value int) int {
	if value <= 0 {
		return defaultUpstreamCheckinIntervalSecs
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
		AutoCheckin:         req.AutoCheckin,
		CheckinIntervalSecs: normalizeUpstreamCheckinInterval(req.CheckinIntervalSecs),
		SyncToChannel:       req.SyncToChannel,
		AutoSyncGroup:       req.AutoSyncGroup,
		AutoSyncPrice:       req.AutoSyncPrice,
		AutoCreateKey:       req.AutoCreateKey,
		KeyQuotaLimit:       req.KeyQuotaLimit,
		KeyExpireDays:       req.KeyExpireDays,
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
			AutoCreateKey:   req.AutoCreateKey,
			KeyQuotaLimit:   req.KeyQuotaLimit,
			KeyExpireDays:   req.KeyExpireDays,
			AutoSyncGroup:   req.AutoSyncGroup,
			AutoSyncPrice:   req.AutoSyncPrice,
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
	if req.AutoCheckin != nil {
		updates["auto_checkin"] = *req.AutoCheckin
	}
	if req.CheckinIntervalSecs != nil {
		updates["checkin_interval_secs"] = normalizeUpstreamCheckinInterval(*req.CheckinIntervalSecs)
	}
	if req.SyncToChannel != nil {
		updates["sync_to_channel"] = *req.SyncToChannel
	}
	if req.AutoSyncGroup != nil {
		updates["auto_sync_group"] = *req.AutoSyncGroup
	}
	if req.AutoSyncPrice != nil {
		updates["auto_sync_price"] = *req.AutoSyncPrice
	}
	if req.LinkedChannelID != nil {
		updates["linked_channel_id"] = *req.LinkedChannelID
	}
	if req.BalanceAlertThreshold != nil {
		updates["balance_alert_threshold"] = *req.BalanceAlertThreshold
	}
	if req.AutoCreateKey != nil {
		updates["auto_create_key"] = *req.AutoCreateKey
	}
	if req.KeyQuotaLimit != nil {
		updates["key_quota_limit"] = *req.KeyQuotaLimit
	}
	if req.KeyExpireDays != nil {
		updates["key_expire_days"] = *req.KeyExpireDays
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
		UpstreamSiteSuppress(site.ID, "inspect_request_failed")
		_ = markUpstreamRefreshFailed(ctx, site.ID, err.Error())
		return model.UpstreamSiteDetail{}, err
	}
	inspect, err := InspectUpstreamGateway(ctx, inspectReq)
	if err != nil {
		UpstreamSiteSuppress(site.ID, "refresh_failed")
		_ = markUpstreamRefreshFailed(ctx, site.ID, err.Error())
		return model.UpstreamSiteDetail{}, err
	}
	shouldApply := req.ApplyChannel || site.SyncToChannel
	if !req.Manual && (site.AutoCreateKey || site.AutoSyncGroup || site.AutoSyncPrice) {
		shouldApply = true
	}
	if shouldApply && site.Enabled {
		applied, err := ApplyUpstreamGateway(ctx, model.UpstreamApplyRequest{
			Inspect:         inspectReq,
			TargetChannelID: site.LinkedChannelID,
			UpstreamSiteID:  site.ID,
			ChannelName:     site.Name,
			AppendKeys:      upstreamBoolPtr(true),
			OverwriteModels: upstreamBoolPtr(false),
			EnableChannel:   upstreamBoolPtr(true),
			AutoCreateKey:   site.AutoCreateKey,
			KeyQuotaLimit:   site.KeyQuotaLimit,
			KeyExpireDays:   site.KeyExpireDays,
			AutoSyncGroup:   site.AutoSyncGroup,
			AutoSyncPrice:   site.AutoSyncPrice,
		})
		if err != nil {
			UpstreamSiteSuppress(site.ID, "apply_failed")
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
	// Successful refresh clears any temporary suppression.
	UpstreamSiteRestore(site.ID)
	return UpstreamSiteDetailGet(ctx, site.ID)
}

func CreateUpstreamKey(ctx context.Context, req model.UpstreamCreateKeyRequest) (model.UpstreamCreateKeyResult, error) {
	var site model.UpstreamSite
	if err := db.GetDB().WithContext(ctx).First(&site, req.SiteID).Error; err != nil {
		return model.UpstreamCreateKeyResult{}, fmt.Errorf("upstream site not found")
	}
	inspectReq, err := inspectRequestFromStoredUpstream(ctx, site)
	if err != nil {
		return model.UpstreamCreateKeyResult{}, err
	}
	if strings.TrimSpace(inspectReq.Token) == "" {
		return model.UpstreamCreateKeyResult{}, fmt.Errorf("upstream site has no management token")
	}
	httpClient, err := newHealthHTTPClientNoProxy()
	if err != nil {
		return model.UpstreamCreateKeyResult{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, upstreamInspectTimeout)
	defer cancel()

	providerType := normalizeUpstreamProvider(site.ProviderType)
	var payload []byte
	var ok bool
	switch providerType {
	case model.UpstreamProviderSub2API:
		payload, ok = createSub2APIKey(ctx, httpClient, site.BaseURL, inspectReq.Token, req)
	case model.UpstreamProviderNewAPI:
		payload, ok = createNewAPIKey(ctx, httpClient, site.BaseURL, inspectReq.Token, inspectReq.UserID, req)
	default:
		return model.UpstreamCreateKeyResult{}, fmt.Errorf("upstream provider %s does not support key creation", providerType)
	}
	if !ok {
		return model.UpstreamCreateKeyResult{}, fmt.Errorf("failed to create upstream key")
	}
	key, name, maskedKey := parseCreatedUpstreamKey(payload)
	if key == "" {
		return model.UpstreamCreateKeyResult{}, fmt.Errorf("upstream key creation response did not contain a key")
	}
	return model.UpstreamCreateKeyResult{Name: name, Key: key, MaskedKey: maskedKey}, nil
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
		AutoCreateKey:   site.AutoCreateKey,
		KeyQuotaLimit:   site.KeyQuotaLimit,
		KeyExpireDays:   site.KeyExpireDays,
		AutoSyncGroup:   site.AutoSyncGroup,
		AutoSyncPrice:   site.AutoSyncPrice,
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

// CheckUpstreamBalance compares the latest refreshed balance of a site against
// its configured alert threshold and records the check timestamp/value. The
// balance value comes from the existing upstream inspection logic that is
// already executed during UpstreamSiteRefresh.
func CheckUpstreamBalance(ctx context.Context, siteID int) (model.UpstreamBalanceCheckResult, error) {
	var site model.UpstreamSite
	if err := db.GetDB().WithContext(ctx).First(&site, siteID).Error; err != nil {
		return model.UpstreamBalanceCheckResult{}, fmt.Errorf("upstream site not found")
	}

	result := model.UpstreamBalanceCheckResult{
		Threshold: site.BalanceAlertThreshold,
		Remain:    site.BalanceRemain,
	}

	if err := db.GetDB().WithContext(ctx).Model(&model.UpstreamSite{}).Where("id = ?", siteID).Updates(map[string]any{
		"last_balance_check_at": time.Now(),
		"last_balance_value":    site.BalanceRemain,
	}).Error; err != nil {
		return result, err
	}

	if !site.BalanceAvailable || site.BalanceUnlimited || site.BalanceAlertThreshold <= 0 {
		return result, nil
	}

	if site.BalanceRemain <= site.BalanceAlertThreshold {
		result.Alert = true
		result.Message = fmt.Sprintf("upstream site %q balance %.4f is below alert threshold %.4f", site.Name, site.BalanceRemain, site.BalanceAlertThreshold)
	}

	return result, nil
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

const upstreamCheckinTimeout = 15 * time.Second

func UpstreamSiteCheckin(ctx context.Context, id int) (model.UpstreamCheckinResult, error) {
	var site model.UpstreamSite
	if err := db.GetDB().WithContext(ctx).First(&site, id).Error; err != nil {
		return model.UpstreamCheckinResult{}, fmt.Errorf("upstream site not found")
	}
	if !site.Enabled {
		return model.UpstreamCheckinResult{}, fmt.Errorf("upstream site is disabled")
	}
	inspectReq, err := inspectRequestFromStoredUpstream(ctx, site)
	if err != nil {
		return model.UpstreamCheckinResult{}, err
	}
	token := strings.TrimSpace(inspectReq.Token)
	if token == "" {
		return model.UpstreamCheckinResult{}, fmt.Errorf("upstream site has no management token for checkin")
	}

	endpoint := "/api/user/checkin"
	if normalizeUpstreamProvider(site.ProviderType) == model.UpstreamProviderSub2API {
		endpoint = "/api/v1/user/checkin"
	}

	result, err := callUpstreamCheckin(ctx, site.BaseURL, endpoint, token)
	now := time.Now()
	if err != nil {
		UpstreamSiteSuppress(site.ID, "checkin_failed")
		_ = appendUpstreamCheckinLog(ctx, site.ID, false, 0, err.Error(), now)
		return model.UpstreamCheckinResult{Success: false, Message: err.Error(), At: now}, err
	}
	result.At = now
	_ = appendUpstreamCheckinLog(ctx, site.ID, result.Success, result.Amount, result.Message, now)
	UpstreamSiteRestore(site.ID)
	return result, nil
}

func callUpstreamCheckin(ctx context.Context, baseURL, endpoint, token string) (model.UpstreamCheckinResult, error) {
	httpClient, err := newHealthHTTPClientNoProxy()
	if err != nil {
		return model.UpstreamCheckinResult{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, upstreamCheckinTimeout)
	defer cancel()

	payload, ok := callUpstreamWithRetry(ctx, httpClient, http.MethodPost, strings.TrimRight(baseURL, "/")+endpoint, []byte{}, upstreamAuthModeBearer, token, "")
	if !ok {
		return model.UpstreamCheckinResult{}, fmt.Errorf("checkin request failed")
	}

	success, amount, message := parseUpstreamCheckinResponse(payload)
	if !success {
		return model.UpstreamCheckinResult{}, fmt.Errorf("%s", firstNonEmptyUpstreamValue(message, "checkin failed"))
	}
	return model.UpstreamCheckinResult{Success: true, Amount: amount, Message: firstNonEmptyUpstreamValue(message, "checkin succeeded")}, nil
}

func parseUpstreamCheckinResponse(payload []byte) (success bool, amount float64, message string) {
	var raw any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return false, 0, "invalid checkin response"
	}
	records := flattenObjectRecords(raw)
	if len(records) == 0 {
		return false, 0, "empty checkin response"
	}

	success = true
	for _, record := range records {
		if value, ok := boolField(record, "success"); ok {
			success = value
		}
		if amount == 0 {
			if value, ok := numberField(record, "data", "quota", "amount", "credit", "reward"); ok {
				amount = value
			}
		}
		if message == "" {
			if value, ok := stringField(record, "message", "msg", "error", "detail"); ok {
				message = value
			}
		}
	}
	return success, amount, message
}

func appendUpstreamCheckinLog(ctx context.Context, siteID int, success bool, amount float64, message string, at time.Time) error {
	var site model.UpstreamSite
	if err := db.GetDB().WithContext(ctx).First(&site, siteID).Error; err != nil {
		return err
	}
	logEntries := upstreamCheckinLogEntries(site.CheckinLog)
	logEntries = append(logEntries, model.UpstreamCheckinLogEntry{
		Success: success,
		Amount:  amount,
		Message: message,
		At:      at,
	})
	if len(logEntries) > maxUpstreamCheckinLogEntries {
		logEntries = logEntries[len(logEntries)-maxUpstreamCheckinLogEntries:]
	}
	logJSON, err := json.Marshal(logEntries)
	if err != nil {
		return err
	}
	return db.GetDB().WithContext(ctx).Model(&model.UpstreamSite{}).Where("id = ?", siteID).Updates(map[string]any{
		"last_checkin_at": at,
		"checkin_log":     string(logJSON),
	}).Error
}

func upstreamCheckinLogEntries(raw string) []model.UpstreamCheckinLogEntry {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var entries []model.UpstreamCheckinLogEntry
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		return nil
	}
	return entries
}

// upstream site suppression state for automatic multi-channel failover.
var (
	upstreamSuppressedMu     sync.RWMutex
	upstreamSuppressedSites  = make(map[int]time.Time) // siteID -> suppressedAt
	upstreamSuppressedReason = make(map[int]string)    // siteID -> reason
)

// UpstreamSiteSuppress marks an upstream site as temporarily unavailable for routing.
// The suppression is in-memory only and can be cleared by UpstreamSiteRestore or a successful refresh.
func UpstreamSiteSuppress(siteID int, reason string) {
	upstreamSuppressedMu.Lock()
	defer upstreamSuppressedMu.Unlock()
	upstreamSuppressedSites[siteID] = time.Now()
	if reason != "" {
		upstreamSuppressedReason[siteID] = reason
	}
}

// UpstreamSiteRestore clears the temporary suppression of an upstream site.
func UpstreamSiteRestore(siteID int) {
	upstreamSuppressedMu.Lock()
	defer upstreamSuppressedMu.Unlock()
	delete(upstreamSuppressedSites, siteID)
	delete(upstreamSuppressedReason, siteID)
}

// UpstreamSiteIsSuppressed reports whether an upstream site is currently suppressed.
func UpstreamSiteIsSuppressed(siteID int) bool {
	upstreamSuppressedMu.RLock()
	defer upstreamSuppressedMu.RUnlock()
	_, ok := upstreamSuppressedSites[siteID]
	return ok
}

// UpstreamSuppressedSiteIDs returns a snapshot of currently suppressed site IDs.
func UpstreamSuppressedSiteIDs() map[int]time.Time {
	upstreamSuppressedMu.RLock()
	defer upstreamSuppressedMu.RUnlock()
	out := make(map[int]time.Time, len(upstreamSuppressedSites))
	for k, v := range upstreamSuppressedSites {
		out[k] = v
	}
	return out
}

func upstreamSiteSuppressedReason(siteID int) string {
	upstreamSuppressedMu.RLock()
	defer upstreamSuppressedMu.RUnlock()
	return upstreamSuppressedReason[siteID]
}

// UpstreamSiteHealthList evaluates the health status of all upstream sites.
func UpstreamSiteHealthList(ctx context.Context) ([]model.UpstreamHealthItem, error) {
	sites, err := UpstreamSiteList(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]model.UpstreamHealthItem, 0, len(sites))
	for _, site := range sites {
		items = append(items, evaluateUpstreamSiteHealth(ctx, site))
	}
	return items, nil
}

// UpstreamSiteHealthGet evaluates the health status of a single upstream site.
func UpstreamSiteHealthGet(ctx context.Context, id int) (model.UpstreamHealthItem, error) {
	var site model.UpstreamSite
	if err := db.GetDB().WithContext(ctx).First(&site, id).Error; err != nil {
		return model.UpstreamHealthItem{}, fmt.Errorf("upstream site not found")
	}
	return evaluateUpstreamSiteHealth(ctx, site), nil
}

func evaluateUpstreamSiteHealth(ctx context.Context, site model.UpstreamSite) model.UpstreamHealthItem {
	item := model.UpstreamHealthItem{
		ID:                site.ID,
		Name:              site.Name,
		Enabled:           site.Enabled,
		LastRefreshAt:     site.LastRefreshAt,
		LastRefreshStatus: site.LastRefreshStatus,
		ModelCount:        site.ModelCount,
		BalanceAvailable:  site.BalanceAvailable,
		BalanceRemain:     site.BalanceRemain,
		BalanceUnlimited:  site.BalanceUnlimited,
		Suppressed:        UpstreamSiteIsSuppressed(site.ID),
	}

	reasons := []string{}
	if !site.Enabled {
		reasons = append(reasons, "site_disabled")
	}

	// Balance alert check.
	if site.BalanceAvailable && !site.BalanceUnlimited && site.BalanceAlertThreshold > 0 && site.BalanceRemain <= site.BalanceAlertThreshold {
		item.BalanceAlert = true
		reasons = append(reasons, fmt.Sprintf("balance_low_%.4f", site.BalanceRemain))
	}

	// Refresh freshness.
	refreshStale := false
	if site.LastRefreshAt.IsZero() {
		refreshStale = true
		reasons = append(reasons, "never_refreshed")
	} else if time.Since(site.LastRefreshAt) > upstreamHealthMaxRefreshAge {
		refreshStale = true
		reasons = append(reasons, "refresh_stale")
	}
	if site.LastRefreshStatus == "failed" {
		reasons = append(reasons, "last_refresh_failed")
	}

	// Model availability.
	noModels := site.ModelCount <= 0
	if noModels {
		reasons = append(reasons, "no_models")
	}

	// Error rate from recent relay logs for linked channels.
	errorRate := upstreamSiteErrorRate(ctx, site.ID)
	item.ErrorRate = errorRate
	if errorRate >= upstreamHealthErrorRateThreshold {
		reasons = append(reasons, fmt.Sprintf("high_error_rate_%.2f", errorRate))
	}

	// Determine overall status.
	switch {
	case !site.Enabled || item.Suppressed || site.LastRefreshStatus == "failed" || refreshStale && noModels || errorRate >= upstreamHealthErrorRateThreshold:
		item.Status = model.UpstreamHealthStatusUnhealthy
	case refreshStale || item.BalanceAlert || errorRate >= 0.2 || noModels:
		item.Status = model.UpstreamHealthStatusDegraded
	default:
		item.Status = model.UpstreamHealthStatusHealthy
	}

	if item.Suppressed {
		reasons = append(reasons, "suppressed_"+upstreamSiteSuppressedReason(site.ID))
	}
	item.Reasons = reasons
	return item
}

func upstreamSiteErrorRate(ctx context.Context, siteID int) float64 {
	if siteID <= 0 {
		return 0
	}
	// Find channels linked to this upstream site.
	var channelIDs []int
	if err := db.GetDB().WithContext(ctx).Model(&model.Channel{}).
		Where("upstream_site_id = ?", siteID).
		Pluck("id", &channelIDs).Error; err != nil || len(channelIDs) == 0 {
		return 0
	}

	end := int(time.Now().Unix())
	start := int(time.Now().Add(-upstreamHealthErrorRateWindow).Unix())
	logs, err := RelayLogExport(ctx, &start, &end, 2000)
	if err != nil {
		return 0
	}

	var success, failure int64
	for _, log := range logs {
		if log.ChannelId == 0 {
			continue
		}
		found := false
		for _, id := range channelIDs {
			if log.ChannelId == id {
				found = true
				break
			}
		}
		if !found {
			continue
		}
		if log.Error != "" {
			failure++
		} else {
			success++
		}
	}
	if success+failure == 0 {
		return 0
	}
	return float64(failure) / float64(success+failure)
}

// UpstreamSiteUsage aggregates usage data for the channels linked to an upstream site.
func UpstreamSiteUsage(ctx context.Context, siteID int, days int) (model.UpstreamUsageResponse, error) {
	if siteID <= 0 {
		return model.UpstreamUsageResponse{}, fmt.Errorf("invalid upstream site id")
	}
	if days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}

	var site model.UpstreamSite
	if err := db.GetDB().WithContext(ctx).First(&site, siteID).Error; err != nil {
		return model.UpstreamUsageResponse{}, fmt.Errorf("upstream site not found")
	}

	var channelIDs []int
	if err := db.GetDB().WithContext(ctx).Model(&model.Channel{}).
		Where("upstream_site_id = ?", siteID).
		Pluck("id", &channelIDs).Error; err != nil {
		return model.UpstreamUsageResponse{}, err
	}

	now := time.Now()
	start := now.Add(-time.Duration(days) * 24 * time.Hour)
	startSec := int(start.Unix())
	endSec := int(now.Unix())

	logs, err := RelayLogExport(ctx, &startSec, &endSec, 10000)
	if err != nil {
		return model.UpstreamUsageResponse{}, err
	}

	pointMap := make(map[string]*model.UpstreamUsagePoint, days)
	for i := 0; i < days; i++ {
		d := now.Add(-time.Duration(i) * 24 * time.Hour)
		dateStr := d.Format("2006-01-02")
		pointMap[dateStr] = &model.UpstreamUsagePoint{
			Date:      dateStr,
			Timestamp: time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.Local).Unix(),
		}
	}

	for _, log := range logs {
		if log.ChannelId == 0 {
			continue
		}
		found := false
		for _, id := range channelIDs {
			if log.ChannelId == id {
				found = true
				break
			}
		}
		if !found {
			continue
		}
		dateStr := time.Unix(log.Time, 0).Format("2006-01-02")
		point, ok := pointMap[dateStr]
		if !ok {
			continue
		}
		point.RequestCount++
		point.InputTokens += int64(log.InputTokens)
		point.OutputTokens += int64(log.OutputTokens)
		point.TotalTokens += int64(log.InputTokens + log.OutputTokens)
		point.Cost += log.Cost
		if log.Error != "" {
			point.FailureCount++
		} else {
			point.SuccessCount++
		}
	}

	points := make([]model.UpstreamUsagePoint, 0, days)
	for i := days - 1; i >= 0; i-- {
		d := now.Add(-time.Duration(i) * 24 * time.Hour)
		dateStr := d.Format("2006-01-02")
		if point, ok := pointMap[dateStr]; ok {
			points = append(points, *point)
		}
	}

	return model.UpstreamUsageResponse{
		SiteID:     siteID,
		Days:       days,
		Points:     points,
		ChannelIDs: channelIDs,
	}, nil
}

// UpstreamSiteRestorePriority is the backend action for the manual restore endpoint.
func UpstreamSiteRestorePriority(ctx context.Context, siteID int) error {
	if siteID <= 0 {
		return fmt.Errorf("invalid upstream site id")
	}
	var site model.UpstreamSite
	if err := db.GetDB().WithContext(ctx).First(&site, siteID).Error; err != nil {
		return fmt.Errorf("upstream site not found")
	}
	UpstreamSiteRestore(siteID)
	return nil
}
