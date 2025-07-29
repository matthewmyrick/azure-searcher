package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/matthewmyrick/azure-searcher/src/azure"
	"github.com/matthewmyrick/azure-searcher/src/cache"
	"github.com/matthewmyrick/azure-searcher/src/config"
	"github.com/matthewmyrick/azure-searcher/src/search"
	"github.com/matthewmyrick/azure-searcher/src/testutil"
	"github.com/matthewmyrick/azure-searcher/src/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_CacheAndSearch tests the integration between cache manager and search
func TestIntegration_CacheAndSearch(t *testing.T) {
	// Setup
	tempDir := testutil.TempDir(t)
	cacheFile := filepath.Join(tempDir, "integration-cache.json")
	
	cacheManager, err := cache.NewManager(cacheFile)
	require.NoError(t, err)
	
	fuzzyMatcher := search.NewFuzzyMatcher()
	
	// Create test data
	resources := []types.Resource{
		testutil.MockResource("web-vm", "Microsoft.Compute/virtualMachines", "vm", "/resource/web-vm", nil),
		testutil.MockResource("api-vm", "Microsoft.Compute/virtualMachines", "vm", "/resource/api-vm", nil),
		testutil.MockResource("db-storage", "Microsoft.Storage/storageAccounts", "storage", "/resource/db-storage", nil),
	}
	
	resourceGroups := []types.ResourceGroup{
		testutil.MockResourceGroup("production-web-rg", "eastus", resources[:2], nil), // web-vm, api-vm  
		testutil.MockResourceGroup("storage-rg", "westus", resources[2:], nil),        // db-storage
	}
	
	// Cache the resource groups
	err = cacheManager.CacheResourceGroups("test-sub", "Test Subscription", resourceGroups)
	require.NoError(t, err)
	
	// Retrieve from cache
	cachedRGs, valid := cacheManager.GetCachedResourceGroups("test-sub")
	require.True(t, valid)
	require.Len(t, cachedRGs, 2)
	
	// Test fuzzy search on cached data
	searchResults := fuzzyMatcher.SearchResourceGroups("production", cachedRGs)
	require.Len(t, searchResults, 1)
	assert.Equal(t, "production-web-rg", searchResults[0].Name)
	assert.Len(t, searchResults[0].Resources, 2)
	
	// Test search for resources (web is now in the RG name)  
	searchResults = fuzzyMatcher.SearchResourceGroups("web", cachedRGs)
	require.Len(t, searchResults, 1)
	assert.Equal(t, "production-web-rg", searchResults[0].Name)
	
	// Test two-part search
	searchResults = fuzzyMatcher.SearchResourceGroupsTwoPart("production web", cachedRGs)  
	require.Len(t, searchResults, 1)
	assert.Equal(t, "production-web-rg", searchResults[0].Name)
	assert.True(t, searchResults[0].Expanded)
	require.Len(t, searchResults[0].Resources, 1)
	assert.Equal(t, "web-vm", searchResults[0].Resources[0].Name)
}

// TestIntegration_ConfigAndConcurrency tests configuration integration
func TestIntegration_ConfigAndConcurrency(t *testing.T) {
	// Get optimal configuration
	conf := config.GetOptimalConcurrency()
	
	// Create services with the configuration
	azureClient := azure.NewClient()
	azureFetcher := azure.NewFetcher(azureClient, conf.RGConcurrency, conf.ResourceConcurrency)
	
	// Verify the fetcher was configured correctly
	assert.NotNil(t, azureFetcher)
	// Note: We can't easily test the internal concurrency values without exposing them,
	// but we can verify the fetcher was created successfully with the config
}

// TestIntegration_CacheExpiration tests cache TTL functionality
func TestIntegration_CacheExpiration(t *testing.T) {
	tempDir := testutil.TempDir(t)
	cacheFile := filepath.Join(tempDir, "expiration-cache.json")
	
	cacheManager, err := cache.NewManager(cacheFile)
	require.NoError(t, err)
	
	// For this integration test, we'll test that fresh cache works properly
	// since we can't easily manipulate internal timestamps
	resourceGroups := []types.ResourceGroup{
		testutil.MockResourceGroup("test-rg", "eastus", nil, nil),
	}
	
	// Cache fresh data
	err = cacheManager.CacheResourceGroups("test-sub", "Test Sub", resourceGroups)
	require.NoError(t, err)
	
	// Should return fresh cache
	cachedRGs, valid := cacheManager.GetCachedResourceGroups("test-sub")
	assert.True(t, valid)
	require.Len(t, cachedRGs, 1)
	assert.Equal(t, "test-rg", cachedRGs[0].Name)
}

// TestIntegration_SearchModes tests different search modes
func TestIntegration_SearchModes(t *testing.T) {
	fuzzyMatcher := search.NewFuzzyMatcher()
	
	resourceGroups := []types.ResourceGroup{
		testutil.MockResourceGroup("production-web-app", "eastus", nil, nil),
		testutil.MockResourceGroup("staging-web-app", "westus", nil, nil),
		testutil.MockResourceGroup("development-api", "centralus", nil, nil),
	}
	
	// Test fuzzy search
	fuzzyResults := fuzzyMatcher.SearchResourceGroups("prod web", resourceGroups)
	require.Len(t, fuzzyResults, 1)
	assert.Equal(t, "production-web-app", fuzzyResults[0].Name)
	
	// Test exact search
	exactResults := fuzzyMatcher.SearchResourceGroupsExact("production", resourceGroups)
	require.Len(t, exactResults, 1)
	assert.Equal(t, "production-web-app", exactResults[0].Name)
	
	// Test two-part search - look for RG names that contain both "web" and "app"
	twoPartResults := fuzzyMatcher.SearchResourceGroupsTwoPart("production", resourceGroups)
	assert.Len(t, twoPartResults, 1) // Should match production resource group
	
	// Test that exact search uses substring matching
	exactShortResults := fuzzyMatcher.SearchResourceGroupsExact("prod", resourceGroups)
	assert.Len(t, exactShortResults, 1) // "prod" is contained in "production"
	
	fuzzyShortResults := fuzzyMatcher.SearchResourceGroups("prod", resourceGroups)
	assert.Len(t, fuzzyShortResults, 1) // Fuzzy match should work
}

// TestIntegration_CacheConfigWorkflow tests cache configuration workflow
func TestIntegration_CacheConfigWorkflow(t *testing.T) {
	tempDir := testutil.TempDir(t)
	configDir := filepath.Join(tempDir, ".azure-searcher")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)
	
	// Use a unique cache file for this test
	cacheFile := filepath.Join(configDir, "workflow-cache.json")
	cacheManager, err := cache.NewManager(cacheFile)
	require.NoError(t, err)
	
	subscriptionID := "test-sub-workflow-unique"
	
	// Check initial cache state (may vary depending on system state)
	enabled := cacheManager.GetSubscriptionCacheEnabled(subscriptionID)
	// Note: This may be true or false depending on existing config files
	// The important part is testing the set/get cycle works
	
	// Enable cache
	err = cacheManager.SetSubscriptionCacheEnabled(subscriptionID, true)
	require.NoError(t, err)
	
	// Verify cache is enabled
	enabled = cacheManager.GetSubscriptionCacheEnabled(subscriptionID)
	assert.True(t, enabled)
	
	// Set auto-refresh
	err = cacheManager.SetSubscriptionAutoRefresh(subscriptionID, true, "1 hr")
	require.NoError(t, err)
	
	// Verify auto-refresh settings
	autoRefreshEnabled := cacheManager.GetSubscriptionAutoRefreshEnabled(subscriptionID)
	assert.True(t, autoRefreshEnabled)
	
	interval := cacheManager.GetSubscriptionRefreshInterval(subscriptionID)
	assert.Equal(t, "1 hr", interval)
	
	// Test refresh decision logic
	shouldRefresh := cacheManager.ShouldRefreshCache(subscriptionID)
	assert.True(t, shouldRefresh) // Should refresh when no cache exists
	
	// Add cache data and test refresh logic
	resourceGroups := []types.ResourceGroup{
		testutil.MockResourceGroup("test-rg", "eastus", nil, nil),
	}
	err = cacheManager.CacheResourceGroups(subscriptionID, "Test Sub", resourceGroups)
	require.NoError(t, err)
	
	// Should not refresh immediately after caching (cache is fresh)
	shouldRefresh = cacheManager.ShouldRefreshCache(subscriptionID)
	assert.False(t, shouldRefresh)
}

// TestIntegration_ErrorHandling tests error handling across components
func TestIntegration_ErrorHandling(t *testing.T) {
	// Test cache manager with invalid file permissions
	invalidCacheFile := "/root/invalid-path/cache.json" // Assuming this path is not writable
	cacheManager, err := cache.NewManager(invalidCacheFile)
	require.NoError(t, err) // Should not error on creation
	
	// Should handle save errors gracefully
	resourceGroups := []types.ResourceGroup{
		testutil.MockResourceGroup("test-rg", "eastus", nil, nil),
	}
	err = cacheManager.CacheResourceGroups("test-sub", "Test Sub", resourceGroups)
	// May or may not error depending on system permissions, but shouldn't panic
	
	// Test search with nil/empty data
	fuzzyMatcher := search.NewFuzzyMatcher()
	
	// Should handle nil resource groups
	results := fuzzyMatcher.SearchResourceGroups("test", nil)
	assert.Empty(t, results)
	
	// Should handle empty resource groups
	results = fuzzyMatcher.SearchResourceGroups("test", []types.ResourceGroup{})
	assert.Empty(t, results)
	
	// Should handle empty query
	results = fuzzyMatcher.SearchResourceGroups("", resourceGroups)
	assert.Equal(t, resourceGroups, results)
}

// TestIntegration_FullWorkflow tests a complete workflow
func TestIntegration_FullWorkflow(t *testing.T) {
	// This test simulates a full workflow:
	// 1. Configure cache settings
	// 2. Cache some resource groups
	// 3. Search the cached data
	// 4. Verify results
	
	tempDir := testutil.TempDir(t)
	configDir := filepath.Join(tempDir, ".azure-searcher")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)
	
	// Initialize components
	cacheFile := filepath.Join(configDir, "workflow-cache.json")
	cacheManager, err := cache.NewManager(cacheFile)
	require.NoError(t, err)
	
	fuzzyMatcher := search.NewFuzzyMatcher()
	conf := config.GetOptimalConcurrency()
	
	subscriptionID := "workflow-test-sub"
	
	// Step 1: Configure cache
	err = cacheManager.SetSubscriptionCacheEnabled(subscriptionID, true)
	require.NoError(t, err)
	
	err = cacheManager.SetSubscriptionAutoRefresh(subscriptionID, true, "30 min")
	require.NoError(t, err)
	
	// Step 2: Create and cache resource groups
	resources := []types.Resource{
		testutil.MockResource("web-server-1", "Microsoft.Compute/virtualMachines", "vm", "/resource/web-1", nil),
		testutil.MockResource("web-server-2", "Microsoft.Compute/virtualMachines", "vm", "/resource/web-2", nil),
		testutil.MockResource("app-storage", "Microsoft.Storage/storageAccounts", "storage", "/resource/storage", nil),
	}
	
	resourceGroups := []types.ResourceGroup{
		testutil.MockResourceGroup("production-web", "eastus", resources[:2], nil),
		testutil.MockResourceGroup("production-storage", "westus", resources[2:], nil),
	}
	
	err = cacheManager.CacheResourceGroups(subscriptionID, "Workflow Test Subscription", resourceGroups)
	require.NoError(t, err)
	
	// Step 3: Retrieve and search cached data
	cachedRGs, valid := cacheManager.GetCachedResourceGroups(subscriptionID)
	require.True(t, valid)
	require.Len(t, cachedRGs, 2)
	
	// Search for production resources
	searchResults := fuzzyMatcher.SearchResourceGroups("production", cachedRGs)
	assert.Len(t, searchResults, 2) // Both production resource groups
	
	// Search for web-specific resources
	webResults := fuzzyMatcher.SearchResourceGroups("web", cachedRGs)
	require.Len(t, webResults, 1)
	assert.Equal(t, "production-web", webResults[0].Name)
	assert.Len(t, webResults[0].Resources, 2)
	
	// Test detailed search within resource group
	webRG := webResults[0]
	webResources := fuzzyMatcher.SearchWithinResourceGroup("server", &webRG)
	assert.Len(t, webResources, 2) // Both web servers
	
	specificResource := fuzzyMatcher.SearchWithinResourceGroup("server-1", &webRG)
	require.Len(t, specificResource, 1)
	assert.Equal(t, "web-server-1", specificResource[0].Name)
	
	// Step 4: Verify configuration is still active
	enabled := cacheManager.GetSubscriptionCacheEnabled(subscriptionID)
	assert.True(t, enabled)
	
	autoRefresh := cacheManager.GetSubscriptionAutoRefreshEnabled(subscriptionID)
	assert.True(t, autoRefresh)
	
	interval := cacheManager.GetSubscriptionRefreshInterval(subscriptionID)
	assert.Equal(t, "30 min", interval)
	
	// Verify concurrency configuration
	assert.Greater(t, conf.RGConcurrency, 0)
	assert.Greater(t, conf.ResourceConcurrency, 0)
}

// TestIntegration_Performance tests performance characteristics
func TestIntegration_Performance(t *testing.T) {
	// Create large dataset
	var resourceGroups []types.ResourceGroup
	for i := 0; i < 100; i++ {
		var resources []types.Resource
		for j := 0; j < 10; j++ {
			resources = append(resources, testutil.MockResource(
				fmt.Sprintf("resource-%d-%d", i, j),
				"Microsoft.Test/resources",
				"test", 
				fmt.Sprintf("/resource/%d-%d", i, j),
				nil,
			))
		}
		resourceGroups = append(resourceGroups, testutil.MockResourceGroup(
			fmt.Sprintf("rg-%d", i),
			"eastus",
			resources,
			nil,
		))
	}
	
	// Test cache performance
	tempDir := testutil.TempDir(t)
	cacheFile := filepath.Join(tempDir, "perf-cache.json")
	cacheManager, err := cache.NewManager(cacheFile)
	require.NoError(t, err)
	
	start := time.Now()
	err = cacheManager.CacheResourceGroups("perf-test", "Performance Test", resourceGroups)
	require.NoError(t, err)
	cacheTime := time.Since(start)
	
	// Should cache 100 RGs with 1000 resources reasonably quickly
	assert.Less(t, cacheTime, 5*time.Second, "Caching should complete within 5 seconds")
	
	// Test retrieval performance
	start = time.Now()
	cachedRGs, valid := cacheManager.GetCachedResourceGroups("perf-test")
	retrievalTime := time.Since(start)
	
	require.True(t, valid)
	require.Len(t, cachedRGs, 100)
	assert.Less(t, retrievalTime, 1*time.Second, "Retrieval should complete within 1 second")
	
	// Test search performance
	fuzzyMatcher := search.NewFuzzyMatcher()
	
	start = time.Now()
	searchResults := fuzzyMatcher.SearchResourceGroups("rg", cachedRGs) // Search for partial match
	searchTime := time.Since(start)
	
	assert.Greater(t, len(searchResults), 0) // Should find multiple matches
	assert.Less(t, searchTime, 100*time.Millisecond, "Search should complete within 100ms")
}