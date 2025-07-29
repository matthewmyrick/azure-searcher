package ui

import (
	"time"

	"azure-searcher/src/types"
)

// Message types for the TUI application
type (
	// SubscriptionsLoadedMsg indicates subscriptions have been loaded
	SubscriptionsLoadedMsg struct {
		Subscriptions []types.Subscription
	}

	// ResourceGroupsLoadedMsg indicates resource groups have been loaded
	ResourceGroupsLoadedMsg struct {
		ResourceGroups []types.ResourceGroup
		FromCache      bool
	}

	// ProgressUpdateMsg provides progress updates during resource fetching
	ProgressUpdateMsg struct {
		Total          int
		Processed      int
		ResourceGroups []types.ResourceGroup
		Completed      bool
		Error          error
	}

	// TickMsg represents a tick for animations
	TickMsg time.Time

	// ErrorMsg represents an error that occurred
	ErrorMsg struct {
		Err error
	}

	// StartAsyncFetchMsg indicates that async fetching should start
	StartAsyncFetchMsg struct {
		Model *Model
	}

	// ShowResourceTypesMsg shows available resource types for filtering
	ShowResourceTypesMsg struct{}
)

func (e ErrorMsg) Error() string {
	return e.Err.Error()
}