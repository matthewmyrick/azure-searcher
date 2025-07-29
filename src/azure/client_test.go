package azure

import (
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	"github.com/matthewmyrick/azure-searcher/src/testutil"
	"github.com/matthewmyrick/azure-searcher/src/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	assert.NotNil(t, client)
}

func TestClient_CheckCLI_Success(t *testing.T) {
	client := NewClient()
	
	// This test depends on the system having 'az' command available
	// We'll check if it exists and skip if not
	_, err := exec.LookPath("az")
	if err != nil {
		t.Skip("Azure CLI not available for testing")
	}

	err = client.CheckCLI()
	assert.NoError(t, err)
}

func TestClient_CheckCLI_NotFound(t *testing.T) {
	client := NewClient()
	
	// Mock the PATH environment to not include az command
	originalPath := os.Getenv("PATH")
	testutil.SetEnv(t, "PATH", "/nonexistent/path")
	
	// Restore PATH after test
	defer func() {
		os.Setenv("PATH", originalPath)
	}()

	err := client.CheckCLI()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Azure CLI not found")
	assert.Contains(t, err.Error(), "https://docs.microsoft.com/en-us/cli/azure/install-azure-cli")
}

func TestGenerateAzurePortalURL(t *testing.T) {
	tests := []struct {
		name           string
		subscriptionID string
		resourceID     string
		expected       string
	}{
		{
			name:           "valid resource ID",
			subscriptionID: "sub-123",
			resourceID:     "/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Compute/virtualMachines/vm-test",
			expected:       "https://portal.azure.com/#resource/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Compute/virtualMachines/vm-test",
		},
		{
			name:           "empty resource ID",
			subscriptionID: "sub-123",
			resourceID:     "",
			expected:       "",
		},
		{
			name:           "simple resource ID",
			subscriptionID: "sub-123",
			resourceID:     "/resource/test",
			expected:       "https://portal.azure.com/#resource/resource/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateAzurePortalURL(tt.subscriptionID, tt.resourceID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Note: The following tests would require mocking exec.Command which is complex in Go.
// For a production implementation, you'd want to refactor the Client to use an interface
// for command execution that can be mocked. Here we provide examples of what the tests
// would look like and implement basic structure tests.

func TestClient_GetSubscriptions_StructureValidation(t *testing.T) {
	// This test validates the JSON unmarshaling logic by testing with known data
	testJSON := `[
		{
			"id": "subscription-1",
			"name": "Test Subscription 1"
		},
		{
			"id": "subscription-2", 
			"name": "Test Subscription 2"
		}
	]`

	var subscriptions []types.Subscription
	err := json.Unmarshal([]byte(testJSON), &subscriptions)
	require.NoError(t, err)

	assert.Len(t, subscriptions, 2)
	assert.Equal(t, "subscription-1", subscriptions[0].ID)
	assert.Equal(t, "Test Subscription 1", subscriptions[0].Name)
	assert.Equal(t, "subscription-2", subscriptions[1].ID)
	assert.Equal(t, "Test Subscription 2", subscriptions[1].Name)
}

func TestClient_GetResourceGroupsList_StructureValidation(t *testing.T) {
	// Test the JSON unmarshaling for resource groups
	testJSON := `[
		{
			"name": "test-rg-1",
			"location": "eastus",
			"tags": {
				"environment": "test"
			}
		},
		{
			"name": "test-rg-2",
			"location": "westus", 
			"tags": null
		}
	]`

	var resourceGroups []types.ResourceGroup
	err := json.Unmarshal([]byte(testJSON), &resourceGroups)
	require.NoError(t, err)

	assert.Len(t, resourceGroups, 2)
	assert.Equal(t, "test-rg-1", resourceGroups[0].Name)
	assert.Equal(t, "eastus", resourceGroups[0].Location)
	assert.Equal(t, "test", resourceGroups[0].Tags["environment"])
	
	assert.Equal(t, "test-rg-2", resourceGroups[1].Name)
	assert.Equal(t, "westus", resourceGroups[1].Location)
	assert.Nil(t, resourceGroups[1].Tags)
}

func TestClient_GetResourcesInGroup_StructureValidation(t *testing.T) {
	// Test the JSON unmarshaling and URL generation for resources
	testJSON := `[
		{
			"name": "test-vm",
			"type": "Microsoft.Compute/virtualMachines",
			"kind": "vm",
			"id": "/subscriptions/sub-123/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachines/test-vm",
			"tags": {
				"environment": "production"
			}
		},
		{
			"name": "test-storage",
			"type": "Microsoft.Storage/storageAccounts",
			"kind": "storage",
			"id": "/subscriptions/sub-123/resourceGroups/test-rg/providers/Microsoft.Storage/storageAccounts/test-storage",
			"tags": {}
		}
	]`

	var resources []types.Resource
	err := json.Unmarshal([]byte(testJSON), &resources)
	require.NoError(t, err)

	// Simulate the URL generation that happens in GetResourcesInGroup
	subscriptionID := "sub-123"
	for i := range resources {
		resources[i].AzureURL = generateAzurePortalURL(subscriptionID, resources[i].ID)
	}

	assert.Len(t, resources, 2)
	
	// Verify first resource
	assert.Equal(t, "test-vm", resources[0].Name)
	assert.Equal(t, "Microsoft.Compute/virtualMachines", resources[0].Type)
	assert.Equal(t, "vm", resources[0].Kind) 
	assert.Equal(t, "production", resources[0].Tags["environment"])
	assert.Contains(t, resources[0].AzureURL, "portal.azure.com")
	assert.Contains(t, resources[0].AzureURL, resources[0].ID)

	// Verify second resource
	assert.Equal(t, "test-storage", resources[1].Name)
	assert.Equal(t, "Microsoft.Storage/storageAccounts", resources[1].Type)
	assert.Equal(t, "storage", resources[1].Kind)
	assert.Equal(t, map[string]string{}, resources[1].Tags) // Empty map
	assert.Contains(t, resources[1].AzureURL, "portal.azure.com")
}

func TestClient_ErrorHandling_InvalidJSON(t *testing.T) {
	// Test that invalid JSON returns appropriate errors
	invalidJSON := `invalid json content`

	var subscriptions []types.Subscription
	err := json.Unmarshal([]byte(invalidJSON), &subscriptions)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid character")

	var resourceGroups []types.ResourceGroup
	err = json.Unmarshal([]byte(invalidJSON), &resourceGroups)
	assert.Error(t, err)

	var resources []types.Resource
	err = json.Unmarshal([]byte(invalidJSON), &resources)
	assert.Error(t, err)
}

func TestClient_EmptyResponses(t *testing.T) {
	// Test handling of empty Azure CLI responses
	emptyArrayJSON := `[]`

	// Empty subscriptions
	var subscriptions []types.Subscription
	err := json.Unmarshal([]byte(emptyArrayJSON), &subscriptions)
	require.NoError(t, err)
	assert.Empty(t, subscriptions)

	// Empty resource groups
	var resourceGroups []types.ResourceGroup  
	err = json.Unmarshal([]byte(emptyArrayJSON), &resourceGroups)
	require.NoError(t, err)
	assert.Empty(t, resourceGroups)

	// Empty resources
	var resources []types.Resource
	err = json.Unmarshal([]byte(emptyArrayJSON), &resources)
	require.NoError(t, err)
	assert.Empty(t, resources)
}

func TestClient_ResourceWithoutTags(t *testing.T) {
	// Test resource parsing when tags are missing or null
	testCases := []string{
		// Missing tags field
		`[{"name": "test-vm", "type": "Microsoft.Compute/virtualMachines", "id": "/resource/vm"}]`,
		// Null tags
		`[{"name": "test-vm", "type": "Microsoft.Compute/virtualMachines", "id": "/resource/vm", "tags": null}]`,
		// Empty tags object
		`[{"name": "test-vm", "type": "Microsoft.Compute/virtualMachines", "id": "/resource/vm", "tags": {}}]`,
	}

	for i, testJSON := range testCases {
		t.Run("case_"+string(rune(i)), func(t *testing.T) {
			var resources []types.Resource
			err := json.Unmarshal([]byte(testJSON), &resources)
			require.NoError(t, err)
			require.Len(t, resources, 1)
			
			resource := resources[0]
			assert.Equal(t, "test-vm", resource.Name)
			assert.Equal(t, "Microsoft.Compute/virtualMachines", resource.Type)
			assert.Equal(t, "/resource/vm", resource.ID)
			
			// Tags should be either nil or empty map depending on JSON
			if resource.Tags != nil {
				assert.Empty(t, resource.Tags)
			}
		})
	}
}

// MockableClient interface for testing (would be used in a real refactor)
type MockableClient interface {
	CheckCLI() error
	CheckLogin() error
	GetSubscriptions() ([]types.Subscription, error)
	GetResourceGroupsList(subscriptionID string) ([]types.Resource, error)
	GetResourcesInGroup(subscriptionID, resourceGroupName string) ([]types.Resource, error)
}

// Example of how you could structure tests with a mockable interface
// This demonstrates the testing approach that would be used in production

type MockClient struct {
	subscriptions   []types.Subscription
	resourceGroups  []types.ResourceGroup
	resources       []types.Resource
	shouldFailCLI   bool
	shouldFailLogin bool
	shouldFailAPI   bool
}

func (m *MockClient) CheckCLI() error {
	if m.shouldFailCLI {
		return assert.AnError
	}
	return nil
}

func (m *MockClient) CheckLogin() error {
	if m.shouldFailLogin {
		return assert.AnError
	}
	return nil
}

func (m *MockClient) GetSubscriptions() ([]types.Subscription, error) {
	if m.shouldFailAPI {
		return nil, assert.AnError
	}
	return m.subscriptions, nil
}

func (m *MockClient) GetResourceGroupsList(subscriptionID string) ([]types.ResourceGroup, error) {
	if m.shouldFailAPI {
		return nil, assert.AnError
	}
	return m.resourceGroups, nil
}

func (m *MockClient) GetResourcesInGroup(subscriptionID, resourceGroupName string) ([]types.Resource, error) {
	if m.shouldFailAPI {
		return nil, assert.AnError
	}
	return m.resources, nil
}

func TestMockClient_Functionality(t *testing.T) {
	// Example test showing how mock client would work
	mockClient := &MockClient{
		subscriptions: []types.Subscription{
			testutil.MockSubscription("sub-1", "Test Sub 1"),
			testutil.MockSubscription("sub-2", "Test Sub 2"),
		},
		resourceGroups: []types.ResourceGroup{
			testutil.MockResourceGroup("rg-1", "eastus", nil, nil),
		},
		resources: []types.Resource{
			testutil.MockResource("vm-1", "Microsoft.Compute/virtualMachines", "vm", "/resource/vm-1", nil),
		},
	}

	// Test successful operations
	subs, err := mockClient.GetSubscriptions()
	require.NoError(t, err)
	assert.Len(t, subs, 2)

	rgs, err := mockClient.GetResourceGroupsList("sub-1")
	require.NoError(t, err)
	assert.Len(t, rgs, 1)

	resources, err := mockClient.GetResourcesInGroup("sub-1", "rg-1")
	require.NoError(t, err)
	assert.Len(t, resources, 1)

	// Test error conditions
	mockClient.shouldFailAPI = true
	_, err = mockClient.GetSubscriptions()
	assert.Error(t, err)
}