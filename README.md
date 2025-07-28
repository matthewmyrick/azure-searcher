# Azure Searcher TUI

A fast, interactive Terminal User Interface (TUI) application for browsing Azure resources with advanced parallel processing and search capabilities.

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
- **Search Functionality**: Real-time search filtering of resource groups
- **Collapsible Tree Structure**: Expand/collapse resource groups like folders
- **Visual Icons**: Different icons for resource groups (📁/📂) vs resources (📄)
- **Cache Status Indicators**: Visual indicators showing cached vs live data

## Requirements

- Go 1.19 or later
- Azure CLI (`az`) installed and configured
- Active Azure subscription

## Installation

```bash
git clone <repository-url>
cd azure-searcher
go build -o azure-searcher
```

## Usage

```bash
./azure-searcher
```

## Controls

### Subscription Selection
- **Arrow keys/j,k**: Navigate up/down
- **Enter**: Select subscription
- **q**: Quit

### Resource Navigation
- **Arrow keys/j,k**: Navigate up/down
- **Enter/Space**: Expand/collapse resource groups
- **/**: Focus search bar
- **r**: Refresh data (bypass cache)
- **Esc**: Return to subscription selection
- **q**: Quit

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