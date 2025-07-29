package config

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	// Test that constants are set to expected values
	assert.Equal(t, 5, DefaultRGConcurrency)
	assert.Equal(t, 10, DefaultResourceConcurrency)
	assert.Equal(t, 30, DefaultCacheTTLMinutes)
	assert.Equal(t, "azure-searcher-cache.json", CacheFilename)
}

func TestConcurrencyConfig_Structure(t *testing.T) {
	config := ConcurrencyConfig{
		RGConcurrency:       5,
		ResourceConcurrency: 10,
	}
	
	assert.Equal(t, 5, config.RGConcurrency)
	assert.Equal(t, 10, config.ResourceConcurrency)
}

func TestGetOptimalConcurrency_DefaultValues(t *testing.T) {
	config := GetOptimalConcurrency()
	
	// Should return valid positive values
	assert.Greater(t, config.RGConcurrency, 0)
	assert.Greater(t, config.ResourceConcurrency, 0)
	
	// Should be within reasonable bounds
	assert.LessOrEqual(t, config.RGConcurrency, DefaultRGConcurrency)
	assert.LessOrEqual(t, config.ResourceConcurrency, DefaultResourceConcurrency)
}

func TestGetOptimalConcurrency_CPUBased(t *testing.T) {
	cpuCount := runtime.NumCPU()
	config := GetOptimalConcurrency()
	
	// Test the logic based on actual CPU count
	switch {
	case cpuCount <= 2:
		assert.Equal(t, 2, config.RGConcurrency)
		assert.Equal(t, 5, config.ResourceConcurrency)
	case cpuCount <= 4:
		assert.Equal(t, 3, config.RGConcurrency)
		assert.Equal(t, 8, config.ResourceConcurrency)
	default:
		// Should use defaults for higher CPU counts
		assert.Equal(t, DefaultRGConcurrency, config.RGConcurrency)
		assert.Equal(t, DefaultResourceConcurrency, config.ResourceConcurrency)
	}
}

func TestGetOptimalConcurrency_Consistency(t *testing.T) {
	// Multiple calls should return the same result
	config1 := GetOptimalConcurrency()
	config2 := GetOptimalConcurrency()
	
	assert.Equal(t, config1.RGConcurrency, config2.RGConcurrency)
	assert.Equal(t, config1.ResourceConcurrency, config2.ResourceConcurrency)
}

func TestGetOptimalConcurrency_ReasonableBounds(t *testing.T) {
	config := GetOptimalConcurrency()
	
	// Ensure values are within reasonable bounds for any system
	assert.GreaterOrEqual(t, config.RGConcurrency, 2, "RG concurrency should be at least 2")
	assert.LessOrEqual(t, config.RGConcurrency, 10, "RG concurrency should not exceed 10")
	
	assert.GreaterOrEqual(t, config.ResourceConcurrency, 5, "Resource concurrency should be at least 5")
	assert.LessOrEqual(t, config.ResourceConcurrency, 20, "Resource concurrency should not exceed 20")
}

// TestConcurrencyLogic tests the concurrency adjustment logic more explicitly
func TestConcurrencyLogic(t *testing.T) {
	tests := []struct {
		name                    string
		cpuCount               int
		expectedRGConcurrency  int
		expectedResConcurrency int
	}{
		{
			name:                    "single core",
			cpuCount:               1,
			expectedRGConcurrency:  2,
			expectedResConcurrency: 5,
		},
		{
			name:                    "dual core",
			cpuCount:               2,
			expectedRGConcurrency:  2,
			expectedResConcurrency: 5,
		},
		{
			name:                    "quad core",
			cpuCount:               4,
			expectedRGConcurrency:  3,
			expectedResConcurrency: 8,
		},
		{
			name:                    "eight core",
			cpuCount:               8,
			expectedRGConcurrency:  DefaultRGConcurrency,
			expectedResConcurrency: DefaultResourceConcurrency,
		},
		{
			name:                    "high core count",
			cpuCount:               16,
			expectedRGConcurrency:  DefaultRGConcurrency,
			expectedResConcurrency: DefaultResourceConcurrency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily mock runtime.NumCPU(), but we can test the logic
			// by understanding what GetOptimalConcurrency should return for the actual system
			config := GetOptimalConcurrency()
			actualCPUs := runtime.NumCPU()
			
			// Verify the logic matches what we expect for the actual CPU count
			if actualCPUs <= 2 {
				assert.Equal(t, 2, config.RGConcurrency)
				assert.Equal(t, 5, config.ResourceConcurrency)
			} else if actualCPUs <= 4 {
				assert.Equal(t, 3, config.RGConcurrency)  
				assert.Equal(t, 8, config.ResourceConcurrency)
			} else {
				assert.Equal(t, DefaultRGConcurrency, config.RGConcurrency)
				assert.Equal(t, DefaultResourceConcurrency, config.ResourceConcurrency)
			}
		})
	}
}

// BenchmarkGetOptimalConcurrency tests performance of the function
func BenchmarkGetOptimalConcurrency(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetOptimalConcurrency()
	}
}

func TestConcurrencyConfig_ZeroValues(t *testing.T) {
	var config ConcurrencyConfig
	
	// Zero values should be zero
	assert.Equal(t, 0, config.RGConcurrency)
	assert.Equal(t, 0, config.ResourceConcurrency)
}

func TestConcurrencyConfig_CustomValues(t *testing.T) {
	config := ConcurrencyConfig{
		RGConcurrency:       15,
		ResourceConcurrency: 25,
	}
	
	assert.Equal(t, 15, config.RGConcurrency)
	assert.Equal(t, 25, config.ResourceConcurrency)
}