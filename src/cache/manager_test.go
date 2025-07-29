package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/matthewmyrick/azure-searcher/src/config"
	"github.com/matthewmyrick/azure-searcher/src/testutil"
	"github.com/matthewmyrick/azure-searcher/src/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	tempDir := testutil.TempDir(t)
	cacheFile := filepath.Join(tempDir, "test-cache.json")

	// Test creating manager with non-existent cache file
	manager, err := NewManager(cacheFile)
	require.NoError(t, err)
	assert.NotNil(t, manager)
	assert.Equal(t, cacheFile, manager.cacheFile)
	assert.Equal(t, config.DefaultCacheTTLMinutes, manager.ttl)
	assert.NotNil(t, manager.data)
	assert.Equal(t, "1.0", manager.data.Version)
	assert.NotNil(t, manager.data.Subscriptions)
}

func TestNewManager_WithExistingCache(t *testing.T) {
	tempDir := testutil.TempDir(t)
	cacheFile := filepath.Join(tempDir, "existing-cache.json")

	// Create existing cache file
	existingCache := testutil.MockCacheData(
		map[string]types.SubscriptionCache{
			"sub-1": testutil.MockSubscriptionCache("sub-1", "Test Sub", nil, testutil.FixedTime()),
		},
		"1.0",
	)
	data, err := json.MarshalIndent(existingCache, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(cacheFile, data, 0644)
	require.NoError(t, err)

	// Test creating manager with existing cache file
	manager, err := NewManager(cacheFile)
	require.NoError(t, err)
	
	assert.Equal(t, 1, len(manager.data.Subscriptions))
	assert.Contains(t, manager.data.Subscriptions, "sub-1")
}

func TestManager_GetCachedResourceGroups_NoCache(t *testing.T) {
	tempDir := testutil.TempDir(t)
	cacheFile := filepath.Join(tempDir, "test-cache.json")
	
	manager, err := NewManager(cacheFile)
	require.NoError(t, err)

	// Test getting cached data for non-existent subscription
	rgs, valid := manager.GetCachedResourceGroups("non-existent")
	assert.False(t, valid)
	assert.Nil(t, rgs)
}

func TestManager_GetCachedResourceGroups_ValidCache(t *testing.T) {
	tempDir := testutil.TempDir(t)
	cacheFile := filepath.Join(tempDir, "test-cache.json")
	
	manager, err := NewManager(cacheFile)
	require.NoError(t, err)

	// Create test data
	now := time.Now()
	resources := []types.Resource{
		testutil.MockResource("vm1", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm1", nil),
		testutil.MockResource("storage1", "Microsoft.Storage/storageAccounts", "storage", "/resource/storage1", nil),
	}
	
	cachedRG := testutil.MockCachedResourceGroup("test-rg", "eastus", resources, now, nil)
	subCache := testutil.MockSubscriptionCache("sub-1", "Test Sub", 
		map[string]types.CachedResourceGroup{"test-rg": cachedRG}, now)
	
	manager.data.Subscriptions["sub-1"] = subCache

	// Test getting valid cached data
	rgs, valid := manager.GetCachedResourceGroups("sub-1")
	assert.True(t, valid)
	assert.Len(t, rgs, 1)
	assert.Equal(t, "test-rg", rgs[0].Name)
	assert.Equal(t, "eastus", rgs[0].Location)
	assert.Len(t, rgs[0].Resources, 2)
	assert.False(t, rgs[0].Expanded) // Should be reset to false
}

func TestManager_GetCachedResourceGroups_ExpiredCache(t *testing.T) {
	tempDir := testutil.TempDir(t)
	cacheFile := filepath.Join(tempDir, "test-cache.json")
	
	manager, err := NewManager(cacheFile)
	require.NoError(t, err)

	// Create expired cache data
	expiredTime := time.Now().Add(-time.Hour) // 1 hour ago, beyond default TTL
	cachedRG := testutil.MockCachedResourceGroup("test-rg", "eastus", nil, expiredTime, nil)
	subCache := testutil.MockSubscriptionCache("sub-1", "Test Sub", 
		map[string]types.CachedResourceGroup{"test-rg": cachedRG}, expiredTime)
	
	manager.data.Subscriptions["sub-1"] = subCache

	// Test getting expired cached data
	rgs, valid := manager.GetCachedResourceGroups("sub-1")
	assert.False(t, valid)
	assert.Nil(t, rgs)
}

func TestManager_CacheResourceGroups(t *testing.T) {
	tempDir := testutil.TempDir(t)
	cacheFile := filepath.Join(tempDir, "test-cache.json")
	
	manager, err := NewManager(cacheFile)
	require.NoError(t, err)

	// Create test data
	resources := []types.Resource{
		testutil.MockResource("vm1", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm1", nil),
	}
	resourceGroups := []types.ResourceGroup{
		testutil.MockResourceGroup("test-rg", "eastus", resources, nil),
	}

	// Cache the resource groups
	err = manager.CacheResourceGroups("sub-1", "Test Subscription", resourceGroups)
	require.NoError(t, err)

	// Verify data was cached
	assert.Contains(t, manager.data.Subscriptions, "sub-1")
	subCache := manager.data.Subscriptions["sub-1"]
	assert.Equal(t, "sub-1", subCache.SubscriptionID)
	assert.Equal(t, "Test Subscription", subCache.SubscriptionName)
	assert.Contains(t, subCache.ResourceGroups, "test-rg")

	cachedRG := subCache.ResourceGroups["test-rg"]
	assert.Equal(t, "test-rg", cachedRG.Name)
	assert.Equal(t, "eastus", cachedRG.Location)
	assert.Len(t, cachedRG.Resources, 1)
	testutil.AssertTimeAlmostEqual(t, time.Now(), cachedRG.CachedAt)
}

func TestManager_InvalidateSubscription(t *testing.T) {
	tempDir := testutil.TempDir(t)
	cacheFile := filepath.Join(tempDir, "test-cache.json")
	
	manager, err := NewManager(cacheFile)
	require.NoError(t, err)

	// Add some cached data
	subCache := testutil.MockSubscriptionCache("sub-1", "Test Sub", nil, time.Now())
	manager.data.Subscriptions["sub-1"] = subCache
	manager.data.Subscriptions["sub-2"] = testutil.MockSubscriptionCache("sub-2", "Test Sub 2", nil, time.Now())

	// Invalidate one subscription
	manager.InvalidateSubscription("sub-1")

	// Verify only sub-1 was removed
	assert.NotContains(t, manager.data.Subscriptions, "sub-1")
	assert.Contains(t, manager.data.Subscriptions, "sub-2")
}

func TestGetFilePath(t *testing.T) {
	path := GetFilePath()
	
	assert.Contains(t, path, ".azure-searcher")
	assert.Contains(t, path, config.CacheFilename)
	assert.True(t, filepath.IsAbs(path) || strings.Contains(path, config.CacheFilename))
}

func TestGetConfigFilePath(t *testing.T) {
	path := GetConfigFilePath()
	
	assert.Contains(t, path, ".azure-searcher")
	assert.Contains(t, path, "config.json")
	assert.True(t, filepath.IsAbs(path) || strings.Contains(path, "config.json"))
}

func TestManager_ValidateRefreshInterval(t *testing.T) {
	tempDir := testutil.TempDir(t)
	cacheFile := filepath.Join(tempDir, "test-cache.json")
	
	manager, err := NewManager(cacheFile)
	require.NoError(t, err)

	tests := []struct {
		interval    string
		shouldError bool
		errorMsg    string
	}{
		{"30 min", false, ""},
		{"2 hr", false, ""},
		{"1 dy", false, ""},
		{"10 min", false, ""},
		{"24 hr", false, ""},
		{"invalid", true, "interval must be in format"},
		{"30", true, "interval must be in format"},
		{"30 minutes", true, "unit must be"},
		{"-5 min", true, "interval must be greater than 0"},
		{"0 min", true, "interval must be greater than 0"},
		{"abc min", true, "first part must be a valid integer"},
		{"30 seconds", true, "unit must be"},
	}

	for _, tt := range tests {
		t.Run(tt.interval, func(t *testing.T) {
			err := manager.ValidateRefreshInterval(tt.interval)
			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_ParseRefreshInterval(t *testing.T) {
	tempDir := testutil.TempDir(t)
	cacheFile := filepath.Join(tempDir, "test-cache.json")
	
	manager, err := NewManager(cacheFile)
	require.NoError(t, err)

	tests := []struct {
		interval string
		expected time.Duration
	}{
		{"30 min", 30 * time.Minute},
		{"2 hr", 2 * time.Hour},
		{"1 dy", 24 * time.Hour},
		{"90 min", 90 * time.Minute},
		{"12 hr", 12 * time.Hour},
		{"7 dy", 7 * 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.interval, func(t *testing.T) {
			duration, err := manager.ParseRefreshInterval(tt.interval)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, duration)
		})
	}

	// Test invalid intervals
	_, err = manager.ParseRefreshInterval("invalid")
	assert.Error(t, err)
}

func TestManager_SubscriptionConfig_Operations(t *testing.T) {
	tempDir := testutil.TempDir(t)
	configDir := filepath.Join(tempDir, ".azure-searcher")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	// Mock the GetConfigFilePath to use our temp directory
	_ = GetConfigFilePath() // Original config path for reference

	// We can't easily mock the function, so we'll work with actual config operations
	cacheFile := filepath.Join(configDir, "cache.json")
	manager, err := NewManager(cacheFile)
	require.NoError(t, err)

	subID := "test-sub-123"

	// Initially, cache should be disabled
	enabled := manager.GetSubscriptionCacheEnabled(subID)
	assert.False(t, enabled)

	// Enable cache for subscription
	err = manager.SetSubscriptionCacheEnabled(subID, true)
	require.NoError(t, err)

	// Verify cache is now enabled
	enabled = manager.GetSubscriptionCacheEnabled(subID)
	assert.True(t, enabled)

	// Test auto-refresh settings
	autoRefreshEnabled := manager.GetSubscriptionAutoRefreshEnabled(subID)
	assert.False(t, autoRefreshEnabled)

	interval := manager.GetSubscriptionRefreshInterval(subID)
	assert.Empty(t, interval)

	// Set auto-refresh
	err = manager.SetSubscriptionAutoRefresh(subID, true, "2 hr")
	require.NoError(t, err)

	// Verify auto-refresh settings
	autoRefreshEnabled = manager.GetSubscriptionAutoRefreshEnabled(subID)
	assert.True(t, autoRefreshEnabled)

	interval = manager.GetSubscriptionRefreshInterval(subID)
	assert.Equal(t, "2 hr", interval)

	// Configuration operations completed successfully
}

func TestManager_ShouldRefreshCache(t *testing.T) {
	tempDir := testutil.TempDir(t)
	cacheFile := filepath.Join(tempDir, "test-cache.json")
	
	manager, err := NewManager(cacheFile)
	require.NoError(t, err)

	subID := "test-sub"

	// Should not refresh when cache is disabled
	shouldRefresh := manager.ShouldRefreshCache(subID)
	assert.False(t, shouldRefresh)

	// Enable cache but not auto-refresh
	err = manager.SetSubscriptionCacheEnabled(subID, true)
	require.NoError(t, err)

	shouldRefresh = manager.ShouldRefreshCache(subID)
	assert.False(t, shouldRefresh)

	// Enable auto-refresh with interval
	err = manager.SetSubscriptionAutoRefresh(subID, true, "1 hr")
	require.NoError(t, err)

	// Should refresh when no cache exists
	shouldRefresh = manager.ShouldRefreshCache(subID)
	assert.True(t, shouldRefresh)

	// Add recent cache data
	now := time.Now()
	subCache := testutil.MockSubscriptionCache(subID, "Test Sub", nil, now)
	manager.data.Subscriptions[subID] = subCache

	// Should not refresh when cache is recent
	shouldRefresh = manager.ShouldRefreshCache(subID)
	assert.False(t, shouldRefresh)

	// Add old cache data
	oldTime := now.Add(-2 * time.Hour) // Older than 1 hr interval
	subCache.LastUpdated = oldTime
	manager.data.Subscriptions[subID] = subCache

	// Should refresh when cache is old
	shouldRefresh = manager.ShouldRefreshCache(subID)
	assert.True(t, shouldRefresh)
}

func TestIsCacheValid(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		cachedAt   time.Time
		ttlMinutes int
		expected   bool
	}{
		{"recent cache", now.Add(-10 * time.Minute), 30, true},
		{"cache at edge", now.Add(-30 * time.Minute), 30, false},
		{"old cache", now.Add(-60 * time.Minute), 30, false},
		{"very recent", now.Add(-1 * time.Minute), 30, true},
		{"zero ttl", now.Add(-1 * time.Second), 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCacheValid(tt.cachedAt, tt.ttlMinutes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoadAndSaveCacheToFile(t *testing.T) {
	tempDir := testutil.TempDir(t)
	cacheFile := filepath.Join(tempDir, "test-cache.json")

	// Test loading non-existent file
	cache, err := loadCacheFromFile(cacheFile)
	require.NoError(t, err)
	assert.NotNil(t, cache)
	assert.Equal(t, "1.0", cache.Version)
	assert.NotNil(t, cache.Subscriptions)

	// Test saving and loading valid cache
	testCache := testutil.MockCacheData(
		map[string]types.SubscriptionCache{
			"sub-1": testutil.MockSubscriptionCache("sub-1", "Test Sub", nil, testutil.FixedTime()),
		},
		"1.0",
	)

	err = saveCacheToFile(cacheFile, &testCache)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(cacheFile)
	require.NoError(t, err)

	// Load and verify
	loadedCache, err := loadCacheFromFile(cacheFile)
	require.NoError(t, err)
	assert.Equal(t, testCache.Version, loadedCache.Version)
	assert.Len(t, loadedCache.Subscriptions, 1)
	assert.Contains(t, loadedCache.Subscriptions, "sub-1")

	// Test loading corrupted file
	corruptedFile := filepath.Join(tempDir, "corrupted.json")
	err = os.WriteFile(corruptedFile, []byte("invalid json"), 0644)
	require.NoError(t, err)

	corruptedCache, err := loadCacheFromFile(corruptedFile)
	require.NoError(t, err) // Should not error, should return default
	assert.Equal(t, "1.0", corruptedCache.Version)
	assert.NotNil(t, corruptedCache.Subscriptions)
}

func TestManager_GetCachedResourceGroups_WithSorting(t *testing.T) {
	tempDir := testutil.TempDir(t)
	cacheFile := filepath.Join(tempDir, "test-cache.json")
	
	manager, err := NewManager(cacheFile)
	require.NoError(t, err)

	// Create test data with multiple resource groups (unsorted)
	now := time.Now()
	cachedRGs := map[string]types.CachedResourceGroup{
		"zebra-rg": testutil.MockCachedResourceGroup("zebra-rg", "eastus", nil, now, nil),
		"alpha-rg": testutil.MockCachedResourceGroup("alpha-rg", "westus", nil, now, nil),
		"beta-rg":  testutil.MockCachedResourceGroup("beta-rg", "centralus", nil, now, nil),
	}
	
	subCache := testutil.MockSubscriptionCache("sub-1", "Test Sub", cachedRGs, now)
	manager.data.Subscriptions["sub-1"] = subCache

	// Get cached resource groups
	rgs, valid := manager.GetCachedResourceGroups("sub-1")
	assert.True(t, valid)
	assert.Len(t, rgs, 3)

	// Verify they are sorted alphabetically
	assert.Equal(t, "alpha-rg", rgs[0].Name)
	assert.Equal(t, "beta-rg", rgs[1].Name)
	assert.Equal(t, "zebra-rg", rgs[2].Name)
}

// BenchmarkManager_CacheResourceGroups tests the performance of caching operations
func BenchmarkManager_CacheResourceGroups(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "azure-searcher-bench-*")
	require.NoError(b, err)
	defer os.RemoveAll(tempDir)

	cacheFile := filepath.Join(tempDir, "bench-cache.json")
	manager, err := NewManager(cacheFile)
	require.NoError(b, err)

	// Create test data
	var resourceGroups []types.ResourceGroup
	for i := 0; i < 100; i++ {
		resources := make([]types.Resource, 10)
		for j := 0; j < 10; j++ {
			resources[j] = testutil.MockResource(
				"resource-"+string(rune(j)),
				"Microsoft.Test/resources",
				"test",
				"/resource/"+string(rune(j)),
				nil,
			)
		}
		resourceGroups = append(resourceGroups, testutil.MockResourceGroup(
			"rg-"+string(rune(i)),
			"eastus",
			resources,
			nil,
		))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := manager.CacheResourceGroups("sub-1", "Test Subscription", resourceGroups)
		require.NoError(b, err)
	}
}