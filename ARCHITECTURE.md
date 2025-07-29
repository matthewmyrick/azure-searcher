# Azure Searcher - Architecture Overview

This document describes the refactored architecture of the Azure Searcher TUI application.

## Directory Structure

```
azure-searcher/
├── main.go                 # Application entry point
├── src/                    # Source code organized by domain
│   ├── types/             # Core data structures and types
│   │   └── types.go       # Subscription, Resource, ResourceGroup, etc.
│   ├── config/            # Configuration and constants
│   │   └── config.go      # Concurrency limits, cache settings
│   ├── azure/             # Azure CLI integration
│   │   ├── client.go      # Azure CLI operations
│   │   └── fetcher.go     # Concurrent resource fetching
│   ├── cache/             # Caching functionality
│   │   └── manager.go     # JSON file-based cache management
│   └── ui/                # Terminal User Interface
│       ├── model.go       # TUI model and initialization
│       ├── update.go      # Event handling and state updates
│       ├── views.go       # View rendering
│       ├── commands.go    # Bubble Tea commands
│       ├── messages.go    # Message types
│       └── styles.go      # UI styling constants
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
└── README.md              # User documentation
```

## Package Responsibilities

### `/src/types`
- **Purpose**: Core data structures used throughout the application
- **Key Types**:
  - `Subscription`: Azure subscription information
  - `Resource`: Individual Azure resource
  - `ResourceGroup`: Collection of resources
  - `CacheData`: Cache file structure
  - `ProgressUpdate`: Progress tracking during fetching

### `/src/config`
- **Purpose**: Application configuration and constants
- **Features**:
  - CPU-based concurrency optimization
  - Cache TTL settings
  - Default values for various settings

### `/src/azure`
- **Purpose**: Azure CLI integration and resource fetching
- **Key Components**:
  - `Client`: Basic Azure CLI operations (login, subscriptions, etc.)
  - `Fetcher`: Concurrent resource fetching with progress tracking
- **Features**:
  - Controlled concurrency with semaphores
  - Progress reporting via channels
  - Error handling and retry logic

### `/src/cache`
- **Purpose**: Local caching system
- **Features**:
  - JSON file-based storage
  - TTL-based cache validation
  - Per-subscription cache management
  - Automatic cache directory creation

### `/src/ui`
- **Purpose**: Terminal User Interface using Bubble Tea
- **Key Components**:
  - `Model`: Main application state
  - `Update`: Event handling and state transitions
  - `Views`: Screen rendering for different states
  - `Commands`: Asynchronous operations
  - `Messages`: Inter-component communication
  - `Styles`: Consistent UI theming

## Data Flow

1. **Initialization**: `main.go` → `ui.NewModel()` → Initialize all services
2. **Subscription Loading**: `ui.InitCmd()` → `azure.Client.GetSubscriptions()` → Display list
3. **Resource Group Fetching**: 
   - Check `cache.Manager` for cached data
   - If cache miss: `azure.Fetcher.FetchResourceGroups()` with progress tracking
   - Save results to cache
4. **UI Updates**: Progress updates flow through channels to update progress bars
5. **User Interaction**: Keyboard events handled in `ui.Update()` with state transitions

## Key Design Patterns

### Dependency Injection
- Services (Azure client, cache manager, fetcher) are injected into the UI model
- Makes testing easier and reduces coupling

### Channel-Based Progress Tracking
- Goroutines report progress via typed channels
- UI listens for updates and renders progress bars in real-time

### State Machine Pattern
- Application has clear states: "subscriptions", "loading", "resources"
- State transitions are explicit and controlled

### Semaphore Pattern for Concurrency Control
- Buffered channels act as semaphores to limit concurrent operations
- Prevents system overload while maintaining good performance

### Cache-First Strategy
- Always check cache before making Azure CLI calls
- Transparent to the user - they see the same interface regardless of data source

## Benefits of This Architecture

1. **Maintainability**: Clear separation of concerns makes code easier to understand and modify
2. **Testability**: Each package can be tested independently
3. **Performance**: Controlled concurrency and intelligent caching
4. **Extensibility**: Easy to add new features or change implementations
5. **Reusability**: Core packages can be used in other applications

## Future Enhancements

- **Pluggable Backends**: Abstract Azure CLI behind an interface for easier testing
- **Configuration File**: Allow users to customize concurrency limits and cache settings
- **Metrics Collection**: Track usage patterns and performance metrics
- **Resource Details**: Fetch additional resource information on demand
- **Export Functionality**: Export resource lists to various formats