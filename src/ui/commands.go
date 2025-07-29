package ui

import (
	"log"
	"os/exec"
	"runtime"
	"time"

	"azure-searcher/src/azure"
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
		// Check if caching is enabled for this subscription
		if m.CacheManager.GetSubscriptionCacheEnabled(m.SelectedSub.ID) {
			// Check if we should refresh based on auto-refresh settings
			if m.CacheManager.ShouldRefreshCache(m.SelectedSub.ID) {
				// Auto-refresh is triggered, fetch fresh data
				return StartAsyncFetchMsg{Model: m}
			}
			
			// Try to get from cache
			if cachedRGs, found := m.CacheManager.GetCachedResourceGroups(m.SelectedSub.ID); found {
				return ResourceGroupsLoadedMsg{ResourceGroups: cachedRGs, FromCache: true}
			}
		}

		// Cache disabled or cache miss - start the async fetching process
		return StartAsyncFetchMsg{Model: m}
	}
}

// StartAsyncFetchCmd starts async resource group fetching with progress updates
func StartAsyncFetchCmd(m *Model, progressChan chan types.ProgressUpdate) tea.Cmd {
	return func() tea.Msg {
		go func() {
			defer close(progressChan)

			rgs, err := m.AzureFetcher.FetchResourceGroups(m.SelectedSub.ID, progressChan)
			if err != nil {
				progressChan <- types.ProgressUpdate{Error: err}
				return
			}

			// Save to cache only if caching is enabled for this subscription
			if m.CacheManager.GetSubscriptionCacheEnabled(m.SelectedSub.ID) {
				if saveErr := m.CacheManager.CacheResourceGroups(m.SelectedSub.ID, m.SelectedSub.Name, rgs); saveErr != nil {
					log.Printf("Failed to save cache: %v", saveErr)
				}
			}

			// Send final result
			progressChan <- types.ProgressUpdate{
				ResourceGroups: rgs,
				Completed:      true,
			}
		}()

		// Return nil since we're handling progress via the separate WaitForProgressCmd
		return nil
	}
}

// TickCmd returns a command that sends tick messages for animations
func TickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// WaitForProgressCmd waits for a single progress update
func WaitForProgressCmd(progressChan <-chan types.ProgressUpdate) tea.Cmd {
	return func() tea.Msg {
		update, ok := <-progressChan
		if !ok {
			// Channel closed
			return nil
		}
		
		if update.Error != nil {
			return ProgressUpdateMsg{Error: update.Error}
		}
		
		if update.Completed {
			return ProgressUpdateMsg{
				Total:          update.Total,
				Processed:      update.Processed,
				ResourceGroups: update.ResourceGroups,
				Completed:      true,
			}
		}
		
		// Send intermediate progress update
		return ProgressUpdateMsg{
			Total:     update.Total,
			Processed: update.Processed,
		}
	}
}

// openURLCmd opens a URL in the default system browser
func openURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "linux":
			cmd = exec.Command("xdg-open", url)
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", url)
		default:
			// Unsupported platform, return nil
			return nil
		}
		
		if err := cmd.Start(); err != nil {
			// Log error but don't crash the app
			log.Printf("Failed to open URL: %v", err)
		}
		
		return nil
	}
}

// showResourceTypesCmd shows available resource type filters
func showResourceTypesCmd() tea.Cmd {
	return func() tea.Msg {
		// This will trigger a help message to be shown
		return ShowResourceTypesMsg{}
	}
}