package ui

import (
	"azure-searcher/src/azure"
	"azure-searcher/src/cache"
	"azure-searcher/src/config"
	"azure-searcher/src/search"
	"azure-searcher/src/types"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the main TUI model
type Model struct {
	// Application state
	State      string // "subscriptions", "loading", "resources", "config"
	SearchMode string // "fuzzy" or "exact"
	
	// Configuration
	ConfiguringSub    types.Subscription // Subscription being configured
	ConfigCursor     int                 // Current cursor position in config
	ConfigMode       string              // "menu" or "interval_input"
	IntervalInput    textinput.Model     // Input for refresh interval

	// Data
	Subscriptions      []types.Subscription
	SelectedSub        types.Subscription
	ResourceGroups     []types.ResourceGroup
	FilteredGroups     []types.ResourceGroup

	// UI components
	SearchInput textinput.Model
	Spinner     spinner.Model
	Progress    progress.Model

	// Navigation and viewport
	Cursor         int
	ScrollOffset   int
	ViewportHeight int
	ViewportWidth  int

	// Services
	AzureClient  *azure.Client
	AzureFetcher *azure.Fetcher
	CacheManager *cache.Manager
	FuzzyMatcher *search.FuzzyMatcher

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
	ti.Placeholder = "Search: <resourcegroup> <resource> (e.g., 'prod vm' or just 'prod')..."
	ti.CharLimit = 50
	ti.Width = 50

	// Initialize interval input
	intervalInput := textinput.New()
	intervalInput.Placeholder = "Enter refresh interval (e.g., '2 hr', '30 min', '1 dy')"
	intervalInput.CharLimit = 20
	intervalInput.Width = 40

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

	// Initialize fuzzy matcher
	fuzzyMatcher := search.NewFuzzyMatcher()

	return &Model{
		State:           "subscriptions",
		SearchMode:      "exact", // Default to exact search
		ConfigMode:      "menu",  // Default config mode
		SearchInput:     ti,
		IntervalInput:   intervalInput,
		Spinner:         s,
		Progress:        prog,
		AzureClient:     azureClient,
		AzureFetcher:    azureFetcher,
		CacheManager:    cacheManager,
		FuzzyMatcher:    fuzzyMatcher,
		Config:          conf,
		Subscriptions:   []types.Subscription{},
		ResourceGroups:  []types.ResourceGroup{},
		FilteredGroups:  []types.ResourceGroup{},
		ViewportHeight:  20, // Default height, will be updated on window size
		ViewportWidth:   80, // Default width, will be updated on window size
	}, nil
}

// Init implements the tea.Model interface
func (m *Model) Init() tea.Cmd {
	return InitCmd(m.AzureClient)
}

// CalculateResourceViewport calculates the available space for the resource list
func (m *Model) CalculateResourceViewport() int {
	// Account for:
	// - Title (2 lines: title + cache status)
	// - Search input (3 lines: input + borders + spacing)
	// - Help text (1 line)
	// - Some padding (2 lines)
	return m.ViewportHeight - 8
}

// EnsureCursorInViewport adjusts scroll offset to keep cursor visible
func (m *Model) EnsureCursorInViewport() {
	viewportSize := m.CalculateResourceViewport()
	
	// If cursor is above the viewport, scroll up
	if m.Cursor < m.ScrollOffset {
		m.ScrollOffset = m.Cursor
	}
	
	// If cursor is below the viewport, scroll down
	if m.Cursor >= m.ScrollOffset+viewportSize {
		m.ScrollOffset = m.Cursor - viewportSize + 1
		if m.ScrollOffset < 0 {
			m.ScrollOffset = 0
		}
	}
}