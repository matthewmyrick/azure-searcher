# Troubleshooting

This guide helps resolve common issues when using or developing Azure Searcher.

## Installation Issues

### "azure-searcher: command not found"

**Cause**: The binary is not in your PATH.

**Solutions**:
1. **If installed via `go install`**:
   ```bash
   # Add Go bin directory to PATH
   export PATH=$PATH:$(go env GOPATH)/bin
   # Or if GOPATH is not set:
   export PATH=$PATH:$HOME/go/bin
   ```

2. **If downloaded manually**:
   ```bash
   # Move binary to a directory in PATH
   sudo mv azure-searcher /usr/local/bin/
   # Or add current directory to PATH
   export PATH=$PATH:$(pwd)
   ```

### Go version compatibility

**Error**: `go: module requires Go 1.23 or later`

**Solution**: Update Go to version 1.23 or later:
```bash
# Check current version
go version

# Update Go (varies by system)
# macOS with Homebrew:
brew update && brew upgrade go

# Linux: Download from https://golang.org/dl/
```

## Authentication Issues

### "Azure CLI not found" or "az: command not found"

**Cause**: Azure CLI is not installed or not in PATH.

**Solutions**:
1. **Install Azure CLI**:
   ```bash
   # macOS
   brew install azure-cli
   
   # Ubuntu/Debian
   curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
   
   # Windows
   # Download installer from https://docs.microsoft.com/en-us/cli/azure/install-azure-cli
   ```

2. **Verify installation**:
   ```bash
   az --version
   ```

### "Please run 'az login' to setup account"

**Cause**: Not authenticated with Azure.

**Solution**:
```bash
# Interactive login
az login

# Device code login (for remote/headless systems)
az login --use-device-code

# Service principal login (for automation)
az login --service-principal -u <app-id> -p <password> --tenant <tenant-id>
```

### "No subscriptions found"

**Cause**: Account has no accessible subscriptions or wrong tenant.

**Solutions**:
1. **Check available subscriptions**:
   ```bash
   az account list --all
   ```

2. **Set specific subscription**:
   ```bash
   az account set --subscription "subscription-name-or-id"
   ```

3. **Login to specific tenant**:
   ```bash
   az login --tenant <tenant-id>
   ```

## Runtime Issues

### Application hangs on startup

**Possible causes**:
- Slow Azure CLI response
- Network connectivity issues
- Large number of subscriptions

**Solutions**:
1. **Check Azure CLI directly**:
   ```bash
   time az account list
   ```

2. **Test with specific subscription**:
   ```bash
   az account set --subscription "specific-subscription"
   ```

3. **Check network connectivity**:
   ```bash
   az account show
   ```

### "Error fetching subscriptions"

**Debugging steps**:
1. **Test Azure CLI directly**:
   ```bash
   az account list --output table
   ```

2. **Check authentication status**:
   ```bash
   az account show
   ```

3. **Re-authenticate if needed**:
   ```bash
   az logout
   az login
   ```

### Slow resource loading

**Cause**: Large number of resource groups or resources.

**Solutions**:
1. **Use cache**: Data is cached for 30 minutes
2. **Filter subscriptions**: Use fewer subscriptions
3. **Wait for initial load**: Subsequent loads will be faster

### Cache issues

**Symptoms**:
- Stale data shown
- "Error reading cache" messages

**Solutions**:
1. **Clear cache**:
   ```bash
   rm -rf ~/.azure-searcher/
   ```

2. **Force refresh**: Press `Ctrl+R` in the application

3. **Check permissions**:
   ```bash
   ls -la ~/.azure-searcher/
   ```

## UI/Navigation Issues

### Search not working as expected

**Two-part search format**: `<resourcegroup> [space] <resource>`

**Examples**:
- `prod` - Shows resource groups containing "prod"
- `prod vm` - Shows "prod" resource groups, filters for "vm" resources
- `east storage` - Shows "east" resource groups, filters for "storage"

### Keyboard shortcuts not responding

**Common shortcuts**:
- `/` - Focus search
- `Esc` - Clear search or return to subscriptions
- `Ctrl+R` - Refresh data
- `Ctrl+Q` - Quit
- `j/k` or arrow keys - Navigate

**If not working**:
- Ensure terminal supports key combinations
- Try alternative keys (arrow keys vs j/k)
- Check if running in compatible terminal

### Display issues

**Text rendering problems**:
- Use a terminal with Unicode support
- Ensure proper font with icon support
- Try different terminal emulators

## Development Issues

### Build failures

**"Cannot find package"**:
```bash
go mod tidy
go mod download
```

**Version conflicts**:
```bash
go clean -modcache
go mod download
```

### Tests failing

**Run specific tests**:
```bash
go test -v ./src/azure/
go test -v ./src/search/
```

**Clear test cache**:
```bash
go clean -testcache
```

### Import cycle errors

**Cause**: Circular dependencies between packages.

**Solution**: Review package structure and move shared types to common package.

## Performance Issues

### High memory usage

**Causes**:
- Large number of resources
- Memory leaks in goroutines

**Solutions**:
1. **Monitor goroutines**: Add debug logging
2. **Limit concurrent operations**: Already implemented with semaphores
3. **Profile memory usage**:
   ```bash
   go build -race -o azure-searcher
   ```

### Slow search performance

**Causes**:
- Large datasets
- Complex search patterns

**Solutions**:
1. **Use simpler search terms**
2. **Filter by resource group first**
3. **Clear cache if very stale**

## Getting Help

### Logs and Debug Information

**Enable verbose Azure CLI output**:
```bash
az config set core.only_show_errors=false
az config set core.output=table
```

**Check Azure CLI logs**:
```bash
# Location varies by OS
# macOS/Linux: ~/.azure/logs/
# Windows: %USERPROFILE%\.azure\logs\
```

### Reporting Issues

When reporting issues, include:

1. **System information**:
   ```bash
   go version
   az --version
   uname -a  # Linux/macOS
   # or
   systeminfo  # Windows
   ```

2. **Error messages**: Full error output
3. **Steps to reproduce**: Detailed reproduction steps
4. **Expected vs actual behavior**
5. **Configuration**: Subscription count, resource count (approximate)

### Community Support

- **GitHub Issues**: https://github.com/matthewmyrick/azure-searcher/issues
- **Discussions**: For questions and general help
- **Pull Requests**: For bug fixes and improvements

### Quick Fixes Summary

| Issue | Quick Fix |
|-------|-----------|
| Command not found | Add to PATH or reinstall |
| Not authenticated | `az login` |
| No subscriptions | `az account list` then `az account set` |
| Slow loading | Wait for cache, use `Ctrl+R` to refresh |
| Search not working | Use format: `<group> <resource>` |
| Cache issues | Delete `~/.azure-searcher/` |
| Build failures | `go mod tidy && go mod download` |