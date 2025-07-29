# Development Setup

This guide helps contributors set up their development environment for Azure Searcher.

## Prerequisites

- Go 1.23 or later
- Azure CLI (`az`) installed and configured
- Git
- A GitHub account (for contributions)

## Initial Setup

1. **Fork and Clone**
   ```bash
   # Fork the repository on GitHub first
   git clone https://github.com/YOUR_USERNAME/azure-searcher.git
   cd azure-searcher
   git remote add upstream https://github.com/matthewmyrick/azure-searcher.git
   ```

2. **Install Dependencies**
   ```bash
   go mod download
   ```

3. **Verify Setup**
   ```bash
   go build -o azure-searcher
   ./azure-searcher --help
   ```

## Development Workflow

### Running During Development

```bash
# Run directly
go run main.go

# Build and run
go build -o azure-searcher
./azure-searcher
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./src/search/
```

### Code Quality

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Run golint (if installed)
golint ./...
```

## Project Structure

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
├── docs/                   # Documentation
└── .github/               # GitHub workflows and templates
```

## Key Development Guidelines

### Adding New Features

1. Create a new branch: `git checkout -b feature/your-feature`
2. Add tests for new functionality
3. Update documentation if needed
4. Follow Go coding standards
5. Test manually with Azure CLI

### Debugging Tips

- Use `fmt.Printf` for quick debugging
- Test with small Azure subscriptions first
- Enable verbose logging in Azure CLI: `az config set core.only_show_errors=false`

### UI Development

The UI is built using [Bubble Tea](https://github.com/charmbracelet/bubbletea). Key concepts:

- **Model**: Application state
- **Update**: Handle messages and update state
- **View**: Render the current state
- **Commands**: Asynchronous operations

### Azure Integration

- All Azure operations go through `src/azure/client.go`
- Use the existing patterns for new Azure CLI calls
- Always handle errors gracefully
- Consider caching for performance

## Common Development Tasks

### Adding a New Search Feature

1. Implement logic in `src/search/`
2. Add tests
3. Integrate with UI in `src/ui/`
4. Update documentation

### Modifying UI Behavior

1. Update the model in `src/ui/model.go`
2. Handle new messages in `src/ui/update.go`
3. Update views in `src/ui/views.go`
4. Test keyboard interactions

### Adding New Configuration

1. Add to `src/config/config.go`
2. Update documentation
3. Consider environment variable support

## Troubleshooting

### Common Issues

**"Azure CLI not found"**
- Ensure `az` command is in your PATH
- Install Azure CLI from https://docs.microsoft.com/en-us/cli/azure/

**"No subscriptions found"**
- Run `az login` to authenticate
- Check `az account list` to see available subscriptions

**Build failures**
- Ensure Go 1.23+ is installed
- Run `go mod tidy` to clean up dependencies

### Getting Help

- Check existing issues on GitHub
- Read the main README.md
- Ask questions in new issues with the "question" label