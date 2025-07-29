package types

import "time"

// Subscription represents an Azure subscription
type Subscription struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Resource represents an Azure resource
type Resource struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Kind     string            `json:"kind"`
	ID       string            `json:"id"`
	Tags     map[string]string `json:"tags"`
	AzureURL string            `json:"azure_url"`
}

// ResourceGroup represents an Azure resource group
type ResourceGroup struct {
	Name      string            `json:"name"`
	Location  string            `json:"location"`
	Tags      map[string]string `json:"tags"`
	Resources []Resource        `json:"resources"`
	Expanded  bool              `json:"-"`
}

// CachedResourceGroup represents a cached resource group with timestamp
type CachedResourceGroup struct {
	Name      string            `json:"name"`
	Location  string            `json:"location"`
	Tags      map[string]string `json:"tags"`
	Resources []Resource        `json:"resources"`
	Expanded  bool              `json:"-"`
	CachedAt  time.Time         `json:"cached_at"`
}

// SubscriptionCache represents cached data for a subscription
type SubscriptionCache struct {
	SubscriptionID   string                         `json:"subscription_id"`
	SubscriptionName string                         `json:"subscription_name"`
	ResourceGroups   map[string]CachedResourceGroup `json:"resource_groups"`
	LastUpdated      time.Time                      `json:"last_updated"`
}

// CacheData represents the entire cache file structure
type CacheData struct {
	Subscriptions map[string]SubscriptionCache `json:"subscriptions"`
	Version       string                       `json:"version"`
}

// ProgressUpdate represents progress information during resource fetching
type ProgressUpdate struct {
	Total          int
	Processed      int
	ResourceGroups []ResourceGroup
	Completed      bool
	Error          error
}