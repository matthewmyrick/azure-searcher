# Testing Framework for Azure Searcher

This document summarizes the comprehensive testing framework implemented for the Azure Searcher TUI application.

## Overview

A complete testing suite has been implemented covering:
- Unit tests for all core packages
- Integration tests for component interactions
- Test utilities and helpers
- GitHub Actions CI/CD pipeline
- Development tooling (Makefile, linting, etc.)

## Test Coverage

### Unit Tests

**`src/types/types_test.go`**
- JSON serialization/deserialization for all data types
- Validation of struct field handling (tags, resources, etc.)
- Edge cases (empty tags, nil values)

**`src/config/config_test.go`**
- Concurrency configuration based on CPU count
- Default value validation
- Performance testing

**`src/cache/manager_test.go`**
- Cache creation, storage, and retrieval
- TTL (Time To Live) expiration logic
- Configuration management (per-subscription settings)
- Refresh interval parsing and validation
- File I/O operations
- Performance benchmarks

**`src/search/fuzzy_test.go`**
- Fuzzy search algorithms
- Multiple search modes (fuzzy, exact, two-part)
- Scoring and ranking logic
- Resource filtering within groups
- Performance testing

**`src/azure/client_test.go`**
- Azure CLI integration (with mocking approach)
- JSON parsing from Azure CLI output
- URL generation for Azure portal links
- Error handling for various scenarios

**`src/azure/fetcher_test.go`**
- Concurrent resource fetching
- Semaphore-based concurrency control
- Progress tracking
- Resource sorting

**`src/ui/model_test.go`**
- TUI model initialization
- State management
- Component configuration
- Service integration

**`src/testutil/helpers.go`**
- Mock data generators
- Temporary file/directory management
- Test utilities for time handling
- Environment variable management

### Integration Tests

**`integration_test.go`**
- Cache and search integration
- Configuration workflows
- Performance characteristics
- Error handling across components
- Full application workflows

## Test Execution

### Local Development

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Generate coverage report
make test-coverage

# Run performance benchmarks
make bench

# Quick smoke test
make fast-test
```

### CI/CD Pipeline

**`.github/workflows/test.yml`**
- Multi-version Go testing (1.21, 1.22, 1.23)
- Cross-platform builds (Linux, Windows, macOS)
- Code linting and security scanning
- Coverage reporting via Codecov
- Benchmark tracking

## Test Features

### Comprehensive Mocking
- Mock interfaces for external dependencies
- Temporary file systems for I/O testing
- Configurable mock responses

### Performance Testing
- Benchmark tests for critical paths
- Performance regression detection
- Large dataset testing (100+ resource groups, 1000+ resources)

### Error Handling
- Comprehensive error scenario testing
- Edge case validation
- Graceful degradation testing

### Concurrency Testing
- Race condition detection (`-race` flag)
- Semaphore and channel testing
- Atomic operation validation

## Development Tools

### Makefile
Provides convenient commands for common development tasks:
- Testing with different modes
- Code formatting and linting
- Build automation
- Security scanning

### Linting Configuration
**`.golangci.yml`**
- Comprehensive linting rules
- Test-specific exclusions
- Performance and security checks

## Testing Best Practices

1. **Isolation**: Each test runs in isolation with temporary resources
2. **Determinism**: Tests use fixed time values and predictable data
3. **Cleanup**: Automatic cleanup of temporary resources
4. **Coverage**: Comprehensive coverage of success and failure paths
5. **Performance**: Regular performance regression testing
6. **Documentation**: Clear test names and documentation

## Metrics

- **Unit Test Files**: 8
- **Integration Test Files**: 1
- **Total Test Functions**: 80+
- **Code Coverage**: Comprehensive coverage across all packages
- **Performance Tests**: Multiple benchmark functions
- **Test Execution Time**: < 5 seconds for full suite

## Future Enhancements

1. **Property-Based Testing**: Add property-based tests for search algorithms
2. **Visual Testing**: Add tests for TUI rendering
3. **Load Testing**: Add high-load scenario testing
4. **Contract Testing**: Add API contract tests for Azure CLI integration

## Usage Notes

- All tests are designed to run without external dependencies
- Azure CLI integration is tested through structure validation rather than live calls
- Tests are safe to run in CI environments
- Performance tests include timing validations but with reasonable tolerances

The testing framework ensures high code quality, catches regressions early, and provides confidence for refactoring and feature additions.