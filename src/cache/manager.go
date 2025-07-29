package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/matthewmyrick/azure-searcher/src/config"
	"github.com/matthewmyrick/azure-searcher/src/types"
)

// Manager handles cache operations
type Manager struct {
	cacheFile string
	data      *types.CacheData
	ttl       int // TTL in minutes
}

// NewManager creates a new cache manager
func NewManager(cacheFile string) (*Manager, error) {
	data, err := loadCacheFromFile(cacheFile)
	if err != nil {
		data = &types.CacheData{
			Subscriptions: make(map[string]types.SubscriptionCache),
			Version:       "1.0",
		}
	}

	return &Manager{
		cacheFile: cacheFile,
		data:      data,
		ttl:       config.DefaultCacheTTLMinutes,
	}, nil
}

// GetCachedResourceGroups retrieves cached resource groups for a subscription
func (m *Manager) GetCachedResourceGroups(subscriptionID string) ([]types.ResourceGroup, bool) {
	subCache, exists := m.data.Subscriptions[subscriptionID]
	if !exists {
		return nil, false
	}

	// Determine the TTL to use: subscription-specific interval or global default
	ttlMinutes := m.ttl
	if m.GetSubscriptionAutoRefreshEnabled(subscriptionID) {
		intervalStr := m.GetSubscriptionRefreshInterval(subscriptionID)
		if intervalStr != "" {
			if interval, err := m.ParseRefreshInterval(intervalStr); err == nil {
				ttlMinutes = int(interval.Minutes())
			}
		}
	}

	if !isCacheValid(subCache.LastUpdated, ttlMinutes) {
		return nil, false
	}

	var resourceGroups []types.ResourceGroup
	for _, cachedRG := range subCache.ResourceGroups {
		if isCacheValid(cachedRG.CachedAt, ttlMinutes) {
			resourceGroups = append(resourceGroups, types.ResourceGroup{
				Name:      cachedRG.Name,
				Location:  cachedRG.Location,
				Tags:      cachedRG.Tags,
				Resources: cachedRG.Resources,
				Expanded:  false,
			})
		}
	}

	if len(resourceGroups) == 0 {
		return nil, false
	}

	sort.Slice(resourceGroups, func(i, j int) bool {
		return resourceGroups[i].Name < resourceGroups[j].Name
	})

	return resourceGroups, true
}

// CacheResourceGroups stores resource groups in the cache
func (m *Manager) CacheResourceGroups(subscriptionID, subscriptionName string, resourceGroups []types.ResourceGroup) error {
	if m.data.Subscriptions == nil {
		m.data.Subscriptions = make(map[string]types.SubscriptionCache)
	}

	cachedRGs := make(map[string]types.CachedResourceGroup)
	now := time.Now()

	for _, rg := range resourceGroups {
		cachedRGs[rg.Name] = types.CachedResourceGroup{
			Name:      rg.Name,
			Location:  rg.Location,
			Tags:      rg.Tags,
			Resources: rg.Resources,
			CachedAt:  now,
		}
	}

	m.data.Subscriptions[subscriptionID] = types.SubscriptionCache{
		SubscriptionID:   subscriptionID,
		SubscriptionName: subscriptionName,
		ResourceGroups:   cachedRGs,
		LastUpdated:      now,
	}

	return m.Save()
}

// InvalidateSubscription removes cached data for a specific subscription
func (m *Manager) InvalidateSubscription(subscriptionID string) {
	delete(m.data.Subscriptions, subscriptionID)
}

// Save writes the cache data to disk
func (m *Manager) Save() error {
	return saveCacheToFile(m.cacheFile, m.data)
}

// GetFilePath returns the cache file path
func GetFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config.CacheFilename
	}
	return filepath.Join(homeDir, ".azure-searcher", config.CacheFilename)
}

// GetConfigFilePath returns the config file path
func GetConfigFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "azure-searcher-config.json"
	}
	return filepath.Join(homeDir, ".azure-searcher", "config.json")
}

// SubscriptionConfig represents cache configuration for a subscription
type SubscriptionConfig struct {
	CacheEnabled      bool   `json:"cache_enabled"`
	AutoRefreshEnabled bool   `json:"auto_refresh_enabled"`
	RefreshInterval   string `json:"refresh_interval"` // e.g., "2 hr", "30 min", "1 dy"
}

// Config represents the application configuration
type Config struct {
	Subscriptions map[string]SubscriptionConfig `json:"subscriptions"`
	Version       string                        `json:"version"`
}

// GetSubscriptionCacheEnabled returns whether caching is enabled for a subscription
func (m *Manager) GetSubscriptionCacheEnabled(subscriptionID string) bool {
	config, err := loadConfigFromFile(GetConfigFilePath())
	if err != nil {
		return false // Default to disabled
	}
	
	subConfig, exists := config.Subscriptions[subscriptionID]
	if !exists {
		return false // Default to disabled
	}
	
	return subConfig.CacheEnabled
}

// SetSubscriptionCacheEnabled sets whether caching is enabled for a subscription
func (m *Manager) SetSubscriptionCacheEnabled(subscriptionID string, enabled bool) error {
	configFile := GetConfigFilePath()
	config, err := loadConfigFromFile(configFile)
	if err != nil {
		config = &Config{
			Subscriptions: make(map[string]SubscriptionConfig),
			Version:       "1.0",
		}
	}
	
	// Get existing config or create new one
	subConfig := config.Subscriptions[subscriptionID]
	subConfig.CacheEnabled = enabled
	config.Subscriptions[subscriptionID] = subConfig
	
	return saveConfigToFile(configFile, config)
}

// GetSubscriptionAutoRefreshEnabled returns whether auto-refresh is enabled for a subscription
func (m *Manager) GetSubscriptionAutoRefreshEnabled(subscriptionID string) bool {
	config, err := loadConfigFromFile(GetConfigFilePath())
	if err != nil {
		return false
	}
	
	subConfig, exists := config.Subscriptions[subscriptionID]
	if !exists {
		return false
	}
	
	return subConfig.AutoRefreshEnabled
}

// GetSubscriptionRefreshInterval returns the refresh interval for a subscription
func (m *Manager) GetSubscriptionRefreshInterval(subscriptionID string) string {
	config, err := loadConfigFromFile(GetConfigFilePath())
	if err != nil {
		return ""
	}
	
	subConfig, exists := config.Subscriptions[subscriptionID]
	if !exists {
		return ""
	}
	
	return subConfig.RefreshInterval
}

// SetSubscriptionAutoRefresh sets the auto-refresh settings for a subscription
func (m *Manager) SetSubscriptionAutoRefresh(subscriptionID string, enabled bool, interval string) error {
	configFile := GetConfigFilePath()
	config, err := loadConfigFromFile(configFile)
	if err != nil {
		config = &Config{
			Subscriptions: make(map[string]SubscriptionConfig),
			Version:       "1.0",
		}
	}
	
	// Get existing config or create new one
	subConfig := config.Subscriptions[subscriptionID]
	subConfig.AutoRefreshEnabled = enabled
	subConfig.RefreshInterval = interval
	config.Subscriptions[subscriptionID] = subConfig
	
	return saveConfigToFile(configFile, config)
}

// ValidateRefreshInterval validates and parses a refresh interval string
// Expected format: "<number> <unit>" where unit is "min", "hr", or "dy"
// Examples: "30 min", "2 hr", "1 dy"
func (m *Manager) ValidateRefreshInterval(interval string) error {
	parts := strings.Fields(interval)
	if len(parts) != 2 {
		return fmt.Errorf("interval must be in format '<number> <unit>' (e.g., '2 hr', '30 min', '1 dy')")
	}
	
	// Parse number
	num, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("first part must be a valid integer")
	}
	
	if num <= 0 {
		return fmt.Errorf("interval must be greater than 0")
	}
	
	// Validate unit
	unit := parts[1]
	if unit != "min" && unit != "hr" && unit != "dy" {
		return fmt.Errorf("unit must be 'min', 'hr', or 'dy'")
	}
	
	return nil
}

// ParseRefreshInterval converts a refresh interval string to a time.Duration
func (m *Manager) ParseRefreshInterval(interval string) (time.Duration, error) {
	if err := m.ValidateRefreshInterval(interval); err != nil {
		return 0, err
	}
	
	parts := strings.Fields(interval)
	num, _ := strconv.Atoi(parts[0]) // Already validated above
	unit := parts[1]
	
	switch unit {
	case "min":
		return time.Duration(num) * time.Minute, nil
	case "hr":
		return time.Duration(num) * time.Hour, nil
	case "dy":
		return time.Duration(num) * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid unit: %s", unit)
	}
}

// ShouldRefreshCache determines if cache should be refreshed based on auto-refresh settings
func (m *Manager) ShouldRefreshCache(subscriptionID string) bool {
	// Check if cache is enabled
	if !m.GetSubscriptionCacheEnabled(subscriptionID) {
		return false
	}
	
	// Check if auto-refresh is enabled
	if !m.GetSubscriptionAutoRefreshEnabled(subscriptionID) {
		return false
	}
	
	// Get refresh interval
	intervalStr := m.GetSubscriptionRefreshInterval(subscriptionID)
	if intervalStr == "" {
		return false
	}
	
	// Parse interval
	interval, err := m.ParseRefreshInterval(intervalStr)
	if err != nil {
		return false
	}
	
	// Check last cache update time
	subCache, exists := m.data.Subscriptions[subscriptionID]
	if !exists {
		return true // No cache exists, should refresh
	}
	
	// Check if enough time has passed
	timeSinceUpdate := time.Since(subCache.LastUpdated)
	return timeSinceUpdate >= interval
}

// loadConfigFromFile loads configuration from file
func loadConfigFromFile(configFile string) (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Subscriptions: make(map[string]SubscriptionConfig),
				Version:       "1.0",
			}, nil
		}
		return nil, err
	}
	
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return &Config{
			Subscriptions: make(map[string]SubscriptionConfig),
			Version:       "1.0",
		}, nil
	}
	
	return &config, nil
}

// saveConfigToFile saves configuration to file
func saveConfigToFile(configFile string, config *Config) error {
	dir := filepath.Dir(configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(configFile, data, 0644)
}

// loadCacheFromFile loads cache data from the specified file
func loadCacheFromFile(cacheFile string) (*types.CacheData, error) {
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &types.CacheData{
				Subscriptions: make(map[string]types.SubscriptionCache),
				Version:       "1.0",
			}, nil
		}
		return nil, err
	}

	var cache types.CacheData
	err = json.Unmarshal(data, &cache)
	if err != nil {
		return &types.CacheData{
			Subscriptions: make(map[string]types.SubscriptionCache),
			Version:       "1.0",
		}, nil
	}

	return &cache, nil
}

// saveCacheToFile saves cache data to the specified file
func saveCacheToFile(cacheFile string, cache *types.CacheData) error {
	dir := filepath.Dir(cacheFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, data, 0644)
}

// isCacheValid checks if cached data is still valid based on TTL
func isCacheValid(cachedAt time.Time, ttlMinutes int) bool {
	return time.Since(cachedAt) < time.Duration(ttlMinutes)*time.Minute
}