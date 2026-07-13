package op

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/model"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
)

// CostAlertFormat constants for supported webhook message formats.
const (
	CostAlertFormatGeneric  = "generic"
	CostAlertFormatSlack    = "slack"
	CostAlertFormatFeishu   = "feishu"
	CostAlertFormatDingTalk = "dingtalk"
)

// CostAlertConfig holds configuration for cost alert webhooks loaded from settings.
type CostAlertConfig struct {
	WebhookURL string
	Thresholds []float64
	Format     string
}

// AlertPayload describes a cost alert event sent to the webhook.
type AlertPayload struct {
	APIKeyName  string    `json:"api_key_name"`
	APIKeyID    uint      `json:"api_key_id"`
	CurrentCost float64   `json:"current_cost"`
	MaxCost     float64   `json:"max_cost"`
	Threshold   float64   `json:"threshold"`
	Percentage  float64   `json:"percentage"`
	Timestamp   time.Time `json:"timestamp"`
}

// costAlertTriggered records which (apiKeyID, threshold) pairs have already fired.
// Values are stored in-memory; a process restart resets all triggered records.
var costAlertTriggered sync.Map

// loadCostAlertConfig reads cost alert settings from the setting cache.
// Returns a config with empty WebhookURL if alerts are disabled.
func loadCostAlertConfig() CostAlertConfig {
	cfg := CostAlertConfig{
		Format: CostAlertFormatGeneric,
	}
	if v, err := SettingGetString(model.SettingKeyCostAlertWebhookURL); err == nil {
		cfg.WebhookURL = strings.TrimSpace(v)
	}
	if v, err := SettingGetString(model.SettingKeyCostAlertThresholds); err == nil {
		cfg.Thresholds = parseCostAlertThresholds(v)
	}
	if len(cfg.Thresholds) == 0 {
		cfg.Thresholds = []float64{0.5, 0.8, 1.0}
	}
	if v, err := SettingGetString(model.SettingKeyCostAlertFormat); err == nil {
		normalized := strings.ToLower(strings.TrimSpace(v))
		if normalized != "" {
			cfg.Format = normalized
		}
	}
	return cfg
}

// parseCostAlertThresholds parses a comma-separated threshold string (e.g. "0.5,0.8,1.0").
// Invalid entries are silently skipped.
func parseCostAlertThresholds(raw string) []float64 {
	parts := strings.Split(raw, ",")
	out := make([]float64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.ParseFloat(p, 64)
		if err != nil {
			continue
		}
		if v <= 0 || v > 2 {
			continue
		}
		out = append(out, v)
	}
	return out
}

// costAlertKey builds the dedup key for a given API key ID and threshold.
func costAlertKey(apiKeyID uint, threshold float64) string {
	return fmt.Sprintf("%d-%g", apiKeyID, threshold)
}

// CheckCostAlert checks whether the API key cost has reached any configured threshold
// and triggers a webhook alert if so. Each threshold fires at most once per API key
// until ResetCostAlerts is called.
//
// This function is safe to call from a goroutine. It returns silently (no error)
// when maxCost <= 0 or when no webhook URL is configured.
func CheckCostAlert(apiKeyID uint, apiKeyName string, currentCost float64, maxCost float64) {
	if maxCost <= 0 {
		return
	}
	cfg := loadCostAlertConfig()
	if cfg.WebhookURL == "" {
		return
	}
	percentage := currentCost / maxCost
	for _, threshold := range cfg.Thresholds {
		if percentage < threshold {
			continue
		}
		key := costAlertKey(apiKeyID, threshold)
		if _, loaded := costAlertTriggered.LoadOrStore(key, struct{}{}); loaded {
			continue
		}
		payload := AlertPayload{
			APIKeyName:  apiKeyName,
			APIKeyID:    apiKeyID,
			CurrentCost: currentCost,
			MaxCost:     maxCost,
			Threshold:   threshold,
			Percentage:  percentage,
			Timestamp:   time.Now(),
		}
		if err := sendWebhookAlert(cfg, payload); err != nil {
			log.Warnf("cost alert webhook failed for api key %d threshold %g: %v", apiKeyID, threshold, err)
		}
	}
}

// ResetCostAlerts clears all triggered threshold records for the given API key.
// Call this when the API key cost quota is reset so thresholds can fire again.
func ResetCostAlerts(apiKeyID uint) {
	prefix := fmt.Sprintf("%d-", apiKeyID)
	costAlertTriggered.Range(func(k, _ any) bool {
		if key, ok := k.(string); ok && strings.HasPrefix(key, prefix) {
			costAlertTriggered.Delete(key)
		}
		return true
	})
}

// ResetAllCostAlerts clears all triggered threshold records for all API keys.
// Primarily used in tests.
func ResetAllCostAlerts() {
	costAlertTriggered.Range(func(k, _ any) bool {
		costAlertTriggered.Delete(k)
		return true
	})
}

// sendWebhookAlert sends a cost alert to the configured webhook URL with a 5s timeout.
func sendWebhookAlert(config CostAlertConfig, payload AlertPayload) error {
	body, err := buildWebhookBody(config.Format, payload)
	if err != nil {
		return fmt.Errorf("build webhook body: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, config.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send webhook request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success status: %d", resp.StatusCode)
	}
	return nil
}

// buildWebhookBody constructs the webhook request body based on the message format.
func buildWebhookBody(format string, payload AlertPayload) ([]byte, error) {
	message := fmt.Sprintf("[Octopus Cost Alert] API Key %q (id=%d) cost reached %.2f%% of max cost (current=%.6f, max=%.6f, threshold=%.2f%%) at %s",
		payload.APIKeyName, payload.APIKeyID, payload.Percentage*100,
		payload.CurrentCost, payload.MaxCost, payload.Threshold*100,
		payload.Timestamp.Format(time.RFC3339))
	switch strings.ToLower(strings.TrimSpace(format)) {
	case CostAlertFormatSlack:
		return json.Marshal(map[string]string{"text": message})
	case CostAlertFormatFeishu:
		return json.Marshal(map[string]any{
			"msg_type": "text",
			"content":  map[string]string{"text": message},
		})
	case CostAlertFormatDingTalk:
		return json.Marshal(map[string]any{
			"msgtype": "text",
			"text":    map[string]string{"content": message},
		})
	default: // CostAlertFormatGeneric
		return json.Marshal(payload)
	}
}
