package ui

import (
	"fmt"

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

	s += "\nPress Enter to select, q to quit"
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

// viewResources renders the resource browsing view
func (m *Model) viewResources() string {
	cacheStatus := "📡 Live data"
	if m.LastLoadFromCache {
		cacheStatus = "⚡ Cached data"
	}

	s := TitleStyle.Render(fmt.Sprintf("Resources in %s", m.SelectedSub.Name)) + "\n"
	s += CacheStatusStyle.Render(cacheStatus) + "\n\n"

	searchView := m.SearchInput.View()
	if m.SearchInput.Focused() {
		searchView = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(searchView)
	}
	s += searchView + "\n\n"

	visibleIdx := 0
	for _, rg := range m.FilteredGroups {
		cursor := " "
		icon := "📁"
		if rg.Expanded {
			icon = "📂"
		}

		if visibleIdx == m.Cursor {
			cursor = ">"
			s += SelectedStyle.Render(fmt.Sprintf("%s %s %s (%d resources)", cursor, icon, rg.Name, len(rg.Resources)))
		} else {
			s += ResourceGroupStyle.Render(fmt.Sprintf("%s %s %s (%d resources)", cursor, icon, rg.Name, len(rg.Resources)))
		}
		s += "\n"
		visibleIdx++

		if rg.Expanded {
			for _, res := range rg.Resources {
				resCursor := " "
				if visibleIdx == m.Cursor {
					resCursor = ">"
					s += SelectedStyle.Render(fmt.Sprintf("  %s 📄 %s (%s)", resCursor, res.Name, res.Type))
				} else {
					s += ResourceStyle.Render(fmt.Sprintf("  %s 📄 %s (%s)", resCursor, res.Name, res.Type))
				}
				s += "\n"
				visibleIdx++
			}
		}
	}

	s += "\nPress Enter/Space to expand/collapse, / to search, r to refresh, Esc to go back, q to quit"
	return s
}