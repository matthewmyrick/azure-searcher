# Azure Searcher TUI

A fast, interactive Terminal User Interface (TUI) application for browsing Azure resources with advanced parallel processing, search, and caching capabilities.

## Features

- **Azure CLI Integration**: Automatic login detection and subscription management
- **Intelligent Caching System**:
  - JSON file-based cache with 30-minute TTL
  - Automatic cache validation and refresh
  - Azure portal URLs included for each resource
  - Cache stored in `~/.azure-searcher/` directory
- **High-Performance Parallel Processing**:
  - Goroutines per resource group for concurrent data fetching
  - Nested goroutines per resource within each resource group for maximum speed
  - Smart concurrency limits to prevent system slowdown
- **Interactive TUI**: Clean, intuitive interface with keyboard navigation
- **Animated Loading**: Spinner animation and progress bar during resource group loading
- **Two-Part Search**: Simple search format - `<resourcegroup> [space] <resource>` for intuitive filtering
- **Collapsible Tree Structure**: Expand/collapse resource groups like folders
- **Visual Icons**: Different icons for resource groups (📁/📂) vs resources (📄)
- **Cache Status Indicators**: Visual indicators showing cached vs live data

## Requirements

- Go 1.23 or later
- Azure CLI (`az`) installed and configured
- Active Azure subscription

## Installation

### Using go install (Recommended)

Install directly from GitHub:

```bash
go install github.com/matthewmyrick/azure-searcher@latest
```

This will install the `azure-searcher` binary to your `$GOPATH/bin` directory (or `$HOME/go/bin` if GOPATH is not set).

### Building from source

```bash
git clone https://github.com/matthewmyrick/azure-searcher
cd azure-searcher
go build -o azure-searcher
```

## Usage

Run the application:

```bash
azure-searcher
```

If installed via `go install`, make sure `$GOPATH/bin` (or `$HOME/go/bin`) is in your PATH.

## Controls

### Subscription Selection

- **Arrow keys/j,k**: Navigate up/down
- **Enter**: Select subscription
- **Ctrl+Q**: Quit

### Resource Navigation

- **Arrow keys/j,k**: Navigate up/down
- **Enter**: Expand/collapse resource groups
- **/**: Focus two-part search bar (press Esc to clear search)
- **Ctrl+R**: Refresh data (bypass cache)
- **PgUp/PgDn**: Scroll half-page up/down
- **Home/g**: Jump to top
- **End/G**: Jump to bottom
- **Esc**: Return to subscription selection (or exit search mode if searching)
- **Ctrl+Q**: Quit

## Performance Optimizations

The application uses a controlled multi-level parallel processing approach:

1. **Resource Group Level**: Resource groups are processed with limited concurrency (default: 5 concurrent)
2. **Resource Level**: Within each resource group, resources are processed in parallel with limits (default: 10 concurrent per group)
3. **CPU-Based Scaling**: Concurrency limits automatically adjust based on your system's CPU count
4. **Semaphore Pattern**: Uses buffered channels as semaphores to prevent system overload

### Concurrency Limits by CPU Count:

- **≤2 CPUs**: 2 resource groups, 5 resources per group
- **≤4 CPUs**: 3 resource groups, 8 resources per group
- **>4 CPUs**: 5 resource groups, 10 resources per group

This approach significantly reduces loading time while preventing system slowdown from excessive goroutines.

## Two-Part Search

The application features a simple and intuitive two-part search system:

### Search Format

```
<resourcegroup> [space] <resource>
```

### How Two-Part Search Works

- **Part 1 (Resource Group)**: The first part filters resource groups by name
- **Part 2 (Resource)**: The second part (after space) filters resources within the matched groups
- **Single Part**: If no space, only filters resource groups
- **Maintains Structure**: Keeps the beautiful folder tree structure during search

### Search Behavior

- **Resource Group Filtering**: First part matches against resource group names
- **Resource Filtering**: Second part (if provided) filters resources within matched groups
- **Auto-Expansion**: Resource groups automatically expand when resource filtering is active
- **Fuzzy Matching**: Supports both exact substring and fuzzy character matching
- **Real-Time**: Results update as you type

### Search Examples

- `"prod"` → Shows all resource groups containing "prod"
- `"prod vm"` → Shows resource groups with "prod", then filters for resources with "vm"
- `"east storage"` → Shows "east" resource groups, filtered for "storage" resources
- `"app func"` → Shows "app" resource groups, filtered for "func" resources
- `"db mysql"` → Shows "db" resource groups, filtered for "mysql" resources

### Navigation During Search

- **Arrow Keys**: Navigate through the filtered tree structure
- **Enter**: Expand/collapse resource group folders
- **Clear Search**: Press `Esc` to clear search and return to full tree
- **Visual Consistency**: Same folder icons (📁/📂) and resource icons (📄)

### Search Logic

1. **Step 1**: Filter resource groups by the first part of your query
2. **Step 2**: If there's a space and second part, filter resources within those groups
3. **Step 3**: Auto-expand groups when resource filtering is active
4. **Step 4**: Maintain folder structure and navigation

### Performance

- Fast string matching with fallback to fuzzy search
- Real-time filtering as you type
- Maintains folder structure and expansion states
- Simple and predictable search behavior

## Caching System

The application features an intelligent JSON file-based caching system:

### Cache Behavior

- **Automatic Caching**: All Azure resource data is automatically cached after first fetch
- **TTL (Time To Live)**: 30-minute cache expiration for optimal balance of performance and data freshness
- **Cache Location**: `~/.azure-searcher/azure-searcher-cache.json`
- **Azure Portal URLs**: Direct links to Azure portal for each resource

### Cache Management

- **Automatic Validation**: Cache is automatically checked for expiration on each access
- **Manual Refresh**: Press `r` to force refresh from Azure (bypasses cache)
- **Visual Indicators**:
  - ⚡ Cached data - Loaded from local cache
  - 📡 Live data - Freshly fetched from Azure
- **Per-Subscription**: Cache is managed separately for each subscription

### Benefits

- **Speed**: Subsequent loads are nearly instantaneous
- **Reduced API Calls**: Minimizes Azure CLI API usage
- **Offline Viewing**: Browse previously loaded data without network connectivity
- **System Performance**: Reduces system load from repeated Azure CLI calls

### Cache File Structure

The cache file is human-readable JSON with the following structure:

```json
{
  "subscriptions": {
    "subscription-id": {
      "subscription_id": "sub-12345",
      "subscription_name": "My Subscription",
      "resource_groups": {
        "my-rg": {
          "name": "my-rg",
          "location": "eastus",
          "resources": [
            {
              "name": "my-vm",
              "type": "Microsoft.Compute/virtualMachines",
              "id": "/subscriptions/.../my-vm",
              "azure_url": "https://portal.azure.com/#resource/..."
            }
          ],
          "cached_at": "2025-01-15T10:30:00Z"
        }
      },
      "last_updated": "2025-01-15T10:30:00Z"
    }
  },
  "version": "1.0"
}
```

## Error Handling

- Automatically detects if Azure CLI is not installed
- Prompts for login if not authenticated
- Graceful error handling for network issues or permission problems
- Empty resource groups are handled without errors

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling

