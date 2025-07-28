package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	// Default concurrency limits - can be adjusted based on system performance
	DEFAULT_RG_CONCURRENCY       = 5  // Max concurrent resource groups
	DEFAULT_RESOURCE_CONCURRENCY = 10 // Max concurrent resources per group
	
	// Cache settings
	DEFAULT_CACHE_TTL_MINUTES = 30 // Cache TTL in minutes
	CACHE_FILENAME           = "azure-searcher-cache.json"
)

type subscription struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type resourceGroup struct {
	Name      string     `json:"name"`
	Location  string     `json:"location"`
	Resources []resource `json:"resources"`
	Expanded  bool       `json:"-"`
}

type resource struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Kind     string `json:"kind"`
	ID       string `json:"id"`
	AzureURL string `json:"azure_url"`
}

type cachedResourceGroup struct {
	Name      string              `json:"name"`
	Location  string              `json:"location"`
	Resources []resource          `json:"resources"`
	Expanded  bool                `json:"-"`
	CachedAt  time.Time           `json:"cached_at"`
}

type subscriptionCache struct {
	SubscriptionID   string                         `json:"subscription_id"`
	SubscriptionName string                         `json:"subscription_name"`
	ResourceGroups   map[string]cachedResourceGroup `json:"resource_groups"`
	LastUpdated      time.Time                      `json:"last_updated"`
}

type cacheData struct {
	Subscriptions map[string]subscriptionCache `json:"subscriptions"`
	Version       string                       `json:"version"`
}

type model struct {
	state              string // "subscriptions", "loading", "resources"
	subscriptions      []subscription
	selectedSub        subscription
	resourceGroups     []resourceGroup
	filteredGroups     []resourceGroup
	searchInput        textinput.Model
	spinner            spinner.Model
	progress           progress.Model
	cursor             int
	scrollOffset       int
	err                error
	rgConcurrency      int // Max concurrent resource groups
	resourceConcurrency int // Max concurrent resources per group
	cacheFile          string
	cache              *cacheData
	lastLoadFromCache  bool
	totalRGs           int
	processedRGs       int
	progressChan       chan progressUpdateMsg
}

var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))
	
	selectedStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#874BFD")).
		Foreground(lipgloss.Color("#FFFFFF"))
	
	resourceGroupStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Bold(true)
	
	resourceStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))
)

func checkAzCLI() error {
	_, err := exec.LookPath("az")
	if err != nil {
		return fmt.Errorf("Azure CLI not found. Please install it from: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli")
	}
	return nil
}

func checkAzLogin() error {
	cmd := exec.Command("az", "account", "show")
	err := cmd.Run()
	if err != nil {
		fmt.Println("You are not logged into Azure. Running 'az login'...")
		loginCmd := exec.Command("az", "login")
		loginCmd.Stdout = os.Stdout
		loginCmd.Stderr = os.Stderr
		return loginCmd.Run()
	}
	return nil
}

func getSubscriptions() ([]subscription, error) {
	cmd := exec.Command("az", "account", "list", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	var subs []subscription
	err = json.Unmarshal(output, &subs)
	return subs, err
}

func getResourceGroups(subscriptionID string, rgConcurrency, resourceConcurrency int, progressChan chan<- progressUpdateMsg) ([]resourceGroup, error) {
	cmd := exec.Command("az", "group", "list", "--subscription", subscriptionID, "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	var rgs []resourceGroup
	err = json.Unmarshal(output, &rgs)
	if err != nil {
		return nil, err
	}
	
	// Send initial progress update with total count
	if progressChan != nil {
		progressChan <- progressUpdateMsg{
			total:     len(rgs),
			processed: 0,
		}
	}
	
	// Create semaphore to limit concurrent resource group processing
	rgSemaphore := make(chan struct{}, rgConcurrency)
	var wg sync.WaitGroup
	rgChan := make(chan resourceGroup, len(rgs))
	var processedCount int64
	
	for _, rg := range rgs {
		wg.Add(1)
		go func(rgName string) {
			defer wg.Done()
			
			// Acquire semaphore
			rgSemaphore <- struct{}{}
			defer func() { <-rgSemaphore }()
			
			// Get list of resources in the resource group first
			cmd := exec.Command("az", "resource", "list", "--resource-group", rgName, "--subscription", subscriptionID, "--output", "json")
			output, err := cmd.Output()
			if err != nil {
				rgChan <- resourceGroup{
					Name:      rgName,
					Resources: []resource{},
					Expanded:  false,
				}
				// Update progress even on error
				processed := atomic.AddInt64(&processedCount, 1)
				if progressChan != nil {
					progressChan <- progressUpdateMsg{
						total:     len(rgs),
						processed: int(processed),
					}
				}
				return
			}
			
			var rawResources []resource
			err = json.Unmarshal(output, &rawResources)
			if err != nil {
				rgChan <- resourceGroup{
					Name:      rgName,
					Resources: []resource{},
					Expanded:  false,
				}
				// Update progress even on error
				processed := atomic.AddInt64(&processedCount, 1)
				if progressChan != nil {
					progressChan <- progressUpdateMsg{
						total:     len(rgs),
						processed: int(processed),
					}
				}
				return
			}
			
			// Use limited goroutines to process resources within this group
			resourceSemaphore := make(chan struct{}, resourceConcurrency)
			var resourceWg sync.WaitGroup
			resChan := make(chan resource, len(rawResources))
			
			for _, res := range rawResources {
				resourceWg.Add(1)
				go func(r resource) {
					defer resourceWg.Done()
					
					// Acquire resource semaphore
					resourceSemaphore <- struct{}{}
					defer func() { <-resourceSemaphore }()
					
					// Add Azure portal URL to the resource
					r.AzureURL = generateAzurePortalURL(subscriptionID, r.ID)
					resChan <- r
				}(res)
			}
			
			go func() {
				resourceWg.Wait()
				close(resChan)
			}()
			
			var resources []resource
			for res := range resChan {
				resources = append(resources, res)
			}
			
			// Sort resources by name
			sort.Slice(resources, func(i, j int) bool {
				return resources[i].Name < resources[j].Name
			})
			
			rgChan <- resourceGroup{
				Name:      rgName,
				Resources: resources,
				Expanded:  false,
			}
			
			// Update progress
			processed := atomic.AddInt64(&processedCount, 1)
			if progressChan != nil {
				progressChan <- progressUpdateMsg{
					total:     len(rgs),
					processed: int(processed),
				}
			}
		}(rg.Name)
	}
	
	go func() {
		wg.Wait()
		close(rgChan)
	}()
	
	var result []resourceGroup
	for rg := range rgChan {
		result = append(result, rg)
	}
	
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	
	return result, nil
}

func getResourcesInGroup(subscriptionID, resourceGroupName string) ([]resource, error) {
	cmd := exec.Command("az", "resource", "list", "--resource-group", resourceGroupName, "--subscription", subscriptionID, "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	var resources []resource
	err = json.Unmarshal(output, &resources)
	if err != nil {
		return nil, err
	}
	
	// Add Azure portal URLs to resources
	for i := range resources {
		resources[i].AzureURL = generateAzurePortalURL(subscriptionID, resources[i].ID)
	}
	
	return resources, err
}

// Cache management functions
func getCacheFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return CACHE_FILENAME
	}
	return filepath.Join(homeDir, ".azure-searcher", CACHE_FILENAME)
}

func generateAzurePortalURL(subscriptionID, resourceID string) string {
	if resourceID == "" {
		return ""
	}
	return fmt.Sprintf("https://portal.azure.com/#resource%s", resourceID)
}

func loadCache(cacheFile string) (*cacheData, error) {
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &cacheData{
				Subscriptions: make(map[string]subscriptionCache),
				Version:       "1.0",
			}, nil
		}
		return nil, err
	}
	
	var cache cacheData
	err = json.Unmarshal(data, &cache)
	if err != nil {
		return &cacheData{
			Subscriptions: make(map[string]subscriptionCache),
			Version:       "1.0",
		}, nil
	}
	
	return &cache, nil
}

func saveCache(cacheFile string, cache *cacheData) error {
	dir := filepath.Dir(cacheFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(cacheFile, data, 0644)
}

func isCacheValid(cachedAt time.Time, ttlMinutes int) bool {
	return time.Since(cachedAt) < time.Duration(ttlMinutes)*time.Minute
}

func (c *cacheData) getCachedResourceGroups(subscriptionID string) ([]resourceGroup, bool) {
	subCache, exists := c.Subscriptions[subscriptionID]
	if !exists {
		return nil, false
	}
	
	if !isCacheValid(subCache.LastUpdated, DEFAULT_CACHE_TTL_MINUTES) {
		return nil, false
	}
	
	var resourceGroups []resourceGroup
	for _, cachedRG := range subCache.ResourceGroups {
		if isCacheValid(cachedRG.CachedAt, DEFAULT_CACHE_TTL_MINUTES) {
			resourceGroups = append(resourceGroups, resourceGroup{
				Name:      cachedRG.Name,
				Location:  cachedRG.Location,
				Resources: cachedRG.Resources,
				Expanded:  false,
			})
		}
	}
	
	if len(resourceGroups) == 0 {
		return nil, false
	}
	
	sort.Slice(resourceGroups, func(i, j int) bool {
		return resourceGroups[i].Name < resourceGroups[j].Name
	})
	
	return resourceGroups, true
}

func (c *cacheData) cacheResourceGroups(subscriptionID, subscriptionName string, resourceGroups []resourceGroup) {
	if c.Subscriptions == nil {
		c.Subscriptions = make(map[string]subscriptionCache)
	}
	
	cachedRGs := make(map[string]cachedResourceGroup)
	now := time.Now()
	
	for _, rg := range resourceGroups {
		cachedRGs[rg.Name] = cachedResourceGroup{
			Name:      rg.Name,
			Location:  rg.Location,
			Resources: rg.Resources,
			CachedAt:  now,
		}
	}
	
	c.Subscriptions[subscriptionID] = subscriptionCache{
		SubscriptionID:   subscriptionID,
		SubscriptionName: subscriptionName,
		ResourceGroups:   cachedRGs,
		LastUpdated:      now,
	}
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Search resource groups..."
	ti.CharLimit = 50
	ti.Width = 50
	
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	
	// Initialize progress bar
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 50
	
	// Set concurrency limits based on system resources
	rgConcurrency := DEFAULT_RG_CONCURRENCY
	resourceConcurrency := DEFAULT_RESOURCE_CONCURRENCY
	
	// Adjust limits based on CPU count for better performance
	cpuCount := runtime.NumCPU()
	if cpuCount <= 2 {
		rgConcurrency = 2
		resourceConcurrency = 5
	} else if cpuCount <= 4 {
		rgConcurrency = 3
		resourceConcurrency = 8
	}
	
	// Initialize cache
	cacheFile := getCacheFilePath()
	cache, err := loadCache(cacheFile)
	if err != nil {
		cache = &cacheData{
			Subscriptions: make(map[string]subscriptionCache),
			Version:       "1.0",
		}
	}
	
	return model{
		state:               "subscriptions",
		searchInput:         ti,
		spinner:             s,
		progress:            prog,
		rgConcurrency:       rgConcurrency,
		resourceConcurrency: resourceConcurrency,
		cacheFile:           cacheFile,
		cache:               cache,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.spinner.Tick,
		loadSubscriptionsCmd,
	)
}

func loadSubscriptionsCmd() tea.Msg {
	if err := checkAzCLI(); err != nil {
		return errMsg{err}
	}
	
	if err := checkAzLogin(); err != nil {
		return errMsg{err}
	}
	
	subs, err := getSubscriptions()
	if err != nil {
		return errMsg{err}
	}
	
	return subscriptionsLoadedMsg{subs}
}

func loadResourceGroupsCmd(subID, subName string, rgConcurrency, resourceConcurrency int, cache *cacheData, cacheFile string, progressChan chan progressUpdateMsg) tea.Cmd {
	return func() tea.Msg {
		// Try to get from cache first
		if cachedRGs, found := cache.getCachedResourceGroups(subID); found {
			return resourceGroupsLoadedMsg{cachedRGs, true} // true indicates cache hit
		}
		
		// Cache miss - fetch from Azure with progress updates
		rgs, err := getResourceGroups(subID, rgConcurrency, resourceConcurrency, progressChan)
		if err != nil {
			return errMsg{err}
		}
		
		// Save to cache
		cache.cacheResourceGroups(subID, subName, rgs)
		if saveErr := saveCache(cacheFile, cache); saveErr != nil {
			log.Printf("Failed to save cache: %v", saveErr)
		}
		
		return resourceGroupsLoadedMsg{rgs, false} // false indicates fresh fetch
	}
}

func waitForProgressCmd(progressChan <-chan progressUpdateMsg) tea.Cmd {
	return func() tea.Msg {
		update, ok := <-progressChan
		if !ok {
			// Channel closed, no more updates
			return nil
		}
		return update
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type subscriptionsLoadedMsg struct {
	subscriptions []subscription
}

type resourceGroupsLoadedMsg struct {
	resourceGroups []resourceGroup
	fromCache      bool
}

type progressUpdateMsg struct {
	total          int
	processed      int
	resourceGroups []resourceGroup
	completed      bool
	err            error
}

type startProgressFetchMsg struct {
	subID               string
	subName             string
	rgConcurrency       int
	resourceConcurrency int
	cache               *cacheData
	cacheFile           string
}

type tickMsg time.Time

type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	return e.err.Error()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tickMsg:
		if m.state == "loading" {
			var spinnerCmd tea.Cmd
			m.spinner, spinnerCmd = m.spinner.Update(msg)
			return m, tea.Batch(spinnerCmd, tickCmd())
		}
		return m, nil
	
	case progressUpdateMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		
		// Update progress
		if msg.total > 0 {
			m.totalRGs = msg.total
		}
		m.processedRGs = msg.processed
		
		// Continue waiting for more updates if not complete
		if m.progressChan != nil {
			return m, waitForProgressCmd(m.progressChan)
		}
		return m, nil
		
	case tea.KeyMsg:
		switch m.state {
		case "subscriptions":
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.subscriptions)-1 {
					m.cursor++
				}
			case "enter":
				if len(m.subscriptions) > 0 {
					m.selectedSub = m.subscriptions[m.cursor]
					m.state = "loading"
					m.totalRGs = 0
					m.processedRGs = 0
					m.progressChan = make(chan progressUpdateMsg, 100)
					return m, tea.Batch(
						loadResourceGroupsCmd(m.selectedSub.ID, m.selectedSub.Name, m.rgConcurrency, m.resourceConcurrency, m.cache, m.cacheFile, m.progressChan),
						waitForProgressCmd(m.progressChan),
						tickCmd(),
					)
				}
			}
		case "resources":
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.state = "subscriptions"
				m.cursor = 0
				m.searchInput.SetValue("")
				m.resourceGroups = nil
				m.filteredGroups = nil
			case "r":
				// Refresh cache - force reload from Azure
				delete(m.cache.Subscriptions, m.selectedSub.ID)
				m.state = "loading"
				m.totalRGs = 0
				m.processedRGs = 0
				m.progressChan = make(chan progressUpdateMsg, 100)
				return m, tea.Batch(
					loadResourceGroupsCmd(m.selectedSub.ID, m.selectedSub.Name, m.rgConcurrency, m.resourceConcurrency, m.cache, m.cacheFile, m.progressChan),
					waitForProgressCmd(m.progressChan),
					tickCmd(),
				)
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				maxItems := m.countVisibleItems()
				if m.cursor < maxItems-1 {
					m.cursor++
				}
			case "enter", " ":
				m.toggleResourceGroup()
			case "/":
				m.searchInput.Focus()
			default:
				if m.searchInput.Focused() {
					m.searchInput, cmd = m.searchInput.Update(msg)
					m.filterResourceGroups()
					m.cursor = 0
					return m, cmd
				}
			}
		}
	
	case subscriptionsLoadedMsg:
		m.subscriptions = msg.subscriptions
		m.state = "subscriptions"
	
	case resourceGroupsLoadedMsg:
		m.resourceGroups = msg.resourceGroups
		m.filteredGroups = m.resourceGroups
		m.state = "resources"
		m.cursor = 0
		m.lastLoadFromCache = msg.fromCache
		// Close progress channel when done
		if m.progressChan != nil {
			close(m.progressChan)
			m.progressChan = nil
		}
	
	case errMsg:
		m.err = msg.err
		return m, tea.Quit
	}
	
	return m, cmd
}

func (m *model) filterResourceGroups() {
	query := strings.ToLower(m.searchInput.Value())
	if query == "" {
		m.filteredGroups = m.resourceGroups
		return
	}
	
	var filtered []resourceGroup
	for _, rg := range m.resourceGroups {
		if strings.Contains(strings.ToLower(rg.Name), query) {
			filtered = append(filtered, rg)
		}
	}
	m.filteredGroups = filtered
}

func (m *model) toggleResourceGroup() {
	visibleIdx := 0
	for i := range m.filteredGroups {
		if visibleIdx == m.cursor {
			m.filteredGroups[i].Expanded = !m.filteredGroups[i].Expanded
			return
		}
		visibleIdx++
		if m.filteredGroups[i].Expanded {
			visibleIdx += len(m.filteredGroups[i].Resources)
		}
	}
}

func (m model) countVisibleItems() int {
	count := 0
	for _, rg := range m.filteredGroups {
		count++
		if rg.Expanded {
			count += len(rg.Resources)
		}
	}
	return count
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress any key to exit.", m.err)
	}
	
	switch m.state {
	case "subscriptions":
		return m.viewSubscriptions()
	case "loading":
		var progressBar string
		if m.totalRGs > 0 {
			percentage := float64(m.processedRGs) / float64(m.totalRGs)
			progressBar = m.progress.ViewAs(percentage)
			return fmt.Sprintf("\n%s Loading resource groups for %s...\n\n%s\n\nProgress: %d/%d resource groups processed (%.0f%%)\n\nFetching data with %d concurrent resource groups and %d concurrent resources per group.", 
				m.spinner.View(), m.selectedSub.Name, progressBar, m.processedRGs, m.totalRGs, percentage*100, m.rgConcurrency, m.resourceConcurrency)
		} else {
			return fmt.Sprintf("\n%s Loading resource groups for %s...\n\nInitializing...\n\nFetching data with %d concurrent resource groups and %d concurrent resources per group.", 
				m.spinner.View(), m.selectedSub.Name, m.rgConcurrency, m.resourceConcurrency)
		}
	case "resources":
		return m.viewResources()
	}
	
	return "Unknown state"
}

func (m model) viewSubscriptions() string {
	s := titleStyle.Render("Select Azure Subscription") + "\n\n"
	
	for i, sub := range m.subscriptions {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
			s += selectedStyle.Render(fmt.Sprintf("%s %s", cursor, sub.Name))
		} else {
			s += fmt.Sprintf("%s %s", cursor, sub.Name)
		}
		s += "\n"
	}
	
	s += "\nPress Enter to select, q to quit"
	return s
}

func (m model) viewResources() string {
	cacheStatus := "📡 Live data"
	if m.lastLoadFromCache {
		cacheStatus = "⚡ Cached data"
	}
	
	s := titleStyle.Render(fmt.Sprintf("Resources in %s", m.selectedSub.Name)) + "\n"
	s += lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(cacheStatus) + "\n\n"
	
	searchView := m.searchInput.View()
	if m.searchInput.Focused() {
		searchView = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(searchView)
	}
	s += searchView + "\n\n"
	
	visibleIdx := 0
	for _, rg := range m.filteredGroups {
		cursor := " "
		icon := "📁"
		if rg.Expanded {
			icon = "📂"
		}
		
		if visibleIdx == m.cursor {
			cursor = ">"
			s += selectedStyle.Render(fmt.Sprintf("%s %s %s (%d resources)", cursor, icon, rg.Name, len(rg.Resources)))
		} else {
			s += resourceGroupStyle.Render(fmt.Sprintf("%s %s %s (%d resources)", cursor, icon, rg.Name, len(rg.Resources)))
		}
		s += "\n"
		visibleIdx++
		
		if rg.Expanded {
			for _, res := range rg.Resources {
				resCursor := " "
				if visibleIdx == m.cursor {
					resCursor = ">"
					s += selectedStyle.Render(fmt.Sprintf("  %s 📄 %s (%s)", resCursor, res.Name, res.Type))
				} else {
					s += resourceStyle.Render(fmt.Sprintf("  %s 📄 %s (%s)", resCursor, res.Name, res.Type))
				}
				s += "\n"
				visibleIdx++
			}
		}
	}
	
	s += "\nPress Enter/Space to expand/collapse, / to search, r to refresh, Esc to go back, q to quit"
	return s
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}