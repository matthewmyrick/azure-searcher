package ui

import (
	"log"
	"time"

	"azure-searcher/src/types"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
)

// InitCmd returns the initial command for the application
func InitCmd(azureClient *azure.Client) tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		LoadSubscriptionsCmd(azureClient),
	)
}

// LoadSubscriptionsCmd loads Azure subscriptions
func LoadSubscriptionsCmd(azureClient *azure.Client) tea.Cmd {
	return func() tea.Msg {
		subs, err := azureClient.GetSubscriptions()
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return SubscriptionsLoadedMsg{Subscriptions: subs}
	}
}

// LoadResourceGroupsCmd loads resource groups for a subscription
func LoadResourceGroupsCmd(m *Model) tea.Cmd {
	return func() tea.Msg {
		// Try to get from cache first
		if cachedRGs, found := m.CacheManager.GetCachedResourceGroups(m.SelectedSub.ID); found {
			return ResourceGroupsLoadedMsg{ResourceGroups: cachedRGs, FromCache: true}
		}

		// Cache miss - fetch from Azure with progress updates
		m.ProgressChan = make(chan types.ProgressUpdate, 100)

		go func() {
			defer close(m.ProgressChan)

			rgs, err := m.AzureFetcher.FetchResourceGroups(m.SelectedSub.ID, m.ProgressChan)
			if err != nil {
				m.ProgressChan <- types.ProgressUpdate{Error: err}
				return
			}

			// Save to cache
			if saveErr := m.CacheManager.CacheResourceGroups(m.SelectedSub.ID, m.SelectedSub.Name, rgs); saveErr != nil {
				log.Printf("Failed to save cache: %v", saveErr)
			}

			// Send final result
			m.ProgressChan <- types.ProgressUpdate{
				ResourceGroups: rgs,
				Completed:      true,
			}
		}()

		return ResourceGroupsLoadedMsg{ResourceGroups: []types.ResourceGroup{}, FromCache: false}
	}
}

// TickCmd returns a command that sends tick messages for animations
func TickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// WaitForProgressCmd waits for progress updates
func WaitForProgressCmd(progressChan <-chan types.ProgressUpdate) tea.Cmd {
	return func() tea.Msg {
		update, ok := <-progressChan
		if !ok {
			// Channel closed, no more updates
			return nil
		}
		return ProgressUpdateMsg{
			Total:          update.Total,
			Processed:      update.Processed,
			ResourceGroups: update.ResourceGroups,
			Completed:      update.Completed,
			Error:          update.Error,
		}
	}
}