# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Open source contribution documentation
- Comprehensive docs/ folder with development guides
- Issue templates for better bug reporting
- Pull request template

### Changed
- Updated release process to use modern GitHub Actions
- Improved GoReleaser configuration

## [1.1.0] - 2025-01-XX

### Added
- Automated release process with GoReleaser
- Cross-platform binary builds
- Improved CI/CD workflows

### Fixed
- Release process improvements

## [1.0.0] - 2025-01-XX

### Added
- Initial release of Azure Searcher TUI
- Azure CLI integration with automatic login detection
- Intelligent caching system with 30-minute TTL
- High-performance parallel processing with goroutines
- Interactive Terminal User Interface with keyboard navigation
- Animated loading with spinner and progress bar
- Two-part search functionality (`<resourcegroup> <resource>`)
- Collapsible tree structure for resource groups
- Visual icons for different resource types
- Cache status indicators
- Support for multiple Azure subscriptions
- CPU-based concurrency optimization
- JSON file-based cache storage
- Fuzzy search capabilities
- Real-time search filtering
- Azure portal URL generation for resources

### Technical Features
- Built with Bubble Tea TUI framework
- Controlled concurrency with semaphores
- Progress tracking via channels
- Cache-first data strategy
- Error handling for Azure CLI operations
- Cross-platform support (Linux, macOS, Windows)

## Format

- `Added` for new features
- `Changed` for changes in existing functionality
- `Deprecated` for soon-to-be removed features
- `Removed` for now removed features
- `Fixed` for any bug fixes
- `Security` for vulnerability fixes