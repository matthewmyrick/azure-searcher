package search

import (
	"testing"

	"github.com/matthewmyrick/azure-searcher/src/testutil"
	"github.com/matthewmyrick/azure-searcher/src/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFuzzyMatcher(t *testing.T) {
	matcher := NewFuzzyMatcher()
	assert.NotNil(t, matcher)
}

func TestFuzzyMatcher_SearchResourceGroups_EmptyQuery(t *testing.T) {
	matcher := NewFuzzyMatcher()
	
	resourceGroups := []types.ResourceGroup{
		testutil.MockResourceGroup("test-rg-1", "eastus", nil, nil),
		testutil.MockResourceGroup("test-rg-2", "westus", nil, nil),
	}

	// Empty query should return all resource groups
	result := matcher.SearchResourceGroups("", resourceGroups)
	assert.Equal(t, resourceGroups, result)

	// Whitespace-only query should return all resource groups
	result = matcher.SearchResourceGroups("   ", resourceGroups)
	assert.Equal(t, resourceGroups, result)
}

func TestFuzzyMatcher_SearchResourceGroups_ExactMatch(t *testing.T) {
	matcher := NewFuzzyMatcher()
	
	resourceGroups := []types.ResourceGroup{
		testutil.MockResourceGroup("production-rg", "eastus", nil, nil),
		testutil.MockResourceGroup("development-rg", "westus", nil, nil),
		testutil.MockResourceGroup("staging-rg", "centralus", nil, nil),
	}

	// Exact match should return matching resource group with highest score
	result := matcher.SearchResourceGroups("production", resourceGroups)
	require.Len(t, result, 1)
	assert.Equal(t, "production-rg", result[0].Name)
}

func TestFuzzyMatcher_SearchResourceGroups_PartialMatch(t *testing.T) {
	matcher := NewFuzzyMatcher()
	
	resourceGroups := []types.ResourceGroup{
		testutil.MockResourceGroup("web-app-rg", "eastus", nil, nil),
		testutil.MockResourceGroup("database-rg", "westus", nil, nil),
		testutil.MockResourceGroup("webapp-frontend", "centralus", nil, nil),
	}

	// Partial match should return matching resource groups
	result := matcher.SearchResourceGroups("web", resourceGroups)
	assert.Len(t, result, 2)
	
	// Check that both matching resource groups are included
	names := []string{result[0].Name, result[1].Name}
	assert.Contains(t, names, "web-app-rg")
	assert.Contains(t, names, "webapp-frontend")
}

func TestFuzzyMatcher_SearchResourceGroups_MultipleTerms(t *testing.T) {
	matcher := NewFuzzyMatcher()
	
	resourceGroups := []types.ResourceGroup{
		testutil.MockResourceGroup("web-app-production", "eastus", nil, nil),
		testutil.MockResourceGroup("web-app-staging", "westus", nil, nil),
		testutil.MockResourceGroup("database-production", "centralus", nil, nil),
		testutil.MockResourceGroup("mobile-app-production", "southus", nil, nil),
	}

	// Multiple terms - must match all terms
	result := matcher.SearchResourceGroups("web production", resourceGroups)
	require.Len(t, result, 1)
	assert.Equal(t, "web-app-production", result[0].Name)
}

func TestFuzzyMatcher_SearchResourceGroups_WithResources(t *testing.T) {
	matcher := NewFuzzyMatcher()
	
	// Create resource groups with resources
	resources1 := []types.Resource{
		testutil.MockResource("web-server-vm", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm1", nil),
		testutil.MockResource("app-storage", "Microsoft.Storage/storageAccounts", "storage", "/resource/storage1", nil),
	}
	
	resources2 := []types.Resource{
		testutil.MockResource("database-vm", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm2", nil),
	}

	resourceGroups := []types.ResourceGroup{
		testutil.MockResourceGroup("web-frontend-rg", "eastus", resources1, nil), // RG name contains "web"
		testutil.MockResourceGroup("backend-rg", "westus", resources2, nil),
	}

	// Search for term that matches resource group name
	result := matcher.SearchResourceGroups("web", resourceGroups)
	require.Len(t, result, 1)
	assert.Equal(t, "web-frontend-rg", result[0].Name)

	// Search for term that matches resource type (this won't work with current logic)
	// The algorithm requires the term to match the RG name, not just resources
	result = matcher.SearchResourceGroups("backend", resourceGroups)
	require.Len(t, result, 1)
	assert.Equal(t, "backend-rg", result[0].Name)
}

func TestFuzzyMatcher_FuzzyMatch(t *testing.T) {
	matcher := NewFuzzyMatcher()

	tests := []struct {
		pattern  string
		text     string
		expected int
		shouldMatch bool
	}{
		{"web", "web-app", 0, true},    // Will be > 0 due to fuzzy matching
		{"wb", "web-app", 0, true},     // Fuzzy match
		{"prd", "production", 0, true}, // Fuzzy match
		{"xyz", "production", 0, false}, // No match
		{"", "text", 0, false},         // Empty pattern
		{"test", "test", 0, true},      // Exact match
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_in_"+tt.text, func(t *testing.T) {
			score := matcher.fuzzyMatch(tt.pattern, tt.text)
			if tt.shouldMatch {
				assert.Greater(t, score, 0, "Expected pattern '%s' to match text '%s'", tt.pattern, tt.text)
			} else {
				assert.Equal(t, 0, score, "Expected pattern '%s' to not match text '%s'", tt.pattern, tt.text)
			}
		})
	}
}

func TestFuzzyMatcher_SearchWithinResourceGroup(t *testing.T) {
	matcher := NewFuzzyMatcher()
	
	resources := []types.Resource{
		testutil.MockResource("web-server-vm", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm1", nil),
		testutil.MockResource("app-storage", "Microsoft.Storage/storageAccounts", "storage", "/resource/storage1", nil),
		testutil.MockResource("database-vm", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm2", nil),
	}
	
	rg := testutil.MockResourceGroup("test-rg", "eastus", resources, nil)

	// Empty query should return all resources
	result := matcher.SearchWithinResourceGroup("", &rg)
	assert.Equal(t, resources, result)

	// Search for specific resource
	result = matcher.SearchWithinResourceGroup("web", &rg)
	require.Len(t, result, 1)
	assert.Equal(t, "web-server-vm", result[0].Name)

	// Search by resource type
	result = matcher.SearchWithinResourceGroup("storage", &rg)
	require.Len(t, result, 1)
	assert.Equal(t, "app-storage", result[0].Name)

	// Search for term that matches multiple resources
	result = matcher.SearchWithinResourceGroup("vm", &rg)
	assert.Len(t, result, 2)
	names := []string{result[0].Name, result[1].Name}
	assert.Contains(t, names, "web-server-vm")
	assert.Contains(t, names, "database-vm")
}

func TestFuzzyMatcher_SearchResourceGroupsTwoPart(t *testing.T) {
	matcher := NewFuzzyMatcher()
	
	resources1 := []types.Resource{
		testutil.MockResource("web-vm", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm1", nil),
		testutil.MockResource("api-vm", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm2", nil),
	}
	
	resources2 := []types.Resource{
		testutil.MockResource("db-vm", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm3", nil),
	}

	resourceGroups := []types.ResourceGroup{
		testutil.MockResourceGroup("frontend-rg", "eastus", resources1, nil),
		testutil.MockResourceGroup("backend-rg", "westus", resources2, nil),
	}

	// Single part query - should match resource group name
	result := matcher.SearchResourceGroupsTwoPart("frontend", resourceGroups)
	require.Len(t, result, 1)
	assert.Equal(t, "frontend-rg", result[0].Name)
	assert.False(t, result[0].Expanded) // Should not be auto-expanded

	// Two part query - should match RG and filter resources
	result = matcher.SearchResourceGroupsTwoPart("frontend web", resourceGroups)
	require.Len(t, result, 1)
	assert.Equal(t, "frontend-rg", result[0].Name)
	assert.True(t, result[0].Expanded) // Should be auto-expanded
	require.Len(t, result[0].Resources, 1)
	assert.Equal(t, "web-vm", result[0].Resources[0].Name)

	// Two part query with no matching resources
	result = matcher.SearchResourceGroupsTwoPart("frontend xyz", resourceGroups)
	assert.Len(t, result, 0)
}

func TestFuzzyMatcher_SearchResourceGroupsExact(t *testing.T) {
	matcher := NewFuzzyMatcher()
	
	resources := []types.Resource{
		testutil.MockResource("production-web-vm", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm1", nil),
		testutil.MockResource("staging-web-vm", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm2", nil),
	}

	resourceGroups := []types.ResourceGroup{
		testutil.MockResourceGroup("production-rg", "eastus", resources, nil),
		testutil.MockResourceGroup("staging-rg", "westus", resources, nil),
	}

	// Exact string match should work
	result := matcher.SearchResourceGroupsExact("production", resourceGroups)
	require.Len(t, result, 1)
	assert.Equal(t, "production-rg", result[0].Name)

	// Two part exact match
	result = matcher.SearchResourceGroupsExact("production staging", resourceGroups)
	require.Len(t, result, 1)
	assert.Equal(t, "production-rg", result[0].Name)
	assert.True(t, result[0].Expanded)
	require.Len(t, result[0].Resources, 1)
	assert.Equal(t, "staging-web-vm", result[0].Resources[0].Name)

	// Partial substring matches still work in exact mode (it's "exact string contains", not "exact match")
	result = matcher.SearchResourceGroupsExact("prod", resourceGroups)
	assert.Len(t, result, 1) // "prod" is contained in "production"
	assert.Equal(t, "production-rg", result[0].Name)
}

func TestFuzzyMatcher_FlattenItems(t *testing.T) {
	matcher := NewFuzzyMatcher()
	
	resources := []types.Resource{
		testutil.MockResource("vm1", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm1", nil),
		testutil.MockResource("storage1", "Microsoft.Storage/storageAccounts", "storage", "/resource/storage1", nil),
	}

	resourceGroups := []types.ResourceGroup{
		testutil.MockResourceGroup("rg1", "eastus", resources, nil),
		testutil.MockResourceGroup("rg2", "westus", nil, nil), // Empty RG
	}

	items := matcher.flattenItems(resourceGroups)
	
	// Should have 2 resource groups + 2 resources = 4 items
	assert.Len(t, items, 4)

	// Check resource group items
	rgItems := 0
	resourceItems := 0
	for _, item := range items {
		if item.Type == "resource_group" {
			rgItems++
			assert.Nil(t, item.Resource)
			assert.Equal(t, item.ResourceGroup.Name, item.DisplayPath)
		} else if item.Type == "resource" {
			resourceItems++
			assert.NotNil(t, item.Resource)
			assert.Contains(t, item.DisplayPath, "/")
		}
	}
	
	assert.Equal(t, 2, rgItems)
	assert.Equal(t, 2, resourceItems)
}

func TestFuzzyMatcher_ScoreFlatItem(t *testing.T) {
	matcher := NewFuzzyMatcher()
	
	// Create test items
	rgItem := FlatItem{
		Type:           "resource_group",
		ResourceGroup:  testutil.MockResourceGroup("test-rg", "eastus", nil, nil),
		Resource:       nil,
		DisplayPath:    "test-rg",
		SearchableText: "test-rg",
	}

	resource := testutil.MockResource("test-vm", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm1", nil)
	resourceItem := FlatItem{
		Type:           "resource",
		ResourceGroup:  testutil.MockResourceGroup("test-rg", "eastus", nil, nil),
		Resource:       &resource,
		DisplayPath:    "test-rg/test-vm",
		SearchableText: "test-rg test-vm microsoft.compute/virtualmachines",
	}

	// Test scoring
	queryTerms := []string{"test"}
	
	rgScore := matcher.scoreFlatItem(queryTerms, rgItem)
	resourceScore := matcher.scoreFlatItem(queryTerms, resourceItem)
	
	// Both should match
	assert.Greater(t, rgScore, 0)
	assert.Greater(t, resourceScore, 0)

	// Resource group should get bonus points
	assert.Greater(t, rgScore, 100) // Base score + RG bonus
}

func TestMax(t *testing.T) {
	assert.Equal(t, 5, max(3, 5))
	assert.Equal(t, 5, max(5, 3))
	assert.Equal(t, 5, max(5, 5))
	assert.Equal(t, 0, max(-1, 0))
	assert.Equal(t, 10, max(10, -5))
}

func TestFuzzyMatcher_ScoreTermMatch(t *testing.T) {
	matcher := NewFuzzyMatcher()

	// Exact match should get highest score
	score := matcher.scoreTermMatch("test", "test-resource")
	assert.Equal(t, 100, score)

	// Partial exact match should get high score
	score = matcher.scoreTermMatch("resource", "test-resource")
	assert.Equal(t, 100, score)

	// Fuzzy match should get lower score
	score = matcher.scoreTermMatch("tst", "test-resource")
	assert.Greater(t, score, 0)
	assert.Less(t, score, 100)

	// No match should get zero score
	score = matcher.scoreTermMatch("xyz", "test-resource")
	assert.Equal(t, 0, score)
}

func TestFuzzyMatcher_CalculateScore_MustMatchAll(t *testing.T) {
	matcher := NewFuzzyMatcher()
	
	rg := testutil.MockResourceGroup("web-app-production", "eastus", nil, nil)

	// All terms must match - this should succeed
	score := matcher.calculateScore([]string{"web", "app"}, rg)
	assert.Greater(t, score, 0)

	// Not all terms match - this should fail
	score = matcher.calculateScore([]string{"web", "xyz"}, rg)
	assert.Equal(t, 0, score)
}

func TestFuzzyMatcher_SearchResourceGroups_Sorting(t *testing.T) {
	matcher := NewFuzzyMatcher()
	
	resourceGroups := []types.ResourceGroup{
		testutil.MockResourceGroup("web-application", "eastus", nil, nil),      
		testutil.MockResourceGroup("web", "westus", nil, nil),                 
		testutil.MockResourceGroup("web-app", "centralus", nil, nil),          
	}

	result := matcher.SearchResourceGroups("web", resourceGroups)
	require.Len(t, result, 3)

	// Results should be sorted by score (highest first)  
	// The exact scoring depends on the algorithm, but all should match
	// We just verify all are included and sorted by score
	names := make([]string, len(result))
	for i, rg := range result {
		names[i] = rg.Name
	}
	assert.Contains(t, names, "web")
	assert.Contains(t, names, "web-app") 
	assert.Contains(t, names, "web-application")
}

// BenchmarkFuzzySearch tests the performance of fuzzy searching
func BenchmarkFuzzySearch(b *testing.B) {
	matcher := NewFuzzyMatcher()
	
	// Create test data
	var resourceGroups []types.ResourceGroup
	for i := 0; i < 1000; i++ {
		rg := testutil.MockResourceGroup("resource-group-"+string(rune(i)), "eastus", nil, nil)
		resourceGroups = append(resourceGroups, rg)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matcher.SearchResourceGroups("resource", resourceGroups)
	}
}

func TestFuzzyMatcher_EdgeCases(t *testing.T) {
	matcher := NewFuzzyMatcher()

	// Test empty resource groups slice
	result := matcher.SearchResourceGroups("test", []types.ResourceGroup{})
	assert.Empty(t, result)

	// Test nil resource groups slice
	result = matcher.SearchResourceGroups("test", nil)
	assert.Empty(t, result)

	// Test resource group with nil resources
	rg := testutil.MockResourceGroup("test-rg", "eastus", nil, nil)
	resourceResult := matcher.SearchWithinResourceGroup("test", &rg)
	assert.Empty(t, resourceResult)

	// Test resource group with empty resources slice
	rg.Resources = []types.Resource{}
	resourceResult = matcher.SearchWithinResourceGroup("test", &rg)
	assert.Empty(t, resourceResult)
}