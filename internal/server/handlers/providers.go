package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/assets"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/middleware"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/server/router"
	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/utils/log"
)

const (
	providersGitHubURL        = "https://raw.githubusercontent.com/xiaoli0412/octopus-xiaoli-repo/bfa27aecbec14329a43550c0d5409aefea257af0/internal/assets/providers.json"
	providersPinnedCommitSHA  = "bfa27aecbec14329a43550c0d5409aefea257af0"
	providersCacheTTL         = 10 * time.Minute
	providersRemoteTimeout    = 3 * time.Second
	maxProvidersPayloadBytes  = 2 << 20
	providersResponseMimeType = "application/json"
	providersSourceHeader     = "X-Octopus-Providers-Source"
	providersCommitHeader     = "X-Octopus-Providers-Commit"
)

type providersCacheState struct {
	data      []byte
	expiresAt time.Time
	source    string
}

type providerPayload struct {
	Name        string `json:"name"`
	ChannelType int    `json:"channel_type"`
	BaseURL     string `json:"base_url"`
	AuthType    string `json:"auth_type,omitempty"`
}

var (
	providersNow          = time.Now
	providersHTTPClient   = newProvidersHTTPClient()
	providersRemoteFetch  = fetchProvidersFromGitHub
	providersEmbeddedRead = readEmbeddedProviders

	providersCacheMu sync.RWMutex
	providersCache   providersCacheState
)

func newProvidersHTTPClient() *http.Client {
	return &http.Client{
		Timeout: providersRemoteTimeout,
		// Keep the trust boundary pinned to the configured GitHub raw URL.
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func getProvidersPayload() ([]byte, string, error) {
	now := providersNow()

	providersCacheMu.RLock()
	if len(providersCache.data) > 0 && now.Before(providersCache.expiresAt) {
		data := append([]byte(nil), providersCache.data...)
		source := providersCache.source
		providersCacheMu.RUnlock()
		return data, source, nil
	}
	providersCacheMu.RUnlock()

	providersCacheMu.Lock()
	defer providersCacheMu.Unlock()

	now = providersNow()
	if len(providersCache.data) > 0 && now.Before(providersCache.expiresAt) {
		return append([]byte(nil), providersCache.data...), providersCache.source, nil
	}

	data, err := providersRemoteFetch()
	source := "remote"
	if err != nil {
		log.Warnf("Failed to fetch providers from GitHub: %s, error: %v", providersGitHubURL, err)
		data, err = providersEmbeddedRead()
		if err != nil {
			return nil, "", err
		}
		source = "embedded"
		log.Infof("Loading providers from embedded providers.json")
	} else {
		if err := validateProvidersPayload(data); err != nil {
			log.Warnf("Failed to validate providers from GitHub: %s, error: %v", providersGitHubURL, err)
			data, err = providersEmbeddedRead()
			if err != nil {
				return nil, "", err
			}
			source = "embedded"
			log.Infof("Loading providers from embedded providers.json")
		} else {
			log.Infof("Successfully fetched providers from GitHub")
		}
	}
	if err := validateProvidersPayload(data); err != nil {
		return nil, "", err
	}

	providersCache = providersCacheState{
		data:      append([]byte(nil), data...),
		expiresAt: now.Add(providersCacheTTL),
		source:    source,
	}

	return append([]byte(nil), data...), source, nil
}

func fetchProvidersFromGitHub() ([]byte, error) {
	log.Infof("Fetching providers from GitHub: %s", providersGitHubURL)
	resp, err := providersHTTPClient.Get(providersGitHubURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxProvidersPayloadBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxProvidersPayloadBytes {
		return nil, fmt.Errorf("payload too large")
	}
	if err := validateProvidersPayload(data); err != nil {
		return nil, err
	}

	return data, nil
}

func readEmbeddedProviders() ([]byte, error) {
	file, err := assets.ProvidersFS.Open("providers.json")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxProvidersPayloadBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxProvidersPayloadBytes {
		return nil, fmt.Errorf("embedded providers payload too large")
	}
	if err := validateProvidersPayload(data); err != nil {
		return nil, fmt.Errorf("embedded providers payload invalid: %w", err)
	}

	return data, nil
}

func validateProvidersPayload(data []byte) error {
	var providers []providerPayload
	if err := json.Unmarshal(data, &providers); err != nil {
		return fmt.Errorf("invalid providers payload: %w", err)
	}
	for i, provider := range providers {
		if provider.Name == "" {
			return fmt.Errorf("invalid providers payload: provider %d missing name", i)
		}
		if provider.BaseURL == "" {
			return fmt.Errorf("invalid providers payload: provider %d missing base_url", i)
		}
		if err := validateAbsoluteHTTPURL(provider.BaseURL, "base_url"); err != nil {
			return fmt.Errorf("invalid providers payload: provider %d base_url must be absolute http or https URL", i)
		}
	}
	return nil
}

func resetProvidersState() {
	providersCacheMu.Lock()
	defer providersCacheMu.Unlock()

	providersNow = time.Now
	providersHTTPClient = newProvidersHTTPClient()
	providersRemoteFetch = fetchProvidersFromGitHub
	providersEmbeddedRead = readEmbeddedProviders
	providersCache = providersCacheState{}
}

// GetProviders returns the list of providers.
// Priority: GitHub raw URL -> embedded providers.json.
func GetProviders(c *gin.Context) {
	data, source, err := getProvidersPayload()
	if err != nil {
		log.Errorf("Failed to load providers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load providers"})
		return
	}

	c.Header(providersSourceHeader, source)
	c.Header(providersCommitHeader, providersPinnedCommitSHA)
	c.Data(http.StatusOK, providersResponseMimeType, data)
}

func init() {
	router.NewGroupRouter("/api/v1/providers").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("", http.MethodGet).Handle(GetProviders),
		)
}
