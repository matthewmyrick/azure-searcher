package config

import "runtime"

const (
	// Default concurrency limits - can be adjusted based on system performance
	DefaultRGConcurrency       = 5  // Max concurrent resource groups
	DefaultResourceConcurrency = 10 // Max concurrent resources per group

	// Cache settings
	DefaultCacheTTLMinutes = 30 // Cache TTL in minutes
	CacheFilename          = "azure-searcher-cache.json"
)

// ConcurrencyConfig holds the concurrency limits for the application
type ConcurrencyConfig struct {
	RGConcurrency       int
	ResourceConcurrency int
}

// GetOptimalConcurrency returns concurrency limits based on system resources
func GetOptimalConcurrency() ConcurrencyConfig {
	rgConcurrency := DefaultRGConcurrency
	resourceConcurrency := DefaultResourceConcurrency

	// Adjust limits based on CPU count for better performance
	cpuCount := runtime.NumCPU()
	if cpuCount <= 2 {
		rgConcurrency = 2
		resourceConcurrency = 5
	} else if cpuCount <= 4 {
		rgConcurrency = 3
		resourceConcurrency = 8
	}

	return ConcurrencyConfig{
		RGConcurrency:       rgConcurrency,
		ResourceConcurrency: resourceConcurrency,
	}
}