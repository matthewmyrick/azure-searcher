package search

import (
	"sort"
	"strings"
	"unicode"

	"azure-searcher/src/types"
)

// Match represents a fuzzy search match with its score
type Match struct {
	Item  types.ResourceGroup
	Score int
}

// FlatItem represents a flattened item for fzf-style search
type FlatItem struct {
	Type           string // "resource_group" or "resource"
	ResourceGroup  types.ResourceGroup
	Resource       *types.Resource // nil for resource groups
	DisplayPath    string          // e.g., "my-rg" or "my-rg/my-vm"
	SearchableText string          // text to search against
}

// FlatMatch represents a fuzzy match for flattened items
type FlatMatch struct {
	Item  FlatItem
	Score int
}

// FuzzyMatcher provides fuzzy search functionality
type FuzzyMatcher struct{}

// NewFuzzyMatcher creates a new fuzzy matcher
func NewFuzzyMatcher() *FuzzyMatcher {
	return &FuzzyMatcher{}
}

// SearchResourceGroups performs fuzzy search on resource groups
func (fm *FuzzyMatcher) SearchResourceGroups(query string, resourceGroups []types.ResourceGroup) []types.ResourceGroup {
	if query == "" {
		return resourceGroups
	}

	query = strings.ToLower(strings.TrimSpace(query))
	queryTerms := strings.Fields(query)
	
	if len(queryTerms) == 0 {
		return resourceGroups
	}

	var matches []Match
	
	for _, rg := range resourceGroups {
		score := fm.calculateScore(queryTerms, rg)
		if score > 0 {
			matches = append(matches, Match{
				Item:  rg,
				Score: score,
			})
		}
	}

	// Sort by score (highest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// Extract the sorted resource groups
	result := make([]types.ResourceGroup, len(matches))
	for i, match := range matches {
		result[i] = match.Item
	}

	return result
}

// calculateScore calculates a fuzzy match score for a resource group
func (fm *FuzzyMatcher) calculateScore(queryTerms []string, rg types.ResourceGroup) int {
	rgName := strings.ToLower(rg.Name)
	
	totalScore := 0
	matchedTerms := 0

	for _, term := range queryTerms {
		termScore := fm.scoreTermMatch(term, rgName)
		if termScore > 0 {
			totalScore += termScore
			matchedTerms++
		}
	}

	// Must match all terms to be considered a match
	if matchedTerms != len(queryTerms) {
		return 0
	}

	// Bonus for matching all terms
	totalScore += matchedTerms * 10

	// Additional scoring for resources within the group
	resourceScore := fm.scoreResourceMatches(queryTerms, rg.Resources)
	totalScore += resourceScore

	return totalScore
}

// scoreTermMatch scores how well a single term matches the resource group name
func (fm *FuzzyMatcher) scoreTermMatch(term, target string) int {
	// Exact match gets highest score
	if strings.Contains(target, term) {
		return 100
	}

	// Fuzzy character-by-character matching
	return fm.fuzzyMatch(term, target)
}

// fuzzyMatch performs character-by-character fuzzy matching
func (fm *FuzzyMatcher) fuzzyMatch(pattern, text string) int {
	if len(pattern) == 0 {
		return 0
	}

	patternIdx := 0
	score := 0
	consecutiveMatches := 0
	
	for i, char := range text {
		if patternIdx < len(pattern) && unicode.ToLower(char) == unicode.ToLower(rune(pattern[patternIdx])) {
			patternIdx++
			consecutiveMatches++
			
			// Bonus for consecutive matches
			score += 1 + consecutiveMatches
			
			// Bonus for matches at word boundaries
			if i == 0 || !unicode.IsLetter(rune(text[i-1])) {
				score += 5
			}
		} else {
			consecutiveMatches = 0
		}
	}

	// Must match all characters in pattern
	if patternIdx == len(pattern) {
		// Bonus for shorter strings (more precise matches)
		lengthBonus := max(0, 50-len(text))
		return score + lengthBonus
	}

	return 0
}

// scoreResourceMatches adds scoring based on resources within the group
func (fm *FuzzyMatcher) scoreResourceMatches(queryTerms []string, resources []types.Resource) int {
	score := 0
	
	for _, resource := range resources {
		resourceName := strings.ToLower(resource.Name)
		resourceType := strings.ToLower(resource.Type)
		
		for _, term := range queryTerms {
			// Check resource name
			if strings.Contains(resourceName, term) {
				score += 20
			} else if fm.fuzzyMatch(term, resourceName) > 0 {
				score += 10
			}
			
			// Check resource type (lower weight)
			if strings.Contains(resourceType, term) {
				score += 5
			}
		}
	}
	
	return score
}

// SearchWithinResourceGroup searches for resources within a specific resource group
func (fm *FuzzyMatcher) SearchWithinResourceGroup(query string, rg *types.ResourceGroup) []types.Resource {
	if query == "" {
		return rg.Resources
	}

	query = strings.ToLower(strings.TrimSpace(query))
	queryTerms := strings.Fields(query)
	
	if len(queryTerms) == 0 {
		return rg.Resources
	}

	type resourceMatch struct {
		Resource types.Resource
		Score    int
	}

	var matches []resourceMatch
	
	for _, resource := range rg.Resources {
		score := fm.scoreResourceMatch(queryTerms, resource)
		if score > 0 {
			matches = append(matches, resourceMatch{
				Resource: resource,
				Score:    score,
			})
		}
	}

	// Sort by score
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// Extract sorted resources
	result := make([]types.Resource, len(matches))
	for i, match := range matches {
		result[i] = match.Resource
	}

	return result
}

// scoreResourceMatch scores how well query terms match a single resource
func (fm *FuzzyMatcher) scoreResourceMatch(queryTerms []string, resource types.Resource) int {
	resourceName := strings.ToLower(resource.Name)
	resourceType := strings.ToLower(resource.Type)
	
	totalScore := 0
	matchedTerms := 0

	for _, term := range queryTerms {
		termScore := 0
		
		// Check name (higher weight)
		if strings.Contains(resourceName, term) {
			termScore += 100
		} else if fuzzyScore := fm.fuzzyMatch(term, resourceName); fuzzyScore > 0 {
			termScore += fuzzyScore
		}
		
		// Check type (lower weight)
		if strings.Contains(resourceType, term) {
			termScore += 50
		} else if fuzzyScore := fm.fuzzyMatch(term, resourceType); fuzzyScore > 0 {
			termScore += fuzzyScore / 2
		}
		
		if termScore > 0 {
			totalScore += termScore
			matchedTerms++
		}
	}

	// Must match all terms
	if matchedTerms != len(queryTerms) {
		return 0
	}

	return totalScore
}

// SearchResourceGroupsTwoPart performs two-part search: resourcegroup [space] resource
func (fm *FuzzyMatcher) SearchResourceGroupsTwoPart(query string, resourceGroups []types.ResourceGroup) []types.ResourceGroup {
	if query == "" {
		return resourceGroups
	}

	query = strings.TrimSpace(query)
	
	// Split query into resource group part and resource part
	parts := strings.SplitN(query, " ", 2)
	rgQuery := strings.ToLower(parts[0])
	
	var resourceQuery string
	if len(parts) > 1 {
		resourceQuery = strings.ToLower(parts[1])
	}

	var filteredGroups []types.ResourceGroup
	
	// First, filter resource groups by the first part of the query
	for _, rg := range resourceGroups {
		rgName := strings.ToLower(rg.Name)
		
		// Check if resource group matches the first part
		rgMatches := strings.Contains(rgName, rgQuery) || fm.fuzzyMatch(rgQuery, rgName) > 0
		
		if rgMatches {
			filteredRG := rg
			
			// If there's a second part, filter resources within this group
			if resourceQuery != "" {
				var matchingResources []types.Resource
				for _, resource := range rg.Resources {
					resourceText := strings.ToLower(resource.Name + " " + resource.Type)
					if strings.Contains(resourceText, resourceQuery) || fm.fuzzyMatch(resourceQuery, resourceText) > 0 {
						matchingResources = append(matchingResources, resource)
					}
				}
				filteredRG.Resources = matchingResources
				
				// Only include the resource group if it has matching resources
				if len(matchingResources) > 0 {
					filteredRG.Expanded = true
					filteredGroups = append(filteredGroups, filteredRG)
				}
			} else {
				// No resource filter, keep all resources but don't auto-expand
				filteredRG.Resources = rg.Resources
				filteredGroups = append(filteredGroups, filteredRG)
			}
		}
	}
	
	return filteredGroups
}

// SearchResourceGroupsExact performs exact string matching: resourcegroup [space] resource
func (fm *FuzzyMatcher) SearchResourceGroupsExact(query string, resourceGroups []types.ResourceGroup) []types.ResourceGroup {
	if query == "" {
		return resourceGroups
	}

	query = strings.TrimSpace(query)
	
	// Split query into resource group part and resource part
	parts := strings.SplitN(query, " ", 2)
	rgQuery := strings.ToLower(parts[0])
	
	var resourceQuery string
	if len(parts) > 1 {
		resourceQuery = strings.ToLower(parts[1])
	}

	var filteredGroups []types.ResourceGroup
	
	// First, filter resource groups by exact string match
	for _, rg := range resourceGroups {
		rgName := strings.ToLower(rg.Name)
		
		// Check if resource group contains the exact string
		rgMatches := strings.Contains(rgName, rgQuery)
		
		if rgMatches {
			filteredRG := rg
			
			// If there's a second part, filter resources within this group by exact match
			if resourceQuery != "" {
				var matchingResources []types.Resource
				for _, resource := range rg.Resources {
					resourceName := strings.ToLower(resource.Name)
					resourceType := strings.ToLower(resource.Type)
					// Check for exact string match in either name or type
					if strings.Contains(resourceName, resourceQuery) || strings.Contains(resourceType, resourceQuery) {
						matchingResources = append(matchingResources, resource)
					}
				}
				filteredRG.Resources = matchingResources
				
				// Only include the resource group if it has matching resources
				if len(matchingResources) > 0 {
					filteredRG.Expanded = true
					filteredGroups = append(filteredGroups, filteredRG)
				}
			} else {
				// No resource filter, keep all resources but don't auto-expand
				filteredRG.Resources = rg.Resources
				filteredGroups = append(filteredGroups, filteredRG)
			}
		}
	}
	
	return filteredGroups
}

// flattenItems creates a flat list of all searchable items
func (fm *FuzzyMatcher) flattenItems(resourceGroups []types.ResourceGroup) []FlatItem {
	var items []FlatItem
	
	for _, rg := range resourceGroups {
		// Add resource group
		rgItem := FlatItem{
			Type:           "resource_group",
			ResourceGroup:  rg,
			Resource:       nil,
			DisplayPath:    rg.Name,
			SearchableText: strings.ToLower(rg.Name),
		}
		items = append(items, rgItem)
		
		// Add all resources in the group
		for _, resource := range rg.Resources {
			resItem := FlatItem{
				Type:           "resource",
				ResourceGroup:  rg,
				Resource:       &resource,
				DisplayPath:    rg.Name + "/" + resource.Name,
				SearchableText: strings.ToLower(rg.Name + " " + resource.Name + " " + resource.Type),
			}
			items = append(items, resItem)
		}
	}
	
	return items
}

// scoreFlatItem scores how well query terms match a flattened item
func (fm *FuzzyMatcher) scoreFlatItem(queryTerms []string, item FlatItem) int {
	totalScore := 0
	matchedTerms := 0

	for _, term := range queryTerms {
		termScore := 0
		
		// Check exact matches in searchable text
		if strings.Contains(item.SearchableText, term) {
			termScore = 100
		} else {
			// Fuzzy match against searchable text
			termScore = fm.fuzzyMatch(term, item.SearchableText)
		}
		
		// Bonus scoring based on item type
		if termScore > 0 {
			if item.Type == "resource_group" {
				// Bonus for resource group matches
				termScore += 20
			} else {
				// Different scoring for resources
				resourceName := strings.ToLower(item.Resource.Name)
				resourceType := strings.ToLower(item.Resource.Type)
				
				// Extra bonus if term matches resource name specifically
				if strings.Contains(resourceName, term) {
					termScore += 30
				}
				
				// Extra bonus if term matches resource type
				if strings.Contains(resourceType, term) {
					termScore += 15
				}
			}
			
			totalScore += termScore
			matchedTerms++
		}
	}

	// Must match all terms to be considered a match
	if matchedTerms != len(queryTerms) {
		return 0
	}

	// Bonus for matching all terms
	totalScore += matchedTerms * 10

	return totalScore
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}