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
	}

	return "Unknown state"
}

// viewSubscriptions renders the subscription selection view
func (m *Model) viewSubscriptions() string {
	s := TitleStyle.Render("Select Azure Subscription") + "\n\n"

	for i, sub := range m.Subscriptions {
		cursor := " "
		if m.Cursor == i {
			cursor = ">"
			s += SelectedStyle.Render(fmt.Sprintf("%s %s", cursor, sub.Name))
		} else {
			s += fmt.Sprintf("%s %s", cursor, sub.Name)
		}
		s += "\n"
	}

	s += "\nPress Enter to select, Ctrl+Q to quit"
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
	header += searchView + "\n\n"

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
				
				resLine := fmt.Sprintf("  %s 📄 %s (%s)", resCursor, res.Name, res.Type)
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
		footer = "\nEnter: expand/collapse | /: exact search | \\: fuzzy search | Ctrl+E: exit search | Ctrl+R: refresh | PgUp/PgDn: scroll | Esc: back | Ctrl+Q: quit"
	} else {
		footer = fmt.Sprintf("\nShowing %d-%d of %d items | Enter: expand/collapse | /: exact | \\: fuzzy | Ctrl+E: exit search | Ctrl+R: refresh | PgUp/PgDn: scroll | Esc: back | Ctrl+Q: quit", 
			startIdx+1, endIdx, len(items))
	}
	
	return header + content + footer
}