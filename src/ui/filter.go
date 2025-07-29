package ui

import (
	"regexp"
	"strings"

	"github.com/matthewmyrick/azure-searcher/src/types"
)

// FilterCriteria represents parsed filter criteria
type FilterCriteria struct {
	Tags         map[string]string
	ResourceTypes []string
}

// parseFilter parses the filter input string
func parseFilter(input string) FilterCriteria {
	criteria := FilterCriteria{
		Tags:          make(map[string]string),
		ResourceTypes: []string{},
	}

	if input == "" {
		return criteria
	}

	// Parse tags
	tagsRegex := regexp.MustCompile(`tags="([^"]+)"`)
	if match := tagsRegex.FindStringSubmatch(input); len(match) > 1 {
		tagPairs := strings.Split(match[1], `","`)
		for _, pair := range tagPairs {
			// Remove quotes from individual tags
			pair = strings.Trim(pair, `"`)
			parts := strings.SplitN(pair, ":", 2)
			if len(parts) == 2 {
				criteria.Tags[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Parse resources
	resourcesRegex := regexp.MustCompile(`resources=([^\s]+)`)
	if match := resourcesRegex.FindStringSubmatch(input); len(match) > 1 {
		resources := strings.Split(match[1], ",")
		for _, res := range resources {
			res = strings.TrimSpace(res)
			if res != "" {
				criteria.ResourceTypes = append(criteria.ResourceTypes, res)
			}
		}
	}

	return criteria
}

// matchesFilter checks if a resource matches the filter criteria
func matchesFilter(resource types.Resource, criteria FilterCriteria) bool {
	// Check resource type filter
	if len(criteria.ResourceTypes) > 0 {
		typeMatches := false
		resourceTypeLower := strings.ToLower(resource.Type)
		
		for _, filterType := range criteria.ResourceTypes {
			filterTypeLower := strings.ToLower(filterType)
			
			// Map common abbreviations to Azure resource types
			switch filterTypeLower {
			case "vm", "virtualmachine":
				if strings.Contains(resourceTypeLower, "microsoft.compute/virtualmachines") {
					typeMatches = true
				}
			case "disk":
				if strings.Contains(resourceTypeLower, "microsoft.compute/disks") {
					typeMatches = true
				}
			case "storage":
				if strings.Contains(resourceTypeLower, "microsoft.storage") {
					typeMatches = true
				}
			case "sql":
				if strings.Contains(resourceTypeLower, "microsoft.sql") {
					typeMatches = true
				}
			case "app-service", "appservice", "webapp":
				if strings.Contains(resourceTypeLower, "microsoft.web/sites") {
					typeMatches = true
				}
			case "keyvault", "kv":
				if strings.Contains(resourceTypeLower, "microsoft.keyvault/vaults") {
					typeMatches = true
				}
			case "aks", "kubernetes":
				if strings.Contains(resourceTypeLower, "microsoft.containerservice/managedclusters") {
					typeMatches = true
				}
			case "acr", "containerregistry":
				if strings.Contains(resourceTypeLower, "microsoft.containerregistry/registries") {
					typeMatches = true
				}
			case "cosmosdb", "cosmos":
				if strings.Contains(resourceTypeLower, "microsoft.documentdb") {
					typeMatches = true
				}
			case "redis":
				if strings.Contains(resourceTypeLower, "microsoft.cache/redis") {
					typeMatches = true
				}
			case "vnet", "virtualnetwork":
				if strings.Contains(resourceTypeLower, "microsoft.network/virtualnetworks") {
					typeMatches = true
				}
			case "nsg":
				if strings.Contains(resourceTypeLower, "microsoft.network/networksecuritygroups") {
					typeMatches = true
				}
			case "lb", "loadbalancer":
				if strings.Contains(resourceTypeLower, "microsoft.network/loadbalancers") {
					typeMatches = true
				}
			case "pip", "publicip":
				if strings.Contains(resourceTypeLower, "microsoft.network/publicipaddresses") {
					typeMatches = true
				}
			case "nic", "networkinterface":
				if strings.Contains(resourceTypeLower, "microsoft.network/networkinterfaces") {
					typeMatches = true
				}
			case "dns":
				if strings.Contains(resourceTypeLower, "microsoft.network/dnszones") {
					typeMatches = true
				}
			case "firewall":
				if strings.Contains(resourceTypeLower, "microsoft.network/azurefirewalls") || 
				   strings.Contains(resourceTypeLower, "microsoft.network/firewalls") {
					typeMatches = true
				}
			case "appgateway", "applicationgateway":
				if strings.Contains(resourceTypeLower, "microsoft.network/applicationgateways") {
					typeMatches = true
				}
			case "eventhub":
				if strings.Contains(resourceTypeLower, "microsoft.eventhub") {
					typeMatches = true
				}
			case "servicebus":
				if strings.Contains(resourceTypeLower, "microsoft.servicebus") {
					typeMatches = true
				}
			case "datafactory":
				if strings.Contains(resourceTypeLower, "microsoft.datafactory") {
					typeMatches = true
				}
			case "databricks":
				if strings.Contains(resourceTypeLower, "microsoft.databricks") {
					typeMatches = true
				}
			case "ml", "machinelearning":
				if strings.Contains(resourceTypeLower, "microsoft.machinelearning") {
					typeMatches = true
				}
			case "cognitiveservices", "ai":
				if strings.Contains(resourceTypeLower, "microsoft.cognitiveservices") {
					typeMatches = true
				}
			case "search":
				if strings.Contains(resourceTypeLower, "microsoft.search/searchservices") {
					typeMatches = true
				}
			case "monitor", "insights":
				if strings.Contains(resourceTypeLower, "microsoft.insights") {
					typeMatches = true
				}
			case "loganalytics":
				if strings.Contains(resourceTypeLower, "microsoft.operationalinsights") {
					typeMatches = true
				}
			case "automation":
				if strings.Contains(resourceTypeLower, "microsoft.automation") {
					typeMatches = true
				}
			case "logic", "logicapp":
				if strings.Contains(resourceTypeLower, "microsoft.logic") {
					typeMatches = true
				}
			case "cdn":
				if strings.Contains(resourceTypeLower, "microsoft.cdn") {
					typeMatches = true
				}
			case "apimanagement", "apim":
				if strings.Contains(resourceTypeLower, "microsoft.apimanagement") {
					typeMatches = true
				}
			case "privateendpoint":
				if strings.Contains(resourceTypeLower, "microsoft.network/privateendpoints") {
					typeMatches = true
				}
			default:
				// Direct match
				if strings.Contains(resourceTypeLower, filterTypeLower) {
					typeMatches = true
				}
			}
			
			if typeMatches {
				break
			}
		}
		
		if !typeMatches {
			return false
		}
	}

	// Check tag filter
	if len(criteria.Tags) > 0 {
		if resource.Tags == nil {
			return false
		}
		
		for key, value := range criteria.Tags {
			if tagValue, ok := resource.Tags[key]; !ok || tagValue != value {
				return false
			}
		}
	}

	return true
}

// getResourceTypeAliases returns a formatted list of available resource type aliases
func getResourceTypeAliases() string {
	return `RESOURCE TYPE FILTERS (Press ESC to close)
══════════════════════════════════════════

COMPUTE                        │ NETWORKING
vm, virtualmachine → VMs       │ vnet → Virtual Networks  
disk → Managed Disks           │ nsg → Security Groups
                               │ lb → Load Balancers
STORAGE & DATA                 │ pip → Public IPs
storage → Storage Accounts     │ nic → Network Interfaces
sql → SQL Server/DB            │ dns → DNS Zones
cosmosdb → Cosmos DB           │ firewall → Firewalls
redis → Redis Cache            │ appgateway → App Gateways
                               │ privateendpoint → Private Endpoints
WEB & APPS                     │
app-service, webapp → Apps     │ ANALYTICS & AI
logic → Logic Apps             │ databricks → Databricks
                               │ ml → Machine Learning
CONTAINERS                     │ ai → Cognitive Services
aks → AKS Clusters             │ search → Search
acr → Container Registry       │ datafactory → Data Factory
                               │
SECURITY                       │ MESSAGING & MONITORING
keyvault, kv → Key Vaults      │ eventhub → Event Hubs
                               │ servicebus → Service Bus
OTHER                          │ insights → App Insights
cdn → CDN Profiles             │ loganalytics → Log Analytics
apim → API Management          │ automation → Automation

Examples: resources=vm,storage,sql  or  resources=microsoft.compute/virtualmachines`
}

// filterResourceGroupsWithCriteria filters resource groups based on both search and filter criteria
func (m *Model) filterResourceGroupsWithCriteria() {
	query := m.SearchInput.Value()
	filterInput := m.FilterInput.Value()
	criteria := parseFilter(filterInput)
	
	// First apply search filter
	var searchFiltered []types.ResourceGroup
	if m.SearchMode == "exact" {
		searchFiltered = m.FuzzyMatcher.SearchResourceGroupsExact(query, m.ResourceGroups)
	} else {
		searchFiltered = m.FuzzyMatcher.SearchResourceGroupsTwoPart(query, m.ResourceGroups)
	}
	
	// Then apply advanced filters
	m.FilteredGroups = []types.ResourceGroup{}
	for _, rg := range searchFiltered {
		filteredResources := []types.Resource{}
		
		// Check if we have filters to apply
		hasFilters := len(criteria.Tags) > 0 || len(criteria.ResourceTypes) > 0
		
		if hasFilters {
			// Filter resources within the group
			for _, res := range rg.Resources {
				if matchesFilter(res, criteria) {
					filteredResources = append(filteredResources, res)
				}
			}
			
			// Only include resource group if it has matching resources
			if len(filteredResources) > 0 {
				rgCopy := rg
				rgCopy.Resources = filteredResources
				m.FilteredGroups = append(m.FilteredGroups, rgCopy)
			}
		} else {
			// No filters, include all resources
			m.FilteredGroups = append(m.FilteredGroups, rg)
		}
	}
}