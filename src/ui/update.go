package ui

import (
	"azure-searcher/src/types"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all UI state updates
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case TickMsg:
		return m.handleTickMsg(msg)

	case ProgressUpdateMsg:
		return m.handleProgressUpdateMsg(msg)

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case SubscriptionsLoadedMsg:
		return m.handleSubscriptionsLoadedMsg(msg)

	case ResourceGroupsLoadedMsg:
		return m.handleResourceGroupsLoadedMsg(msg)

	case ErrorMsg:
		m.Err = msg.Err
		return m, tea.Quit

	case StartAsyncFetchMsg:
		m.State = "loading"
		m.TotalRGs = 0
		m.ProcessedRGs = 0
		// Create and store the progress channel
		m.ProgressChan = make(chan types.ProgressUpdate, 100)
		return m, tea.Batch(
			StartAsyncFetchCmd(msg.Model, m.ProgressChan),
			WaitForProgressCmd(m.ProgressChan),
			TickCmd(),
		)

	case tea.WindowSizeMsg:
		m.ViewportHeight = msg.Height
		m.ViewportWidth = msg.Width
		m.SearchInput.Width = msg.Width - 4 // Account for borders and padding
		if m.SearchInput.Width < 20 {
			m.SearchInput.Width = 20
		}
		m.Progress.Width = msg.Width - 4
		if m.Progress.Width < 20 {
			m.Progress.Width = 20
		}
	}

	return m, cmd
}

// handleTickMsg handles animation tick messages
func (m *Model) handleTickMsg(msg TickMsg) (tea.Model, tea.Cmd) {
	if m.State == "loading" {
		var spinnerCmd tea.Cmd
		m.Spinner, spinnerCmd = m.Spinner.Update(msg)
		return m, tea.Batch(spinnerCmd, TickCmd())
	}
	return m, nil
}

// handleProgressUpdateMsg handles progress update messages
func (m *Model) handleProgressUpdateMsg(msg ProgressUpdateMsg) (tea.Model, tea.Cmd) {
	if msg.Error != nil {
		m.Err = msg.Error
		return m, tea.Quit
	}

	// Update progress
	if msg.Total > 0 {
		m.TotalRGs = msg.Total
	}
	m.ProcessedRGs = msg.Processed

	// Check if completed
	if msg.Completed && len(msg.ResourceGroups) > 0 {
		return m.handleResourceGroupsLoadedMsg(ResourceGroupsLoadedMsg{
			ResourceGroups: msg.ResourceGroups,
			FromCache:      false,
		})
	}

	// Continue waiting for more updates if we have a progress channel
	var cmd tea.Cmd = TickCmd()
	if m.ProgressChan != nil {
		cmd = tea.Batch(TickCmd(), WaitForProgressCmd(m.ProgressChan))
	}
	return m, cmd
}

// handleKeyMsg handles keyboard input
func (m *Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.State {
	case "subscriptions":
		return m.handleSubscriptionKeys(msg)
	case "resources":
		return m.handleResourceKeys(msg)
	}
	return m, nil
}

// handleSubscriptionKeys handles keys in subscription selection state
func (m *Model) handleSubscriptionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		if m.Cursor < len(m.Subscriptions)-1 {
			m.Cursor++
		}
	case "enter":
		if len(m.Subscriptions) > 0 {
			m.SelectedSub = m.Subscriptions[m.Cursor]
			m.State = "loading"
			m.TotalRGs = 0
			m.ProcessedRGs = 0
			return m, tea.Batch(
				LoadResourceGroupsCmd(m),
				TickCmd(),
			)
		}
	}
	return m, nil
}

// handleResourceKeys handles keys in resource browsing state
func (m *Model) handleResourceKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "ctrl+q", "ctrl+c":
		return m, tea.Quit
	case "esc":
		// Only go back to subscriptions if search input is not focused
		if !m.SearchInput.Focused() {
			m.State = "subscriptions"
			m.Cursor = 0
			m.SearchInput.SetValue("")
			m.ResourceGroups = nil
			m.FilteredGroups = nil
			m.SearchInput.Blur()
		}
	case "ctrl+r":
		// Refresh cache - force reload from Azure
		m.CacheManager.InvalidateSubscription(m.SelectedSub.ID)
		m.State = "loading"
		m.TotalRGs = 0
		m.ProcessedRGs = 0
		m.ProgressChan = make(chan types.ProgressUpdate, 100)
		return m, tea.Batch(
			LoadResourceGroupsCmd(m),
			TickCmd(),
		)
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
			m.EnsureCursorInViewport()
		}
	case "down", "j":
		maxItems := m.countVisibleItems()
		if m.Cursor < maxItems-1 {
			m.Cursor++
			m.EnsureCursorInViewport()
		}
	case "pgup", "ctrl+u":
		viewportSize := m.CalculateResourceViewport()
		m.Cursor -= viewportSize / 2
		if m.Cursor < 0 {
			m.Cursor = 0
		}
		m.EnsureCursorInViewport()
	case "pgdown", "ctrl+d":
		viewportSize := m.CalculateResourceViewport()
		maxItems := m.countVisibleItems()
		m.Cursor += viewportSize / 2
		if m.Cursor >= maxItems {
			m.Cursor = maxItems - 1
		}
		if m.Cursor < 0 {
			m.Cursor = 0
		}
		m.EnsureCursorInViewport()
	case "home", "g":
		m.Cursor = 0
		m.ScrollOffset = 0
	case "end", "G":
		maxItems := m.countVisibleItems()
		if maxItems > 0 {
			m.Cursor = maxItems - 1
			m.EnsureCursorInViewport()
		}
	case "enter":
		m.toggleResourceGroup()
	case "/":
		m.SearchMode = "exact"
		m.SearchInput.Focus()
		m.filterResourceGroups()
		m.Cursor = 0
		m.ScrollOffset = 0
	case "\\":
		m.SearchMode = "fuzzy"
		m.SearchInput.Focus()
		m.filterResourceGroups()
		m.Cursor = 0
		m.ScrollOffset = 0
	default:
		if m.SearchInput.Focused() {
			// Handle special keys even when search is focused
			switch msg.String() {
			case "esc":
				m.SearchInput.Blur()
				m.SearchInput.SetValue("")
				m.filterResourceGroups() // This will reset to normal mode
				m.Cursor = 0
				m.ScrollOffset = 0
				return m, nil
			case "ctrl+e":
				// Exit search but keep the search text and filtered results
				m.SearchInput.Blur()
				m.Cursor = 0
				m.ScrollOffset = 0
				return m, nil
			case "ctrl+q", "ctrl+c":
				return m, tea.Quit
			case "ctrl+r":
				// Allow refresh even when search is focused
				m.CacheManager.InvalidateSubscription(m.SelectedSub.ID)
				m.State = "loading"
				m.TotalRGs = 0
				m.ProcessedRGs = 0
				m.ProgressChan = make(chan types.ProgressUpdate, 100)
				m.SearchInput.Blur() // Unfocus search when refreshing
				return m, tea.Batch(
					LoadResourceGroupsCmd(m),
					TickCmd(),
				)
			default:
				// Regular typing in search input
				m.SearchInput, cmd = m.SearchInput.Update(msg)
				m.filterResourceGroups()
				m.Cursor = 0
				m.ScrollOffset = 0
				return m, cmd
			}
		}
	}
	return m, nil
}

// handleSubscriptionsLoadedMsg handles subscription loading completion
func (m *Model) handleSubscriptionsLoadedMsg(msg SubscriptionsLoadedMsg) (tea.Model, tea.Cmd) {
	m.Subscriptions = msg.Subscriptions
	m.State = "subscriptions"
	return m, nil
}

// handleResourceGroupsLoadedMsg handles resource group loading completion
func (m *Model) handleResourceGroupsLoadedMsg(msg ResourceGroupsLoadedMsg) (tea.Model, tea.Cmd) {
	m.ResourceGroups = msg.ResourceGroups
	m.FilteredGroups = msg.ResourceGroups
	m.State = "resources"
	m.Cursor = 0
	m.LastLoadFromCache = msg.FromCache
	
	// Clean up progress channel
	if m.ProgressChan != nil {
		// Don't close here as it might already be closed by the goroutine
		m.ProgressChan = nil
	}
	
	return m, nil
}

// Helper methods

// filterResourceGroups filters resource groups based on search input and current search mode
func (m *Model) filterResourceGroups() {
	query := m.SearchInput.Value()
	
	// Use the appropriate search method based on search mode
	if m.SearchMode == "exact" {
		m.FilteredGroups = m.FuzzyMatcher.SearchResourceGroupsExact(query, m.ResourceGroups)
	} else {
		// Default to fuzzy search
		m.FilteredGroups = m.FuzzyMatcher.SearchResourceGroupsTwoPart(query, m.ResourceGroups)
	}
}

// toggleResourceGroup toggles the expansion state of a resource group
func (m *Model) toggleResourceGroup() {
	visibleIdx := 0
	for i := range m.FilteredGroups {
		if visibleIdx == m.Cursor {
			// Toggle in both FilteredGroups and ResourceGroups to maintain state
			m.FilteredGroups[i].Expanded = !m.FilteredGroups[i].Expanded
			
			// Also update the original ResourceGroups to preserve state
			for j := range m.ResourceGroups {
				if m.ResourceGroups[j].Name == m.FilteredGroups[i].Name {
					m.ResourceGroups[j].Expanded = m.FilteredGroups[i].Expanded
					break
				}
			}
			return
		}
		visibleIdx++
		if m.FilteredGroups[i].Expanded {
			visibleIdx += len(m.FilteredGroups[i].Resources)
		}
	}
}

// countVisibleItems counts the total number of visible items in the list
func (m *Model) countVisibleItems() int {
	count := 0
	for _, rg := range m.FilteredGroups {
		count++
		if rg.Expanded {
			count += len(rg.Resources)
		}
	}
	return count
}