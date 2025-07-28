package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"azure-searcher/src/config"
	"azure-searcher/src/types"
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

	if !isCacheValid(subCache.LastUpdated, m.ttl) {
		return nil, false
	}

	var resourceGroups []types.ResourceGroup
	for _, cachedRG := range subCache.ResourceGroups {
		if isCacheValid(cachedRG.CachedAt, m.ttl) {
			resourceGroups = append(resourceGroups, types.ResourceGroup{
				Name:      cachedRG.Name,
				Location:  cachedRG.Location,
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