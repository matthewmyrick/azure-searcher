package azure

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"sync"
	"sync/atomic"

	"azure-searcher/src/types"
)

// Client handles Azure CLI operations
type Client struct{}

// NewClient creates a new Azure CLI client
func NewClient() *Client {
	return &Client{}
}

// CheckCLI verifies that Azure CLI is installed
func (c *Client) CheckCLI() error {
	_, err := exec.LookPath("az")
	if err != nil {
		return fmt.Errorf("Azure CLI not found. Please install it from: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli")
	}
	return nil
}

// CheckLogin verifies that the user is logged into Azure
func (c *Client) CheckLogin() error {
	cmd := exec.Command("az", "account", "show")
	err := cmd.Run()
	if err != nil {
		fmt.Println("You are not logged into Azure. Running 'az login'...")
		loginCmd := exec.Command("az", "login")
		loginCmd.Stdout = os.Stdout
		loginCmd.Stderr = os.Stderr
		return loginCmd.Run()
	}
	return nil
}

// GetSubscriptions retrieves all available Azure subscriptions
func (c *Client) GetSubscriptions() ([]types.Subscription, error) {
	cmd := exec.Command("az", "account", "list", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	var subs []types.Subscription
	err = json.Unmarshal(output, &subs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse subscriptions: %w", err)
	}

	return subs, nil
}

// GetResourceGroupsList retrieves basic resource group information
func (c *Client) GetResourceGroupsList(subscriptionID string) ([]types.ResourceGroup, error) {
	cmd := exec.Command("az", "group", "list", "--subscription", subscriptionID, "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get resource groups: %w", err)
	}

	var rgs []types.ResourceGroup
	err = json.Unmarshal(output, &rgs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resource groups: %w", err)
	}

	return rgs, nil
}

// GetResourcesInGroup retrieves all resources in a specific resource group
func (c *Client) GetResourcesInGroup(subscriptionID, resourceGroupName string) ([]types.Resource, error) {
	cmd := exec.Command("az", "resource", "list", "--resource-group", resourceGroupName, "--subscription", subscriptionID, "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get resources for group %s: %w", resourceGroupName, err)
	}

	var resources []types.Resource
	err = json.Unmarshal(output, &resources)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resources for group %s: %w", resourceGroupName, err)
	}

	// Add Azure portal URLs to resources
	for i := range resources {
		resources[i].AzureURL = generateAzurePortalURL(subscriptionID, resources[i].ID)
	}

	return resources, nil
}

// generateAzurePortalURL creates a direct link to the Azure portal for a resource
func generateAzurePortalURL(subscriptionID, resourceID string) string {
	if resourceID == "" {
		return ""
	}
	return fmt.Sprintf("https://portal.azure.com/#resource%s", resourceID)
}