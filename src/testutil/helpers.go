package testutil

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/matthewmyrick/azure-searcher/src/types"
	"github.com/stretchr/testify/require"
)

// TempDir creates a temporary directory for testing
func TempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "azure-searcher-test-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// TempFile creates a temporary file for testing
func TempFile(t *testing.T, content string) string {
	t.Helper()
	file, err := os.CreateTemp("", "azure-searcher-test-*.json")
	require.NoError(t, err)
	defer file.Close()
	
	_, err = file.WriteString(content)
	require.NoError(t, err)
	
	t.Cleanup(func() {
		os.Remove(file.Name())
	})
	
	return file.Name()
}

// MockSubscription creates a mock Azure subscription for testing
func MockSubscription(id, name string) types.Subscription {
	return types.Subscription{
		ID:   id,
		Name: name,
	}
}

// MockResource creates a mock Azure resource for testing
func MockResource(name, resourceType, kind, id string, tags map[string]string) types.Resource {
	if tags == nil {
		tags = make(map[string]string)
	}
	return types.Resource{
		Name:     name,
		Type:     resourceType,
		Kind:     kind,
		ID:       id,
		Tags:     tags,
		AzureURL: "https://portal.azure.com/#@/resource" + id,
	}
}

// MockResourceGroup creates a mock Azure resource group for testing
func MockResourceGroup(name, location string, resources []types.Resource, tags map[string]string) types.ResourceGroup {
	if tags == nil {
		tags = make(map[string]string)
	}
	if resources == nil {
		resources = []types.Resource{}
	}
	return types.ResourceGroup{
		Name:      name,
		Location:  location,
		Tags:      tags,
		Resources: resources,
		Expanded:  false,
	}
}

// MockCachedResourceGroup creates a mock cached resource group for testing
func MockCachedResourceGroup(name, location string, resources []types.Resource, cachedAt time.Time, tags map[string]string) types.CachedResourceGroup {
	if tags == nil {
		tags = make(map[string]string)
	}
	if resources == nil {
		resources = []types.Resource{}
	}
	return types.CachedResourceGroup{
		Name:      name,
		Location:  location,
		Tags:      tags,
		Resources: resources,
		Expanded:  false,
		CachedAt:  cachedAt,
	}
}

// MockSubscriptionCache creates a mock subscription cache for testing
func MockSubscriptionCache(subscriptionID, subscriptionName string, resourceGroups map[string]types.CachedResourceGroup, lastUpdated time.Time) types.SubscriptionCache {
	if resourceGroups == nil {
		resourceGroups = make(map[string]types.CachedResourceGroup)
	}
	return types.SubscriptionCache{
		SubscriptionID:   subscriptionID,
		SubscriptionName: subscriptionName,
		ResourceGroups:   resourceGroups,
		LastUpdated:      lastUpdated,
	}
}

// MockCacheData creates a mock cache data structure for testing
func MockCacheData(subscriptions map[string]types.SubscriptionCache, version string) types.CacheData {
	if subscriptions == nil {
		subscriptions = make(map[string]types.SubscriptionCache)
	}
	if version == "" {
		version = "1.0"
	}
	return types.CacheData{
		Subscriptions: subscriptions,
		Version:       version,
	}
}

// SetEnv sets an environment variable for testing and restores it after the test
func SetEnv(t *testing.T, key, value string) {
	t.Helper()
	original := os.Getenv(key)
	err := os.Setenv(key, value)
	require.NoError(t, err)
	
	t.Cleanup(func() {
		if original == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, original)
		}
	})
}

// CreateTestConfigDir creates a test config directory structure
func CreateTestConfigDir(t *testing.T) string {
	t.Helper()
	baseDir := TempDir(t)
	configDir := filepath.Join(baseDir, ".azure-searcher")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)
	return configDir
}

// AssertTimeAlmostEqual asserts that two times are within a reasonable delta (1 second)
func AssertTimeAlmostEqual(t *testing.T, expected, actual time.Time, msgAndArgs ...interface{}) {
	t.Helper()
	delta := time.Second
	diff := expected.Sub(actual)
	if diff < 0 {
		diff = -diff
	}
	require.True(t, diff <= delta, "Time difference %v exceeds delta %v. Expected: %v, Actual: %v", diff, delta, expected, actual)
}

// FixedTime returns a fixed time for consistent testing
func FixedTime() time.Time {
	return time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
}

// TimeAfter returns a time that's the specified duration after FixedTime
func TimeAfter(d time.Duration) time.Time {
	return FixedTime().Add(d)
}