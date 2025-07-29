package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscription_JSONSerialization(t *testing.T) {
	sub := Subscription{
		ID:   "sub-123",
		Name: "Test Subscription",
	}

	// Test marshaling
	data, err := json.Marshal(sub)
	require.NoError(t, err)

	expected := `{"id":"sub-123","name":"Test Subscription"}`
	assert.JSONEq(t, expected, string(data))

	// Test unmarshaling
	var unmarshaled Subscription
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, sub, unmarshaled)
}

func TestResource_JSONSerialization(t *testing.T) {
	resource := Resource{
		Name:     "test-vm",
		Type:     "Microsoft.Compute/virtualMachines",
		Kind:     "vm",
		ID:       "/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Compute/virtualMachines/test-vm",
		Tags:     map[string]string{"environment": "test", "owner": "john"},
		AzureURL: "https://portal.azure.com/#@/resource/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Compute/virtualMachines/test-vm",
	}

	// Test marshaling
	data, err := json.Marshal(resource)
	require.NoError(t, err)

	// Test unmarshaling
	var unmarshaled Resource
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, resource, unmarshaled)

	// Verify tags are properly handled
	assert.Equal(t, "test", unmarshaled.Tags["environment"])
	assert.Equal(t, "john", unmarshaled.Tags["owner"])
}

func TestResource_EmptyTags(t *testing.T) {
	resource := Resource{
		Name: "test-vm",
		Type: "Microsoft.Compute/virtualMachines",
		Tags: nil,
	}

	data, err := json.Marshal(resource)
	require.NoError(t, err)

	var unmarshaled Resource
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// nil map should remain nil after unmarshaling
	assert.Nil(t, unmarshaled.Tags)
}

func TestResourceGroup_JSONSerialization(t *testing.T) {
	resources := []Resource{
		{
			Name: "vm1",
			Type: "Microsoft.Compute/virtualMachines",
			Tags: map[string]string{"env": "prod"},
		},
		{
			Name: "storage1",
			Type: "Microsoft.Storage/storageAccounts",
			Tags: map[string]string{"env": "prod"},
		},
	}

	rg := ResourceGroup{
		Name:      "test-rg",
		Location:  "eastus",
		Tags:      map[string]string{"project": "test"},
		Resources: resources,
		Expanded:  true,
	}

	// Test marshaling - Expanded field should be omitted
	data, err := json.Marshal(rg)
	require.NoError(t, err)

	// Expanded field should not be in JSON
	assert.NotContains(t, string(data), "expanded")
	assert.Contains(t, string(data), "test-rg")
	assert.Contains(t, string(data), "eastus")

	// Test unmarshaling
	var unmarshaled ResourceGroup
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Expanded should be false (default value) since it's not serialized
	assert.False(t, unmarshaled.Expanded)
	assert.Equal(t, rg.Name, unmarshaled.Name)
	assert.Equal(t, rg.Location, unmarshaled.Location)
	assert.Equal(t, rg.Resources, unmarshaled.Resources)
	assert.Equal(t, rg.Tags, unmarshaled.Tags)
}

func TestCachedResourceGroup_JSONSerialization(t *testing.T) {
	cachedAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	
	cachedRG := CachedResourceGroup{
		Name:      "cached-rg",
		Location:  "westus",
		Tags:      map[string]string{"cached": "true"},
		Resources: []Resource{{Name: "test-resource", Type: "test-type"}},
		Expanded:  true,
		CachedAt:  cachedAt,
	}

	// Test marshaling
	data, err := json.Marshal(cachedRG)
	require.NoError(t, err)

	// Expanded field should not be in JSON, but CachedAt should be
	assert.NotContains(t, string(data), "expanded")
	assert.Contains(t, string(data), "cached_at")

	// Test unmarshaling
	var unmarshaled CachedResourceGroup
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.False(t, unmarshaled.Expanded) // Should be default value
	assert.Equal(t, cachedRG.Name, unmarshaled.Name)
	assert.Equal(t, cachedRG.Location, unmarshaled.Location)
	assert.Equal(t, cachedRG.Resources, unmarshaled.Resources)
	assert.Equal(t, cachedRG.Tags, unmarshaled.Tags)
	assert.True(t, cachedRG.CachedAt.Equal(unmarshaled.CachedAt))
}

func TestSubscriptionCache_JSONSerialization(t *testing.T) {
	lastUpdated := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	cachedAt := time.Date(2024, 1, 15, 10, 25, 0, 0, time.UTC)

	resourceGroups := map[string]CachedResourceGroup{
		"rg1": {
			Name:     "rg1",
			Location: "eastus",
			CachedAt: cachedAt,
		},
		"rg2": {
			Name:     "rg2",
			Location: "westus",
			CachedAt: cachedAt,
		},
	}

	subCache := SubscriptionCache{
		SubscriptionID:   "sub-123",
		SubscriptionName: "Test Sub",
		ResourceGroups:   resourceGroups,
		LastUpdated:      lastUpdated,
	}

	// Test marshaling
	data, err := json.Marshal(subCache)
	require.NoError(t, err)

	// Test unmarshaling
	var unmarshaled SubscriptionCache
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, subCache.SubscriptionID, unmarshaled.SubscriptionID)
	assert.Equal(t, subCache.SubscriptionName, unmarshaled.SubscriptionName)
	assert.Equal(t, len(subCache.ResourceGroups), len(unmarshaled.ResourceGroups))
	assert.True(t, subCache.LastUpdated.Equal(unmarshaled.LastUpdated))

	// Check individual resource groups
	for name, expectedRG := range subCache.ResourceGroups {
		actualRG, exists := unmarshaled.ResourceGroups[name]
		assert.True(t, exists)
		assert.Equal(t, expectedRG.Name, actualRG.Name)
		assert.Equal(t, expectedRG.Location, actualRG.Location)
		assert.True(t, expectedRG.CachedAt.Equal(actualRG.CachedAt))
	}
}

func TestCacheData_JSONSerialization(t *testing.T) {
	subscriptions := map[string]SubscriptionCache{
		"sub-1": {
			SubscriptionID:   "sub-1",
			SubscriptionName: "Subscription 1",
			ResourceGroups:   make(map[string]CachedResourceGroup),
			LastUpdated:      time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		"sub-2": {
			SubscriptionID:   "sub-2",
			SubscriptionName: "Subscription 2",
			ResourceGroups:   make(map[string]CachedResourceGroup),
			LastUpdated:      time.Date(2024, 1, 15, 11, 30, 0, 0, time.UTC),
		},
	}

	cacheData := CacheData{
		Subscriptions: subscriptions,
		Version:       "1.0",
	}

	// Test marshaling
	data, err := json.Marshal(cacheData)
	require.NoError(t, err)

	// Test unmarshaling
	var unmarshaled CacheData
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, cacheData.Version, unmarshaled.Version)
	assert.Equal(t, len(cacheData.Subscriptions), len(unmarshaled.Subscriptions))

	for subID, expectedSub := range cacheData.Subscriptions {
		actualSub, exists := unmarshaled.Subscriptions[subID]
		assert.True(t, exists)
		assert.Equal(t, expectedSub.SubscriptionID, actualSub.SubscriptionID)
		assert.Equal(t, expectedSub.SubscriptionName, actualSub.SubscriptionName)
		assert.True(t, expectedSub.LastUpdated.Equal(actualSub.LastUpdated))
	}
}

func TestProgressUpdate_StructureValidation(t *testing.T) {
	// Test that ProgressUpdate can be created with various states
	tests := []struct {
		name     string
		progress ProgressUpdate
	}{
		{
			name: "initial state",
			progress: ProgressUpdate{
				Total:          0,
				Processed:      0,
				ResourceGroups: nil,
				Completed:      false,
				Error:          nil,
			},
		},
		{
			name: "in progress",
			progress: ProgressUpdate{
				Total:          10,
				Processed:      5,
				ResourceGroups: []ResourceGroup{{Name: "rg1"}, {Name: "rg2"}},
				Completed:      false,
				Error:          nil,
			},
		},
		{
			name: "completed successfully",
			progress: ProgressUpdate{
				Total:          10,
				Processed:      10,
				ResourceGroups: []ResourceGroup{{Name: "rg1"}, {Name: "rg2"}},
				Completed:      true,
				Error:          nil,
			},
		},
		{
			name: "completed with error",
			progress: ProgressUpdate{
				Total:          10,
				Processed:      5,
				ResourceGroups: []ResourceGroup{{Name: "rg1"}},
				Completed:      true,
				Error:          assert.AnError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the structure can be created and accessed
			assert.Equal(t, tt.progress.Total, tt.progress.Total)
			assert.Equal(t, tt.progress.Processed, tt.progress.Processed)
			assert.Equal(t, tt.progress.Completed, tt.progress.Completed)
			assert.Equal(t, tt.progress.Error, tt.progress.Error)
			assert.Equal(t, len(tt.progress.ResourceGroups), len(tt.progress.ResourceGroups))
		})
	}
}

func TestResource_TagsHandling(t *testing.T) {
	// Test with empty tags map
	resource1 := Resource{
		Name: "test1",
		Tags: map[string]string{},
	}
	
	data1, err := json.Marshal(resource1)
	require.NoError(t, err)
	
	var unmarshaled1 Resource  
	err = json.Unmarshal(data1, &unmarshaled1)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{}, unmarshaled1.Tags)

	// Test with nil tags
	resource2 := Resource{
		Name: "test2",
		Tags: nil,
	}
	
	data2, err := json.Marshal(resource2)
	require.NoError(t, err)
	
	var unmarshaled2 Resource
	err = json.Unmarshal(data2, &unmarshaled2)
	require.NoError(t, err)
	assert.Nil(t, unmarshaled2.Tags)
}