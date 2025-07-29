package azure

import (
	"sort"
	"sync"
	"sync/atomic"

	"github.com/matthewmyrick/azure-searcher/src/types"
)

// Fetcher handles concurrent resource fetching operations
type Fetcher struct {
	client              *Client
	rgConcurrency       int
	resourceConcurrency int
}

// NewFetcher creates a new resource fetcher with specified concurrency limits
func NewFetcher(client *Client, rgConcurrency, resourceConcurrency int) *Fetcher {
	return &Fetcher{
		client:              client,
		rgConcurrency:       rgConcurrency,
		resourceConcurrency: resourceConcurrency,
	}
}

// FetchResourceGroups retrieves all resource groups with their resources using controlled concurrency
func (f *Fetcher) FetchResourceGroups(subscriptionID string, progressChan chan<- types.ProgressUpdate) ([]types.ResourceGroup, error) {
	// Get list of resource groups first
	rgs, err := f.client.GetResourceGroupsList(subscriptionID)
	if err != nil {
		return nil, err
	}

	// Send initial progress update with total count
	if progressChan != nil {
		progressChan <- types.ProgressUpdate{
			Total:     len(rgs),
			Processed: 0,
		}
	}

	// Create semaphore to limit concurrent resource group processing
	rgSemaphore := make(chan struct{}, f.rgConcurrency)
	var wg sync.WaitGroup
	rgChan := make(chan types.ResourceGroup, len(rgs))
	var processedCount int64

	for _, rg := range rgs {
		wg.Add(1)
		go f.processResourceGroup(subscriptionID, rg.Name, rgSemaphore, &wg, rgChan, &processedCount, len(rgs), progressChan)
	}

	go func() {
		wg.Wait()
		close(rgChan)
	}()

	var result []types.ResourceGroup
	for rg := range rgChan {
		result = append(result, rg)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

// processResourceGroup processes a single resource group with its resources
func (f *Fetcher) processResourceGroup(subscriptionID, rgName string, rgSemaphore chan struct{}, wg *sync.WaitGroup, rgChan chan<- types.ResourceGroup, processedCount *int64, totalRGs int, progressChan chan<- types.ProgressUpdate) {
	defer wg.Done()

	// Acquire semaphore
	rgSemaphore <- struct{}{}
	defer func() { <-rgSemaphore }()

	// Get resources in this resource group
	resources, err := f.processResourcesInGroup(subscriptionID, rgName)
	if err != nil {
		// Send empty resource group on error but still update progress
		rgChan <- types.ResourceGroup{
			Name:      rgName,
			Resources: []types.Resource{},
			Expanded:  false,
		}
	} else {
		rgChan <- types.ResourceGroup{
			Name:      rgName,
			Resources: resources,
			Expanded:  false,
		}
	}

	// Update progress
	processed := atomic.AddInt64(processedCount, 1)
	if progressChan != nil {
		progressChan <- types.ProgressUpdate{
			Total:     totalRGs,
			Processed: int(processed),
		}
	}
}

// processResourcesInGroup fetches and processes all resources in a resource group
func (f *Fetcher) processResourcesInGroup(subscriptionID, rgName string) ([]types.Resource, error) {
	rawResources, err := f.client.GetResourcesInGroup(subscriptionID, rgName)
	if err != nil {
		return []types.Resource{}, err
	}

	// Use limited goroutines to process resources within this group
	resourceSemaphore := make(chan struct{}, f.resourceConcurrency)
	var resourceWg sync.WaitGroup
	resChan := make(chan types.Resource, len(rawResources))

	for _, res := range rawResources {
		resourceWg.Add(1)
		go f.processResource(res, resourceSemaphore, &resourceWg, resChan)
	}

	go func() {
		resourceWg.Wait()
		close(resChan)
	}()

	var resources []types.Resource
	for res := range resChan {
		resources = append(resources, res)
	}

	// Sort resources by name
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Name < resources[j].Name
	})

	return resources, nil
}

// processResource processes a single resource (placeholder for future enhancements)
func (f *Fetcher) processResource(resource types.Resource, resourceSemaphore chan struct{}, wg *sync.WaitGroup, resChan chan<- types.Resource) {
	defer wg.Done()

	// Acquire resource semaphore
	resourceSemaphore <- struct{}{}
	defer func() { <-resourceSemaphore }()

	// Currently just pass through the resource
	// In the future, could fetch additional details here
	resChan <- resource
}