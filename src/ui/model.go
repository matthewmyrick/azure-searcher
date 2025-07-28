package ui

import (
	"azure-searcher/src/azure"
	"azure-searcher/src/cache"
	"azure-searcher/src/config"
	"azure-searcher/src/types"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the main TUI model
type Model struct {
	// Application state
	State string // "subscriptions", "loading", "resources"

	// Data
	Subscriptions      []types.Subscription
	SelectedSub        types.Subscription
	ResourceGroups     []types.ResourceGroup
	FilteredGroups     []types.ResourceGroup

	// UI components
	SearchInput textinput.Model
	Spinner     spinner.Model
	Progress    progress.Model

	// Navigation
	Cursor       int
	ScrollOffset int

	// Services
	AzureClient  *azure.Client
	AzureFetcher *azure.Fetcher
	CacheManager *cache.Manager

	// Configuration
	Config config.ConcurrencyConfig

	// Error handling
	Err error

	// Cache and progress tracking
	LastLoadFromCache bool
	TotalRGs          int
	ProcessedRGs      int
	ProgressChan      chan types.ProgressUpdate
}

// NewModel creates a new TUI model
func NewModel() (*Model, error) {
	// Initialize search input
	ti := textinput.New()
	ti.Placeholder = "Search resource groups..."
	ti.CharLimit = 50
	ti.Width = 50

	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = CacheStatusStyle

	// Initialize progress bar
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 50

	// Get optimal concurrency configuration
	conf := config.GetOptimalConcurrency()

	// Initialize Azure client and fetcher
	azureClient := azure.NewClient()
	azureFetcher := azure.NewFetcher(azureClient, conf.RGConcurrency, conf.ResourceConcurrency)

	// Initialize cache manager
	cacheFile := cache.GetFilePath()
	cacheManager, err := cache.NewManager(cacheFile)
	if err != nil {
		return nil, err
	}

	return &Model{
		State:           "subscriptions",
		SearchInput:     ti,
		Spinner:         s,
		Progress:        prog,
		AzureClient:     azureClient,
		AzureFetcher:    azureFetcher,
		CacheManager:    cacheManager,
		Config:          conf,
		Subscriptions:   []types.Subscription{},
		ResourceGroups:  []types.ResourceGroup{},
		FilteredGroups:  []types.ResourceGroup{},
	}, nil
}

// Init implements the tea.Model interface
func (m *Model) Init() tea.Cmd {
	return InitCmd(m.AzureClient)
}