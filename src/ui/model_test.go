package ui

import (
	"testing"

	"github.com/matthewmyrick/azure-searcher/src/testutil"
	"github.com/matthewmyrick/azure-searcher/src/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	model, err := NewModel()
	require.NoError(t, err)
	assert.NotNil(t, model)
	
	// Test initial state
	assert.Equal(t, "subscriptions", model.State)
	assert.Equal(t, "exact", model.SearchMode) // Default is exact, not fuzzy
	assert.False(t, model.ShowResourceTypes)

	// Test UI components are initialized
	assert.NotNil(t, model.SearchInput)
	assert.NotNil(t, model.FilterInput)
	assert.NotNil(t, model.Spinner)
	assert.NotNil(t, model.Progress)
	assert.NotNil(t, model.IntervalInput)

	// Test services are initialized
	assert.NotNil(t, model.AzureClient)
	assert.NotNil(t, model.AzureFetcher)
	assert.NotNil(t, model.CacheManager)
	assert.NotNil(t, model.FuzzyMatcher)

	// Test initial values
	assert.Equal(t, 0, model.Cursor)
	assert.Equal(t, 0, model.ScrollOffset)
	assert.Nil(t, model.Err)
	assert.False(t, model.LastLoadFromCache)
	assert.Equal(t, 0, model.TotalRGs)
	assert.Equal(t, 0, model.ProcessedRGs)
}

func TestModel_InitialSearchInputSetup(t *testing.T) {
	model, err := NewModel()
	require.NoError(t, err)

	// Test search input configuration
	searchInput := model.SearchInput
	assert.Contains(t, searchInput.Placeholder, "Search:")
	assert.Contains(t, searchInput.Placeholder, "resourcegroup")
	assert.Equal(t, 50, searchInput.CharLimit)
	assert.Equal(t, 50, searchInput.Width)
}

func TestModel_InitialFilterInputSetup(t *testing.T) {
	model, err := NewModel()
	require.NoError(t, err)

	// Test filter input configuration
	filterInput := model.FilterInput
	assert.Contains(t, filterInput.Placeholder, "Filter:")
	assert.Contains(t, filterInput.Placeholder, "tags=")
	assert.Equal(t, 100, filterInput.CharLimit)
	assert.Equal(t, 50, filterInput.Width)
}

func TestModel_InitialIntervalInputSetup(t *testing.T) {
	model, err := NewModel()
	require.NoError(t, err)

	// Test interval input configuration
	intervalInput := model.IntervalInput
	assert.Contains(t, intervalInput.Placeholder, "refresh interval")
	assert.Contains(t, intervalInput.Placeholder, "2 hr")
	assert.Equal(t, 20, intervalInput.CharLimit)
	assert.Equal(t, 40, intervalInput.Width)
}

func TestModel_StateManagement(t *testing.T) {
	model, err := NewModel()
	require.NoError(t, err)

	// Test initial state
	assert.Equal(t, "subscriptions", model.State)

	// Test state transitions
	validStates := []string{"subscriptions", "loading", "resources", "config"}
	for _, state := range validStates {
		model.State = state
		assert.Equal(t, state, model.State)
	}
}

func TestModel_SearchModeManagement(t *testing.T) {
	model, err := NewModel()
	require.NoError(t, err)

	// Test initial search mode
	assert.Equal(t, "exact", model.SearchMode) // Default is exact

	// Test search mode switching
	model.SearchMode = "fuzzy"
	assert.Equal(t, "fuzzy", model.SearchMode)

	model.SearchMode = "exact"
	assert.Equal(t, "exact", model.SearchMode)
}

func TestModel_SubscriptionHandling(t *testing.T) {
	model, err := NewModel()
	require.NoError(t, err)

	// Test empty subscriptions initially
	assert.Empty(t, model.Subscriptions)
	assert.Equal(t, "", model.SelectedSub.ID)

	// Test setting subscriptions
	subscriptions := []types.Subscription{
		testutil.MockSubscription("sub-1", "Test Subscription 1"),
		testutil.MockSubscription("sub-2", "Test Subscription 2"),
	}
	
	model.Subscriptions = subscriptions
	assert.Len(t, model.Subscriptions, 2)
	assert.Equal(t, "sub-1", model.Subscriptions[0].ID)

	// Test setting selected subscription
	model.SelectedSub = subscriptions[0]
	assert.Equal(t, "sub-1", model.SelectedSub.ID)
	assert.Equal(t, "Test Subscription 1", model.SelectedSub.Name)
}

func TestModel_ResourceGroupHandling(t *testing.T) {
	model, err := NewModel()
	require.NoError(t, err)

	// Test empty resource groups initially
	assert.Empty(t, model.ResourceGroups)
	assert.Empty(t, model.FilteredGroups)

	// Create test resource groups
	resourceGroups := []types.ResourceGroup{
		testutil.MockResourceGroup("rg-1", "eastus", nil, nil),
		testutil.MockResourceGroup("rg-2", "westus", nil, nil),
	}

	// Test setting resource groups
	model.ResourceGroups = resourceGroups
	assert.Len(t, model.ResourceGroups, 2)
	assert.Equal(t, "rg-1", model.ResourceGroups[0].Name)

	// Test setting filtered groups
	model.FilteredGroups = resourceGroups[:1] // Only first group
	assert.Len(t, model.FilteredGroups, 1)
	assert.Equal(t, "rg-1", model.FilteredGroups[0].Name)
}

func TestModel_NavigationState(t *testing.T) {
	model, err := NewModel()
	require.NoError(t, err)

	// Test initial navigation state
	assert.Equal(t, 0, model.Cursor)
	assert.Equal(t, 0, model.ScrollOffset)
	assert.Equal(t, 20, model.ViewportHeight) // Has default value
	assert.Equal(t, 80, model.ViewportWidth)  // Has default value

	// Test navigation updates
	model.Cursor = 5
	model.ScrollOffset = 2
	model.ViewportHeight = 24
	model.ViewportWidth = 80

	assert.Equal(t, 5, model.Cursor)
	assert.Equal(t, 2, model.ScrollOffset)
	assert.Equal(t, 24, model.ViewportHeight)
	assert.Equal(t, 80, model.ViewportWidth)
}

func TestModel_ConfigurationState(t *testing.T) {
	model, err := NewModel()
	require.NoError(t, err)

	// Test initial configuration state
	assert.Equal(t, "", model.ConfiguringSub.ID)
	assert.Equal(t, 0, model.ConfigCursor)
	assert.Equal(t, "menu", model.ConfigMode) // Has default value

	// Test configuration updates
	testSub := testutil.MockSubscription("sub-config", "Config Subscription")
	model.ConfiguringSub = testSub
	model.ConfigCursor = 3
	model.ConfigMode = "interval_input"

	assert.Equal(t, "sub-config", model.ConfiguringSub.ID)
	assert.Equal(t, 3, model.ConfigCursor)
	assert.Equal(t, "interval_input", model.ConfigMode)
}

func TestModel_ProgressTracking(t *testing.T) {
	model, err := NewModel()
	require.NoError(t, err)

	// Test initial progress state
	assert.False(t, model.LastLoadFromCache)
	assert.Equal(t, 0, model.TotalRGs)
	assert.Equal(t, 0, model.ProcessedRGs)
	assert.Nil(t, model.ProgressChan)

	// Test progress updates
	model.LastLoadFromCache = true
	model.TotalRGs = 10
	model.ProcessedRGs = 5

	assert.True(t, model.LastLoadFromCache)
	assert.Equal(t, 10, model.TotalRGs)
	assert.Equal(t, 5, model.ProcessedRGs)

	// Test progress channel creation
	progressChan := make(chan types.ProgressUpdate, 100)
	model.ProgressChan = progressChan
	assert.NotNil(t, model.ProgressChan)
}

func TestModel_ErrorHandling(t *testing.T) {
	model, err := NewModel()
	require.NoError(t, err)

	// Test initial error state
	assert.Nil(t, model.Err)

	// Test setting error
	testError := assert.AnError
	model.Err = testError
	assert.Equal(t, testError, model.Err)

	// Test clearing error
	model.Err = nil
	assert.Nil(t, model.Err)
}

func TestModel_ServiceInitialization(t *testing.T) {
	model, err := NewModel()
	require.NoError(t, err)

	// Test that all services are properly initialized and not nil
	assert.NotNil(t, model.AzureClient, "AzureClient should be initialized")
	assert.NotNil(t, model.AzureFetcher, "AzureFetcher should be initialized")
	assert.NotNil(t, model.CacheManager, "CacheManager should be initialized")
	assert.NotNil(t, model.FuzzyMatcher, "FuzzyMatcher should be initialized")

	// Test configuration is loaded
	assert.Greater(t, model.Config.RGConcurrency, 0, "RG concurrency should be greater than 0")
	assert.Greater(t, model.Config.ResourceConcurrency, 0, "Resource concurrency should be greater than 0")
}

func TestModel_UIComponentStyling(t *testing.T) {
	model, err := NewModel()
	require.NoError(t, err)

	// Test that UI components have reasonable configurations
	assert.Greater(t, model.SearchInput.Width, 0, "Search input should have width")
	assert.Greater(t, model.FilterInput.Width, 0, "Filter input should have width")
	assert.Greater(t, model.IntervalInput.Width, 0, "Interval input should have width")
	
	assert.Greater(t, model.SearchInput.CharLimit, 0, "Search input should have char limit")
	assert.Greater(t, model.FilterInput.CharLimit, 0, "Filter input should have char limit")
	assert.Greater(t, model.IntervalInput.CharLimit, 0, "Interval input should have char limit")
}

func TestModel_DefaultValues(t *testing.T) {
	model, err := NewModel()
	require.NoError(t, err)

	// Test default values are sensible
	assert.Equal(t, "subscriptions", model.State, "Should start in subscriptions state")
	assert.Equal(t, "exact", model.SearchMode, "Should default to exact search")
	assert.False(t, model.ShowResourceTypes, "Should not show resource types initially")
	assert.False(t, model.LastLoadFromCache, "Should not have loaded from cache initially")
	
	// Test numeric defaults
	assert.Equal(t, 0, model.Cursor, "Cursor should start at 0")
	assert.Equal(t, 0, model.ScrollOffset, "Scroll offset should start at 0")
	assert.Equal(t, 0, model.ConfigCursor, "Config cursor should start at 0")
	assert.Equal(t, 0, model.TotalRGs, "Total RGs should start at 0")
	assert.Equal(t, 0, model.ProcessedRGs, "Processed RGs should start at 0")
}

// Test that the model implements the tea.Model interface
func TestModel_ImplementsTeaModel(t *testing.T) {
	model, err := NewModel()
	require.NoError(t, err)

	// Test that the model can be used as a tea.Model
	var teaModel tea.Model = model
	assert.NotNil(t, teaModel)

	// The model should have Update and View methods (tested implicitly by interface conformance)
	// We can't easily test Update and View without setting up the full tea environment,
	// but we can verify the methods exist by checking interface compliance
}

func TestModel_Cleanup(t *testing.T) {
	model, err := NewModel()
	require.NoError(t, err)

	// Test that resources can be cleaned up properly
	progressChan := make(chan types.ProgressUpdate, 100)
	model.ProgressChan = progressChan

	// Simulate cleanup
	if model.ProgressChan != nil {
		close(model.ProgressChan)
		model.ProgressChan = nil
	}

	assert.Nil(t, model.ProgressChan, "Progress channel should be cleaned up")
}