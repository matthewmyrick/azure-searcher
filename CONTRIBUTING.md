# Contributing to Azure Searcher

Thank you for your interest in contributing to Azure Searcher! This document provides guidelines and instructions for contributing to the project.

## Table of Contents
- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Issue Guidelines](#issue-guidelines)

## Code of Conduct

This project adheres to a Code of Conduct. By participating, you are expected to uphold this code. Please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for details.

## Getting Started

### Prerequisites
- Go 1.23 or later
- Azure CLI (`az`) installed and configured
- Git
- A GitHub account

### Development Setup

1. **Fork the Repository**
   ```bash
   # Click the "Fork" button on GitHub, then clone your fork
   git clone https://github.com/YOUR_USERNAME/azure-searcher.git
   cd azure-searcher
   ```

2. **Add Upstream Remote**
   ```bash
   git remote add upstream https://github.com/matthewmyrick/azure-searcher.git
   ```

3. **Install Dependencies**
   ```bash
   go mod download
   ```

4. **Build and Test**
   ```bash
   go build -o azure-searcher
   ./azure-searcher
   ```

## How to Contribute

### Types of Contributions

We welcome several types of contributions:

- **Bug Reports**: Help us identify and fix issues
- **Feature Requests**: Suggest new functionality
- **Code Contributions**: Fix bugs or implement features
- **Documentation**: Improve or add documentation
- **Testing**: Add or improve tests

### Before You Start

1. **Check Existing Issues**: Look through existing issues and PRs to avoid duplication
2. **Create an Issue**: For significant changes, create an issue first to discuss the approach
3. **Start Small**: If you're new to the project, start with small changes or good first issues

## Pull Request Process

### 1. Create a Branch
```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### 2. Make Your Changes
- Follow the coding standards (see below)
- Add tests for new functionality
- Update documentation as needed
- Keep commits focused and atomic

### 3. Test Your Changes
```bash
# Build the project
go build -o azure-searcher

# Run any existing tests
go test ./...

# Test manually with Azure CLI
./azure-searcher
```

### 4. Commit Your Changes
```bash
git add .
git commit -m "feat: add new search feature"
```

Follow [Conventional Commits](https://conventionalcommits.org/) format:
- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `test:` for adding tests
- `refactor:` for code refactoring

### 5. Push and Create PR
```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub using the provided template.

## Coding Standards

### Go Style Guide
- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Use meaningful variable and function names
- Add comments for exported functions and complex logic

### Project Structure
The project follows this structure:
```
azure-searcher/
├── main.go                 # Application entry point
├── src/                    # Source code organized by domain
│   ├── types/             # Core data structures
│   ├── config/            # Configuration
│   ├── azure/             # Azure CLI integration
│   ├── cache/             # Caching functionality
│   ├── search/            # Search algorithms
│   └── ui/                # Terminal User Interface
```

### Key Principles
- **Separation of Concerns**: Keep different functionalities in separate packages
- **Error Handling**: Always handle errors appropriately
- **Concurrency**: Use goroutines and channels responsibly
- **Performance**: Consider performance implications, especially for UI responsiveness

## Testing Guidelines

### Writing Tests
- Write unit tests for new functions and methods
- Test both happy paths and error conditions
- Use table-driven tests for multiple test cases
- Mock external dependencies (Azure CLI calls)

### Test Organization
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./src/search/
```

### Manual Testing
- Test with different Azure subscriptions
- Test search functionality with various inputs
- Test caching behavior
- Test keyboard navigation and UI responsiveness

## Issue Guidelines

### Bug Reports
When reporting bugs, please include:
- **Description**: Clear description of the issue
- **Steps to Reproduce**: Detailed steps to reproduce the bug
- **Expected Behavior**: What should happen
- **Actual Behavior**: What actually happens
- **Environment**: OS, Go version, Azure CLI version
- **Screenshots/Logs**: If applicable

### Feature Requests
When requesting features, please include:
- **Description**: Clear description of the feature
- **Use Case**: Why this feature would be useful
- **Proposed Solution**: If you have ideas on implementation
- **Alternatives**: Any alternative solutions you've considered

### Labels
We use labels to categorize issues:
- `bug`: Something isn't working
- `enhancement`: New feature or request
- `documentation`: Improvements or additions to documentation
- `good first issue`: Good for newcomers
- `help wanted`: Extra attention is needed

## Development Tips

### Architecture Overview
Read [ARCHITECTURE.md](ARCHITECTURE.md) to understand the project structure and design patterns.

### Key Components
- **UI Layer**: Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Azure Integration**: Uses Azure CLI commands
- **Caching**: JSON file-based caching system
- **Search**: Fuzzy search with two-part filtering
- **Concurrency**: Controlled parallel processing with semaphores

### Debugging
- Use `fmt.Printf` or logging for debugging
- Test with small Azure subscriptions first
- Use `go run main.go` for quick testing during development

## Questions?

If you have questions about contributing, feel free to:
- Open an issue with the `question` label
- Reach out to the maintainers
- Check existing documentation

Thank you for contributing to Azure Searcher! 🚀