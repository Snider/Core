# Module System

The Module system (`pkg/module`) provides a declarative way to register UI menus, routes, and API endpoints using the `.itw3.json` configuration format.

## Features

- Declarative module configuration
- UI menu contributions
- Frontend route registration
- API endpoint declarations
- Multi-context support (developer, retail, miner)
- Binary/daemon management
- Module dependencies

## Module Config Format

Modules are defined using `.itw3.json` files:

```json
{
    "code": "wallet",
    "type": "core",
    "name": "Wallet Manager",
    "version": "1.0.0",
    "namespace": "finance",
    "description": "Cryptocurrency wallet management",
    "author": "Your Name",
    "contexts": ["default", "retail"],
    "menu": [...],
    "routes": [...],
    "api": [...],
    "config": {...}
}
```

## Module Types

| Type | Description |
|------|-------------|
| `core` | Built-in core functionality |
| `app` | External web application |
| `bin` | Binary/daemon wrapper |

## UI Contexts

Modules can target specific UI contexts:

| Context | Description |
|---------|-------------|
| `default` | Standard user interface |
| `developer` | Developer tools and debugging |
| `retail` | Point-of-sale interface |
| `miner` | Mining operations interface |

## Menu Contributions

Add items to the application menu:

```json
{
    "menu": [
        {
            "id": "wallet-send",
            "label": "Send Funds",
            "icon": "send",
            "route": "/wallet/send",
            "accelerator": "CmdOrCtrl+Shift+S",
            "contexts": ["default", "retail"],
            "order": 10
        },
        {
            "id": "wallet-receive",
            "label": "Receive",
            "icon": "receive",
            "route": "/wallet/receive",
            "order": 20
        },
        {
            "separator": true
        },
        {
            "id": "wallet-settings",
            "label": "Settings",
            "action": "wallet.open_settings",
            "children": [
                {"id": "wallet-backup", "label": "Backup", "action": "wallet.backup"},
                {"id": "wallet-restore", "label": "Restore", "action": "wallet.restore"}
            ]
        }
    ]
}
```

## Route Contributions

Register frontend routes:

```json
{
    "routes": [
        {
            "path": "/wallet",
            "component": "wallet-dashboard",
            "title": "Wallet",
            "icon": "wallet",
            "contexts": ["default"]
        },
        {
            "path": "/wallet/send",
            "component": "wallet-send-form",
            "title": "Send Funds"
        }
    ]
}
```

## API Declarations

Declare API endpoints the module provides:

```json
{
    "api": [
        {
            "method": "GET",
            "path": "/balance",
            "description": "Get wallet balance"
        },
        {
            "method": "POST",
            "path": "/send",
            "description": "Send transaction"
        }
    ]
}
```

## Binary Downloads

For `bin` type modules, specify platform binaries:

```json
{
    "downloads": {
        "app": "https://example.com/wallet-ui.tar.gz",
        "x86_64": {
            "darwin": {
                "url": "https://example.com/wallet-darwin-x64",
                "checksum": "sha256:abc123..."
            },
            "linux": {
                "url": "https://example.com/wallet-linux-x64",
                "checksum": "sha256:def456..."
            },
            "windows": {
                "url": "https://example.com/wallet-win-x64.exe",
                "checksum": "sha256:ghi789..."
            }
        },
        "aarch64": {
            "darwin": {
                "url": "https://example.com/wallet-darwin-arm64"
            }
        }
    }
}
```

## Web App Configuration

For `app` type modules:

```json
{
    "app": {
        "url": "https://example.com/wallet-app.tar.gz",
        "type": "spa",
        "hooks": [
            {
                "type": "rename",
                "from": "dist",
                "to": "wallet"
            }
        ]
    }
}
```

## Dependencies

Declare module dependencies:

```json
{
    "depends": ["core", "crypto"]
}
```

## Using in Go

### Module Registration

```go
import "github.com/Snider/Core/pkg/module"

// Create from config
cfg := module.Config{
    Code:      "wallet",
    Type:      module.TypeCore,
    Name:      "Wallet",
    Namespace: "finance",
}

mod := module.Module{
    Config:  cfg,
    Handler: myHandler,
}
```

### Gin Router Integration

```go
type WalletModule struct{}

func (m *WalletModule) RegisterRoutes(group *gin.RouterGroup) {
    group.GET("/balance", m.getBalance)
    group.POST("/send", m.sendTransaction)
}

// Register with Gin
router := gin.Default()
apiGroup := router.Group("/api/finance/wallet")
walletModule.RegisterRoutes(apiGroup)
```

## Registry Service

The registry manages all modules:

```go
import "github.com/Snider/Core/pkg/module"

registry := module.NewRegistry()

// Register module
registry.Register(walletModule)

// Get module by code
mod := registry.Get("wallet")

// List all modules
modules := registry.List()

// Get modules for context
devModules := registry.ForContext(module.ContextDeveloper)
```

## Built-in Modules

Core provides several built-in modules:

- System information
- Configuration management
- Process management
- File operations

Access via:

```go
builtins := module.BuiltinModules()
```
