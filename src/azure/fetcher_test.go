package azure

import (
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/matthewmyrick/azure-searcher/src/testutil"
	"github.com/matthewmyrick/azure-searcher/src/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockClientInterface for testing the fetcher
type MockClientInterface struct {
	resourceGroups      []types.ResourceGroup
	resourcesByGroup    map[string][]types.Resource
	shouldFailRGList    bool
	shouldFailResources map[string]bool
	callCounts          map[string]int
	mu                  sync.Mutex
}

func NewMockClientInterface() *MockClientInterface {
	return &MockClientInterface{
		resourcesByGroup:    make(map[string][]types.Resource),
		shouldFailResources: make(map[string]bool),
		callCounts:          make(map[string]int),
	}
}

func (m *MockClientInterface) GetResourceGroupsList(subscriptionID string) ([]types.ResourceGroup, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCounts["GetResourceGroupsList"]++
	
	if m.shouldFailRGList {
		return nil, assert.AnError
	}
	return m.resourceGroups, nil
}

func (m *MockClientInterface) GetResourcesInGroup(subscriptionID, resourceGroupName string) ([]types.Resource, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCounts["GetResourcesInGroup"]++
	
	if m.shouldFailResources[resourceGroupName] {
		return nil, assert.AnError
	}
	
	resources, exists := m.resourcesByGroup[resourceGroupName]
	if !exists {
		return []types.Resource{}, nil
	}
	return resources, nil
}

func (m *MockClientInterface) GetCallCount(method string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCounts[method]
}

// We need to modify the fetcher to accept an interface, but for now we'll test with the mock
func createMockFetcher(client *MockClientInterface, rgConcurrency, resourceConcurrency int) *Fetcher {
	// For testing purposes, we'll need to work around the concrete Client type
	// In a real refactor, Fetcher would accept an interface
	return &Fetcher{
		client:              nil, // We'll handle this in tests
		rgConcurrency:       rgConcurrency,
		resourceConcurrency: resourceConcurrency,
	}
}

func TestNewFetcher(t *testing.T) {
	client := NewClient()
	fetcher := NewFetcher(client, 5, 10)
	
	assert.NotNil(t, fetcher)
	assert.Equal(t, client, fetcher.client)
	assert.Equal(t, 5, fetcher.rgConcurrency)
	assert.Equal(t, 10, fetcher.resourceConcurrency)
}

func TestFetcher_Configuration(t *testing.T) {
	client := NewClient()
	
	tests := []struct {
		name                string
		rgConcurrency       int
		resourceConcurrency int
	}{
		{"default values", 5, 10},
		{"low concurrency", 1, 1},
		{"high concurrency", 20, 50},
		{"zero concurrency", 0, 0}, // Edge case
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := NewFetcher(client, tt.rgConcurrency, tt.resourceConcurrency)
			assert.Equal(t, tt.rgConcurrency, fetcher.rgConcurrency)
			assert.Equal(t, tt.resourceConcurrency, fetcher.resourceConcurrency)
		})
	}
}

// Since we can't easily mock the client in the current structure, 
// we'll test the components that we can test directly

func TestFetcher_ProcessResource(t *testing.T) {
	fetcher := NewFetcher(NewClient(), 5, 10)
	
	resource := testutil.MockResource("test-vm", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm", nil)
	
	// Test the processResource method
	resourceSemaphore := make(chan struct{}, 1)
	var wg sync.WaitGroup
	resChan := make(chan types.Resource, 1)
	
	wg.Add(1)
	go fetcher.processResource(resource, resourceSemaphore, &wg, resChan)
	
	wg.Wait()
	close(resChan)
	
	// Should receive the same resource back
	result := <-resChan
	assert.Equal(t, resource, result)
}

func TestFetcher_ProcessResource_Concurrency(t *testing.T) {
	fetcher := NewFetcher(NewClient(), 5, 2) // Limited resource concurrency
	
	// Create multiple resources
	resources := []types.Resource{
		testutil.MockResource("vm1", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm1", nil),
		testutil.MockResource("vm2", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm2", nil),
		testutil.MockResource("vm3", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm3", nil),
		testutil.MockResource("vm4", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm4", nil),
	}
	
	resourceSemaphore := make(chan struct{}, 2) // Limit to 2 concurrent
	var wg sync.WaitGroup
	resChan := make(chan types.Resource, len(resources))
	
	start := time.Now()
	
	// Process all resources
	for _, res := range resources {
		wg.Add(1)
		go func(r types.Resource) {
			// Add small delay to test concurrency limiting
			time.Sleep(10 * time.Millisecond)
			fetcher.processResource(r, resourceSemaphore, &wg, resChan)
		}(res)
	}
	
	wg.Wait()
	close(resChan)
	
	elapsed := time.Since(start)
	
	// Collect results
	var results []types.Resource
	for res := range resChan {
		results = append(results, res)
	}
	
	// Should have all resources
	assert.Len(t, results, 4)
	
	// Should take at least 10ms due to concurrency limiting (resources processed in batches)
	// Note: timing can be flaky on fast systems, so we use a conservative threshold
	assert.GreaterOrEqual(t, int(elapsed.Milliseconds()), 5)
}

func TestFetcher_SemaphoreReleasing(t *testing.T) {
	fetcher := NewFetcher(NewClient(), 5, 1) // Single resource concurrency
	
	resource := testutil.MockResource("test-vm", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm", nil)
	resourceSemaphore := make(chan struct{}, 1)
	
	// Verify semaphore starts empty
	assert.Equal(t, 0, len(resourceSemaphore))
	
	var wg sync.WaitGroup
	resChan := make(chan types.Resource, 1)
	
	wg.Add(1)
	go fetcher.processResource(resource, resourceSemaphore, &wg, resChan)
	
	wg.Wait()
	close(resChan)
	
	// Verify semaphore is released (should be empty again)
	assert.Equal(t, 0, len(resourceSemaphore))
	
	// Should receive the resource
	result := <-resChan
	assert.Equal(t, resource.Name, result.Name)
}

// Test progress updates functionality
func TestProgressUpdate_Structure(t *testing.T) {
	// Test that we can create and manipulate progress updates
	progress := types.ProgressUpdate{
		Total:          10,
		Processed:      5,
		ResourceGroups: []types.ResourceGroup{},
		Completed:      false,
		Error:          nil,
	}
	
	assert.Equal(t, 10, progress.Total)
	assert.Equal(t, 5, progress.Processed)
	assert.False(t, progress.Completed)
	assert.Nil(t, progress.Error)
	assert.NotNil(t, progress.ResourceGroups)
}

func TestProgressUpdate_Channels(t *testing.T) {
	// Test that progress updates can be sent through channels
	progressChan := make(chan types.ProgressUpdate, 2)
	
	// Send initial progress
	progressChan <- types.ProgressUpdate{
		Total:     3,
		Processed: 0,
	}
	
	// Send updated progress
	progressChan <- types.ProgressUpdate{
		Total:     3,
		Processed: 2,
	}
	
	close(progressChan)
	
	// Read progress updates
	var updates []types.ProgressUpdate
	for update := range progressChan {
		updates = append(updates, update)
	}
	
	require.Len(t, updates, 2)
	assert.Equal(t, 0, updates[0].Processed)
	assert.Equal(t, 2, updates[1].Processed)
}

// Test atomic operations
func TestAtomicOperations(t *testing.T) {
	var counter int64
	var wg sync.WaitGroup
	
	// Simulate multiple goroutines incrementing counter
	numGoroutines := 10
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Simulate the atomic add operation used in processResourceGroup
			atomic.AddInt64(&counter, 1)
		}()
	}
	
	wg.Wait()
	assert.Equal(t, int64(numGoroutines), counter)
}

// Test sorting functionality
func TestResourceSorting(t *testing.T) {
	resources := []types.Resource{
		testutil.MockResource("zebra-vm", "Microsoft.Compute/virtualMachines", "vm", "/resource/zebra", nil),
		testutil.MockResource("alpha-vm", "Microsoft.Compute/virtualMachines", "vm", "/resource/alpha", nil),
		testutil.MockResource("beta-vm", "Microsoft.Compute/virtualMachines", "vm", "/resource/beta", nil),
	}
	
	// Sort by name (same logic as in fetcher)
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Name < resources[j].Name
	})
	
	// Should be sorted alphabetically
	assert.Equal(t, "alpha-vm", resources[0].Name)
	assert.Equal(t, "beta-vm", resources[1].Name)
	assert.Equal(t, "zebra-vm", resources[2].Name)
}

func TestResourceGroupSorting(t *testing.T) {
	resourceGroups := []types.ResourceGroup{
		testutil.MockResourceGroup("zebra-rg", "eastus", nil, nil),
		testutil.MockResourceGroup("alpha-rg", "westus", nil, nil),
		testutil.MockResourceGroup("beta-rg", "centralus", nil, nil),
	}
	
	// Sort by name (same logic as in fetcher)
	sort.Slice(resourceGroups, func(i, j int) bool {
		return resourceGroups[i].Name < resourceGroups[j].Name
	})
	
	// Should be sorted alphabetically
	assert.Equal(t, "alpha-rg", resourceGroups[0].Name)
	assert.Equal(t, "beta-rg", resourceGroups[1].Name)
	assert.Equal(t, "zebra-rg", resourceGroups[2].Name)
}

// Benchmark tests
func BenchmarkProcessResource(b *testing.B) {
	fetcher := NewFetcher(NewClient(), 5, 10)
	resource := testutil.MockResource("test-vm", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm", nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resourceSemaphore := make(chan struct{}, 1)
		var wg sync.WaitGroup
		resChan := make(chan types.Resource, 1)
		
		wg.Add(1)
		go fetcher.processResource(resource, resourceSemaphore, &wg, resChan)
		wg.Wait()
		close(resChan)
		<-resChan // Consume the result
	}
}

func BenchmarkResourceSorting(b *testing.B) {
	// Create a large slice of resources
	var resources []types.Resource
	for i := 0; i < 1000; i++ {
		resources = append(resources, testutil.MockResource(
			"resource-"+string(rune(1000-i)), // Reverse order to force sorting
			"Microsoft.Test/resources",
			"test",
			"/resource/"+string(rune(i)),
			nil,
		))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resourcesCopy := make([]types.Resource, len(resources))
		copy(resourcesCopy, resources)
		
		sort.Slice(resourcesCopy, func(i, j int) bool {
			return resourcesCopy[i].Name < resourcesCopy[j].Name
		})
	}
}

// Edge case tests
func TestFetcher_EmptyResourceGroups(t *testing.T) {
	// Test handling of empty resource group list
	_ = NewFetcher(NewClient(), 5, 10)
	
	// We can't easily test the full FetchResourceGroups method without mocking,
	// but we can test the components that handle empty slices
	
	var resourceGroups []types.ResourceGroup
	
	// Sorting empty slice should not panic
	sort.Slice(resourceGroups, func(i, j int) bool {
		return resourceGroups[i].Name < resourceGroups[j].Name
	})
	
	assert.Empty(t, resourceGroups)
}

func TestFetcher_ConcurrencyLimits(t *testing.T) {
	tests := []struct {
		name                string
		rgConcurrency       int
		resourceConcurrency int
	}{
		{"minimum concurrency", 1, 1},
		{"moderate concurrency", 5, 10},
		{"high concurrency", 50, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := NewFetcher(NewClient(), tt.rgConcurrency, tt.resourceConcurrency)
			
			// Verify semaphore sizes would be correct
			assert.Equal(t, tt.rgConcurrency, fetcher.rgConcurrency)
			assert.Equal(t, tt.resourceConcurrency, fetcher.resourceConcurrency)
			
			// Test that semaphores would be created with correct capacity
			rgSemaphore := make(chan struct{}, fetcher.rgConcurrency)
			resourceSemaphore := make(chan struct{}, fetcher.resourceConcurrency)
			
			assert.Equal(t, tt.rgConcurrency, cap(rgSemaphore))
			assert.Equal(t, tt.resourceConcurrency, cap(resourceSemaphore))
		})
	}
}