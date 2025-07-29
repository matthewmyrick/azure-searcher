package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the main application view
func (m *Model) View() string {
	if m.Err != nil {
		return fmt.Sprintf("Error: %v\nPress any key to exit.", m.Err)
	}

	switch m.State {
	case "subscriptions":
		return m.viewSubscriptions()
	case "loading":
		return m.viewLoading()
	case "resources":
		return m.viewResources()
	case "config":
		return m.viewConfig()
	}

	return "Unknown state"
}

// getResourceIcon returns an appropriate icon based on the resource type
func getResourceIcon(resourceType string) string {
	// Convert to lowercase and check for common Azure resource types
	lowerType := strings.ToLower(resourceType)
	
	// Virtual Machines and Compute
	if strings.Contains(lowerType, "virtualmachine") || strings.Contains(lowerType, "microsoft.compute/virtualmachines") {
		return "💻"
	}
	if strings.Contains(lowerType, "microsoft.compute/disks") {
		return "💾"
	}
	if strings.Contains(lowerType, "microsoft.compute/snapshots") {
		return "📸"
	}
	if strings.Contains(lowerType, "microsoft.compute/images") {
		return "🖼️"
	}
	if strings.Contains(lowerType, "microsoft.compute/availabilitysets") {
		return "🔗"
	}
	
	// Storage
	if strings.Contains(lowerType, "microsoft.storage/storageaccounts") {
		return "🗄️"
	}
	if strings.Contains(lowerType, "microsoft.storage") {
		return "📦"
	}
	
	// Networking
	if strings.Contains(lowerType, "microsoft.network/virtualnetworks") {
		return "🌐"
	}
	if strings.Contains(lowerType, "microsoft.network/networkinterfaces") {
		return "🔌"
	}
	if strings.Contains(lowerType, "microsoft.network/networksecuritygroups") {
		return "🔒"
	}
	if strings.Contains(lowerType, "microsoft.network/publicipaddresses") {
		return "📡"
	}
	if strings.Contains(lowerType, "microsoft.network/loadbalancers") {
		return "⚖️"
	}
	if strings.Contains(lowerType, "microsoft.network/applicationgateways") {
		return "🚪"
	}
	if strings.Contains(lowerType, "microsoft.network/vpngateways") || strings.Contains(lowerType, "microsoft.network/virtualnetworkgateways") {
		return "🔐"
	}
	if strings.Contains(lowerType, "microsoft.network/dnszones") {
		return "🌍"
	}
	if strings.Contains(lowerType, "microsoft.network/privatednszones") {
		return "🔏"
	}
	if strings.Contains(lowerType, "microsoft.network/trafficmanagerprofiles") {
		return "🚦"
	}
	if strings.Contains(lowerType, "microsoft.network/frontdoors") {
		return "🚪"
	}
	if strings.Contains(lowerType, "microsoft.network/firewalls") || strings.Contains(lowerType, "microsoft.network/azurefirewalls") {
		return "🔥"
	}
	if strings.Contains(lowerType, "microsoft.network/privateendpoints") {
		return "🔗"
	}
	
	// Databases
	if strings.Contains(lowerType, "microsoft.sql/servers") {
		return "🗃️"
	}
	if strings.Contains(lowerType, "microsoft.sql/databases") {
		return "🗄️"
	}
	if strings.Contains(lowerType, "microsoft.dbformysql") {
		return "🐬"
	}
	if strings.Contains(lowerType, "microsoft.dbforpostgresql") {
		return "🐘"
	}
	if strings.Contains(lowerType, "microsoft.documentdb") || strings.Contains(lowerType, "cosmosdb") {
		return "🌌"
	}
	if strings.Contains(lowerType, "microsoft.cache/redis") {
		return "⚡"
	}
	
	// Web and App Services
	if strings.Contains(lowerType, "microsoft.web/sites") {
		return "🌐"
	}
	if strings.Contains(lowerType, "microsoft.web/serverfarms") {
		return "🏗️"
	}
	if strings.Contains(lowerType, "microsoft.web/certificates") {
		return "📜"
	}
	
	// Containers
	if strings.Contains(lowerType, "microsoft.containerregistry/registries") {
		return "🐳"
	}
	if strings.Contains(lowerType, "microsoft.containerservice/managedclusters") {
		return "☸️"
	}
	if strings.Contains(lowerType, "microsoft.containerinstance") {
		return "📦"
	}
	
	// Security and Identity
	if strings.Contains(lowerType, "microsoft.keyvault/vaults") {
		return "🔐"
	}
	if strings.Contains(lowerType, "microsoft.managedidentity") {
		return "🆔"
	}
	
	// Analytics and AI
	if strings.Contains(lowerType, "microsoft.databricks") {
		return "⚡"
	}
	if strings.Contains(lowerType, "microsoft.machinelearning") || strings.Contains(lowerType, "microsoft.machinelearningservices") {
		return "🤖"
	}
	if strings.Contains(lowerType, "microsoft.cognitiveservices") {
		return "🧠"
	}
	if strings.Contains(lowerType, "microsoft.search/searchservices") {
		return "🔍"
	}
	if strings.Contains(lowerType, "microsoft.datafactory") {
		return "🏭"
	}
	if strings.Contains(lowerType, "microsoft.synapse") {
		return "⚡"
	}
	if strings.Contains(lowerType, "microsoft.streamanalytics") {
		return "📊"
	}
	if strings.Contains(lowerType, "microsoft.eventhub") {
		return "📨"
	}
	
	// Monitoring and Management
	if strings.Contains(lowerType, "microsoft.insights") {
		return "📊"
	}
	if strings.Contains(lowerType, "microsoft.operationalinsights") || strings.Contains(lowerType, "microsoft.operationsmanagement") {
		return "📈"
	}
	if strings.Contains(lowerType, "microsoft.automation") {
		return "🤖"
	}
	if strings.Contains(lowerType, "microsoft.logic") {
		return "⚙️"
	}
	
	// Service Bus and Messaging
	if strings.Contains(lowerType, "microsoft.servicebus") {
		return "🚌"
	}
	if strings.Contains(lowerType, "microsoft.eventgrid") {
		return "📬"
	}
	
	// IoT
	if strings.Contains(lowerType, "microsoft.devices/iothubs") || strings.Contains(lowerType, "microsoft.iothub") {
		return "📡"
	}
	if strings.Contains(lowerType, "microsoft.iotcentral") {
		return "🏭"
	}
	
	// Media Services
	if strings.Contains(lowerType, "microsoft.media") {
		return "🎬"
	}
	
	// CDN
	if strings.Contains(lowerType, "microsoft.cdn") {
		return "🌍"
	}
	
	// API Management
	if strings.Contains(lowerType, "microsoft.apimanagement") {
		return "🔗"
	}
	
	// Default icon for unknown types
	return "📄"
}

// viewSubscriptions renders the subscription selection view
func (m *Model) viewSubscriptions() string {
	s := TitleStyle.Render("Select Azure Subscription") + "\n\n"

	// Status styling
	cacheDisabledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))  // Light red
	autoRefreshOffStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D"))  // Yellow
	autoRefreshOnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#74B9FF"))   // Blue

	for i, sub := range m.Subscriptions {
		cursor := " "
		
		// Create status text for this specific subscription
		cacheEnabled := m.CacheManager.GetSubscriptionCacheEnabled(sub.ID)
		autoRefreshEnabled := m.CacheManager.GetSubscriptionAutoRefreshEnabled(sub.ID)
		refreshInterval := m.CacheManager.GetSubscriptionRefreshInterval(sub.ID)
		
		var statusText string
		var statusStyle lipgloss.Style
		
		if !cacheEnabled {
			statusText = "(cache: disabled)"
			statusStyle = cacheDisabledStyle
		} else if !autoRefreshEnabled {
			statusText = "(cache: enabled, auto refresh: off)"
			statusStyle = autoRefreshOffStyle
		} else {
			if refreshInterval == "" {
				statusText = "(cache: enabled, auto refresh: 1 dy)" // Default
			} else {
				statusText = fmt.Sprintf("(cache: enabled, auto refresh: %s)", refreshInterval)
			}
			statusStyle = autoRefreshOnStyle
		}
		
		subscriptionLine := fmt.Sprintf("%s %s %s", cursor, sub.Name, statusStyle.Render(statusText))
		
		if m.Cursor == i {
			cursor = ">"
			subscriptionLine = fmt.Sprintf("%s %s %s", cursor, sub.Name, statusStyle.Render(statusText))
			s += SelectedStyle.Render(subscriptionLine)
		} else {
			s += subscriptionLine
		}
		s += "\n"
	}

	s += "\nPress Enter to select, C to configure, Ctrl+Q to quit"
	return s
}

// viewLoading renders the loading view with progress
func (m *Model) viewLoading() string {
	if m.TotalRGs > 0 {
		percentage := float64(m.ProcessedRGs) / float64(m.TotalRGs)
		progressBar := m.Progress.ViewAs(percentage)
		return fmt.Sprintf(
			"\n%s Loading resource groups for %s...\n\n%s\n\nProgress: %d/%d resource groups processed (%.0f%%)\n\nFetching data with %d concurrent resource groups and %d concurrent resources per group.",
			m.Spinner.View(),
			m.SelectedSub.Name,
			progressBar,
			m.ProcessedRGs,
			m.TotalRGs,
			percentage*100,
			m.Config.RGConcurrency,
			m.Config.ResourceConcurrency,
		)
	} else {
		return fmt.Sprintf(
			"\n%s Loading resource groups for %s...\n\nInitializing...\n\nFetching data with %d concurrent resource groups and %d concurrent resources per group.",
			m.Spinner.View(),
			m.SelectedSub.Name,
			m.Config.RGConcurrency,
			m.Config.ResourceConcurrency,
		)
	}
}

// viewResources renders the resource browsing view with scrolling
func (m *Model) viewResources() string {
	// If showing resource types help, show only that
	if m.ShowResourceTypes {
		helpText := getResourceTypeAliases()
		helpBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2).
			MaxWidth(80).
			MaxHeight(m.ViewportHeight - 4).
			Render(helpText)
		
		centered := lipgloss.Place(
			m.ViewportWidth, m.ViewportHeight,
			lipgloss.Center, lipgloss.Center,
			helpBox,
		)
		return centered
	}

	cacheStatus := "📡 Live data"
	if m.LastLoadFromCache {
		cacheStatus = "⚡ Cached data"
	}

	// Header (always visible)
	header := TitleStyle.Render(fmt.Sprintf("Resources in %s", m.SelectedSub.Name)) + "\n"
	header += CacheStatusStyle.Render(cacheStatus) + "\n\n"
	
	// Search mode and documentation
	searchModeText := fmt.Sprintf("Search Mode: %s | Format: <resourcegroup> <resource> (e.g., 'prod sql' or just 'prod')", strings.Title(m.SearchMode))
	header += CacheStatusStyle.Render(searchModeText) + "\n\n"

	// Search input (always visible)
	searchView := m.SearchInput.View()
	if m.SearchInput.Focused() {
		searchView = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(searchView)
	}
	header += searchView + "\n"
	
	// Filter input (collapsible - only show if it has content or is focused)
	if m.FilterInput.Focused() || m.FilterInput.Value() != "" {
		filterView := m.FilterInput.View()
		if m.FilterInput.Focused() {
			filterView = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(filterView)
		}
		header += "\n" + filterView + "\n"
		
		// Show filter help text when focused
		if m.FilterInput.Focused() {
			filterHelp := CacheStatusStyle.Render("Examples: tags=\"Environment:Production\",\"Team:Backend\" resources=vm,storage | Press ? for types | Ctrl+E: exit filter | Esc: clear")
			header += filterHelp + "\n"
		}
	}
	
	header += "\n"

	// Build the full list of items (visible items with their display strings)
	var items []string
	var itemCursors []bool // Track which items should show cursor
	visibleIdx := 0
	
	for _, rg := range m.FilteredGroups {
		icon := "📁"
		if rg.Expanded {
			icon = "📂"
		}

		// Resource group item
		isSelected := visibleIdx == m.Cursor
		itemCursors = append(itemCursors, isSelected)
		
		cursor := " "
		if isSelected {
			cursor = ">"
		}
		
		rgLine := fmt.Sprintf("%s %s %s (%d resources)", cursor, icon, rg.Name, len(rg.Resources))
		if isSelected {
			items = append(items, SelectedStyle.Render(rgLine))
		} else {
			items = append(items, ResourceGroupStyle.Render(rgLine))
		}
		visibleIdx++

		// Resource items (if expanded)
		if rg.Expanded {
			for _, res := range rg.Resources {
				isSelected := visibleIdx == m.Cursor
				itemCursors = append(itemCursors, isSelected)
				
				resCursor := " "
				if isSelected {
					resCursor = ">"
				}
				
				resLine := fmt.Sprintf("  %s %s %s (%s)", resCursor, getResourceIcon(res.Type), res.Name, res.Type)
				if isSelected {
					items = append(items, SelectedStyle.Render(resLine))
				} else {
					items = append(items, ResourceStyle.Render(resLine))
				}
				visibleIdx++
			}
		}
	}

	// Calculate viewport
	viewportSize := m.CalculateResourceViewport()
	if viewportSize < 1 {
		viewportSize = 1
	}

	// Determine visible range
	startIdx := m.ScrollOffset
	endIdx := startIdx + viewportSize
	if endIdx > len(items) {
		endIdx = len(items)
	}
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx >= len(items) {
		startIdx = len(items) - 1
		if startIdx < 0 {
			startIdx = 0
		}
	}

	// Build the scrollable content
	content := ""
	if len(items) == 0 {
		content = CacheStatusStyle.Render("No resource groups found")
	} else {
		for i := startIdx; i < endIdx && i < len(items); i++ {
			content += items[i] + "\n"
		}
		
		// Add scroll indicators
		if startIdx > 0 {
			content = CacheStatusStyle.Render("↑ More items above") + "\n" + content
		}
		if endIdx < len(items) {
			content += CacheStatusStyle.Render("↓ More items below")
		}
	}

	// Footer (always visible)
	var footer string
	if len(items) == 0 {
		footer = "\nEnter: expand/collapse | Ctrl+O: open in browser | Ctrl+F: filter | /: exact search | \\: fuzzy search | Ctrl+E: exit search | Ctrl+R: refresh | Esc: back | Ctrl+Q: quit"
	} else {
		footer = fmt.Sprintf("\nShowing %d-%d of %d items | Enter: expand | Ctrl+O: open | Ctrl+F: filter | /: exact | \\: fuzzy | Ctrl+R: refresh | Esc: back | Ctrl+Q: quit", 
			startIdx+1, endIdx, len(items))
	}
	
	return header + content + footer
}

// viewConfig renders the configuration view
func (m *Model) viewConfig() string {
	if m.ConfigMode == "interval_input" {
		return m.viewIntervalInput()
	}
	
	// Title
	title := TitleStyle.Render(fmt.Sprintf("Configuration - %s", m.ConfiguringSub.Name))
	
	// Get current settings
	cacheEnabled := m.CacheManager.GetSubscriptionCacheEnabled(m.ConfiguringSub.ID)
	autoRefreshEnabled := m.CacheManager.GetSubscriptionAutoRefreshEnabled(m.ConfiguringSub.ID)
	refreshInterval := m.CacheManager.GetSubscriptionRefreshInterval(m.ConfiguringSub.ID)
	
	// Menu options
	var menuItems []string
	
	// Cache toggle option
	cacheStatus := "disabled"
	cacheColor := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")) // Light red
	if cacheEnabled {
		cacheStatus = "enabled"
		cacheColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#51CF66")) // Light green
	}
	
	cacheItem := fmt.Sprintf("Cache: %s", cacheColor.Render(cacheStatus))
	if m.ConfigCursor == 0 {
		cacheItem = SelectedStyle.Render("> " + cacheItem)
	} else {
		cacheItem = "  " + cacheItem
	}
	menuItems = append(menuItems, cacheItem)
	
	// Auto-refresh option (only show if cache is enabled)
	if cacheEnabled {
		autoRefreshStatus := "off"
		autoRefreshColor := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D")) // Yellow
		if autoRefreshEnabled {
			if refreshInterval == "" {
				autoRefreshStatus = "1 dy" // Default
			} else {
				autoRefreshStatus = refreshInterval
			}
			autoRefreshColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#74B9FF")) // Blue
		}
		
		autoRefreshItem := fmt.Sprintf("Auto Refresh: %s", autoRefreshColor.Render(autoRefreshStatus))
		if m.ConfigCursor == 1 {
			autoRefreshItem = SelectedStyle.Render("> " + autoRefreshItem)
		} else {
			autoRefreshItem = "  " + autoRefreshItem
		}
		menuItems = append(menuItems, autoRefreshItem)
	}
	
	// Instructions
	instructions := "\nUse ↑/↓ to navigate, Enter/Space to select, Esc to save and go back"
	
	// Create content
	content := strings.Join(menuItems, "\n") + instructions
	
	// Create a centered floating box
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Width(60).
		Align(lipgloss.Center).
		Render(content)
	
	// Center the box on screen
	screenWidth := m.ViewportWidth
	screenHeight := m.ViewportHeight
	
	centered := lipgloss.Place(
		screenWidth, screenHeight,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, title, "", box),
	)
	
	return centered
}

// viewIntervalInput renders the refresh interval input view
func (m *Model) viewIntervalInput() string {
	title := TitleStyle.Render("Set Auto-Refresh Interval")
	
	instructions := "Enter refresh interval (e.g., '2 hr', '30 min', '1 dy')\nPress Enter to save, Esc to cancel"
	
	// Create input view
	inputView := m.IntervalInput.View()
	if m.IntervalInput.Focused() {
		inputView = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(inputView)
	}
	
	content := fmt.Sprintf("%s\n\n%s", instructions, inputView)
	
	// Create a centered floating box
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Width(70).
		Align(lipgloss.Center).
		Render(content)
	
	// Center the box on screen
	centered := lipgloss.Place(
		m.ViewportWidth, m.ViewportHeight,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, title, "", box),
	)
	
	return centered
}