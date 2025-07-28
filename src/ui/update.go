package ui

import (
	"strings"

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
		m.ResourceGroups = msg.ResourceGroups
		m.FilteredGroups = msg.ResourceGroups
		m.State = "resources"
		m.Cursor = 0
		m.LastLoadFromCache = false
		if m.ProgressChan != nil {
			close(m.ProgressChan)
			m.ProgressChan = nil
		}
		return m, nil
	}

	// Continue waiting for more updates if not complete
	if m.ProgressChan != nil {
		return m, WaitForProgressCmd(m.ProgressChan)
	}
	return m, nil
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
	case "q", "ctrl+c":
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
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.State = "subscriptions"
		m.Cursor = 0
		m.SearchInput.SetValue("")
		m.ResourceGroups = nil
		m.FilteredGroups = nil
	case "r":
		// Refresh cache - force reload from Azure
		m.CacheManager.InvalidateSubscription(m.SelectedSub.ID)
		m.State = "loading"
		m.TotalRGs = 0
		m.ProcessedRGs = 0
		return m, tea.Batch(
			LoadResourceGroupsCmd(m),
			TickCmd(),
		)
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		maxItems := m.countVisibleItems()
		if m.Cursor < maxItems-1 {
			m.Cursor++
		}
	case "enter", " ":
		m.toggleResourceGroup()
	case "/":
		m.SearchInput.Focus()
	default:
		if m.SearchInput.Focused() {
			m.SearchInput, cmd = m.SearchInput.Update(msg)
			m.filterResourceGroups()
			m.Cursor = 0
			return m, cmd
		}
	}
	return m, nil
}

// handleSubscriptionsLoadedMsg handles subscription loading completion
func (m *Model) handleSubscriptionsLoadedMsg(msg SubscriptionsLoadedMsg) (tea.Model, tea.Cmd) {
	// Load subscriptions using the Azure client
	subs, err := m.AzureClient.GetSubscriptions()
	if err != nil {
		m.Err = err
		return m, tea.Quit
	}

	m.Subscriptions = subs
	m.State = "subscriptions"
	return m, nil
}

// handleResourceGroupsLoadedMsg handles resource group loading completion
func (m *Model) handleResourceGroupsLoadedMsg(msg ResourceGroupsLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.FromCache {
		m.ResourceGroups = msg.ResourceGroups
		m.FilteredGroups = msg.ResourceGroups
		m.State = "resources"
		m.Cursor = 0
		m.LastLoadFromCache = true
		return m, nil
	}

	// For non-cache loads, we need to set up progress tracking
	m.State = "loading"
	m.TotalRGs = 0
	m.ProcessedRGs = 0

	if m.ProgressChan != nil {
		return m, WaitForProgressCmd(m.ProgressChan)
	}

	return m, nil
}

// Helper methods

// filterResourceGroups filters resource groups based on search input
func (m *Model) filterResourceGroups() {
	query := strings.ToLower(m.SearchInput.Value())
	if query == "" {
		m.FilteredGroups = m.ResourceGroups
		return
	}

	var filtered []types.ResourceGroup
	for _, rg := range m.ResourceGroups {
		if strings.Contains(strings.ToLower(rg.Name), query) {
			filtered = append(filtered, rg)
		}
	}
	m.FilteredGroups = filtered
}

// toggleResourceGroup toggles the expansion state of a resource group
func (m *Model) toggleResourceGroup() {
	visibleIdx := 0
	for i := range m.FilteredGroups {
		if visibleIdx == m.Cursor {
			m.FilteredGroups[i].Expanded = !m.FilteredGroups[i].Expanded
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