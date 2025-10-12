# zlog 🪶

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-%2300ADD8)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**Elegant structured logging for Go** - A lightweight wrapper around Go's standard `log/slog` with a fluent API and automatic context enrichment.

## ✨ Features

- 🎯 **Fluent API** - Chain methods for readable, intuitive logging
- ⚙️ **Automatic Source Tracking** - Configurable per-level source file and line numbers
- 📚 **Smart Call Stacks** - Automatic stack traces with configurable depth
- 🔍 **Context Integration** - Extract and log context values effortlessly
- 🏗️ **Hierarchical Segments** - Organize logs by component/feature paths
- 📁 **JSON Configuration** - Load settings from external files for easy management
- 📝 **Flexible Output** - Write to files, network, or any io.Writer destination
- 🚀 **Zero Dependencies** - Built on Go's standard library
- 📦 **JSON Output** - Structured logs ready for aggregation tools

## 📦 Installation

```bash
go get github.com/GokselKUCUKSAHIN/zlog
```

## 🚀 Quick Start

```go
package main

import (
    "github.com/GokselKUCUKSAHIN/zlog"
)

func main() {
    // Simple logging
    zlog.Info().Message("Application started")
    
    // With context
    zlog.Error().
        Err(err).
        Segment("payment", "process").
        Message("Payment failed")
}
```

**Output:**
```json
{"time":"2024-03-07T10:00:00Z","level":"INFO","msg":"Application started"}
{"time":"2024-03-07T10:00:00Z","level":"ERROR","segment":"payment/process","error_msg":"connection timeout","msg":"Payment failed"}
```

## 📖 Usage

### Log Levels

```go
zlog.Debug().Message("Detailed debug information")
zlog.Info().Message("General information")
zlog.Warn().Message("Warning message")
zlog.Error().Err(err).Message("Error occurred")
```

### Global Configuration

Configure automatic features once at startup:

```go
import "log/slog"

zlog.SetConfig(zlog.Configure(
    // Automatically add source info for errors
    zlog.AutoSourceConfig(slog.LevelError, true),
    zlog.AutoCallStackConfig(slog.LevelError, true),
    zlog.MaxCallStackDepthConfig(slog.LevelError, 10),
    
    // Add source info for warnings
    zlog.AutoSourceConfig(slog.LevelWarn, true),
    
    // Debug with deep call stacks
    zlog.AutoSourceConfig(slog.LevelDebug, true),
    zlog.AutoCallStackConfig(slog.LevelDebug, true),
    zlog.MaxCallStackDepthConfig(slog.LevelDebug, 20),
))
```

**Before configuration:**
```json
{"level":"ERROR","msg":"Database error","error_msg":"connection refused"}
```

**After configuration:**
```json
{"level":"ERROR","msg":"Database error","source":"#main.processOrder @ /app/order.go:42","callstack":["#main.processOrder @ /app/order.go:42","#main.main @ /app/main.go:15"],"error_msg":"connection refused"}
```

### Configuration from JSON File

Load configuration from a JSON file for easier management:

```go
// Load config from file
zlog.SetConfig(zlog.ConfigureFromJSONFile("log-config.json"))
```

**Example JSON configuration file:**
```json
{
    "debug": {
        "autoSource": true,
        "autoCallStack": true,
        "maxCallStackDepth": 20
    },
    "info": {
        "autoSource": true,
        "autoCallStack": false,
        "maxCallStackDepth": 5
    },
    "warn": {
        "autoSource": true,
        "autoCallStack": false,
        "maxCallStackDepth": 5
    },
    "error": {
        "autoSource": true,
        "autoCallStack": true,
        "maxCallStackDepth": 10
    }
}
```

This approach allows you to:
- Change configuration without recompiling
- Maintain different configs for different environments
- Share configuration across team members

### Custom Output Writer

Redirect log output to files, network connections, or any `io.Writer`:

```go
// Write to a file
file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
if err != nil {
    log.Fatal(err)
}
defer file.Close()
zlog.SetOutputWriter(file)

// Write to multiple destinations
multiWriter := io.MultiWriter(os.Stdout, file)
zlog.SetOutputWriter(multiWriter)

// Write to a buffer for testing
var buf bytes.Buffer
zlog.SetOutputWriter(&buf)
```

By default, logs are written to `os.Stdout`.

### Context Values

Extract and log specific context keys:

```go
ctx := context.WithValue(context.Background(), "userID", "12345")
ctx = context.WithValue(ctx, "requestID", "req-abc-123")

zlog.Info().
    Context(ctx, []string{"userID", "requestID"}).
    Message("User action completed")
```

**Output:**
```json
{"level":"INFO","app_ctx":{"userID":"12345","requestID":"req-abc-123"},"msg":"User action completed"}
```

### Hierarchical Segments

Organize logs by component or feature:

```go
zlog.Info().Segment("api", "users", "create").Message("User created")
zlog.Error().Segment("database", "orders").Err(err).Message("Query failed")
```

**Output:**
```json
{"level":"INFO","segment":"api/users/create","msg":"User created"}
{"level":"ERROR","segment":"database/orders","error_msg":"timeout","msg":"Query failed"}
```

### Manual Source and Call Stacks

Override automatic configuration when needed:

```go
// Add source manually
zlog.Info().WithSource().Message("Manual source tracking")

// Add full call stack
zlog.Error().WithCallStack().Err(err).Message("Critical error")
```

### Formatted Messages

```go
zlog.Info().Messagef("Processed %d items in %v", count, duration)
// Alias: Msgf()
zlog.Error().Msgf("Failed after %d retries", maxRetries)
```

### Alert Flag

Mark critical logs for monitoring systems:

```go
zlog.Error().
    Alert().
    Err(err).
    Message("Disk space critically low")
```

**Output:**
```json
{"level":"ERROR","alert":true,"error_msg":"less than 5% available","msg":"Disk space critically low"}
```

### Fatal Logging

Log and exit with status code 1:

```go
zlog.Error().Fatal("Critical configuration missing")
zlog.Error().Fatalf("Cannot start without %s", requiredConfig)
```

### Panic

Immediately panic (use sparingly):

```go
zlog.Panic("Unrecoverable state detected")
zlog.Panicf("Invalid state: %s", state)
```

## 🎯 Real-World Example

```go
package main

import (
    "context"
    "errors"
    "log/slog"
    "github.com/GokselKUCUKSAHIN/zlog"
)

func main() {
    // Configure automatic features
    zlog.SetConfig(zlog.Configure(
        zlog.AutoSourceConfig(slog.LevelError, true),
        zlog.AutoCallStackConfig(slog.LevelError, true),
    ))
    
    // Create request context
    ctx := context.WithValue(context.Background(), "requestID", "req-123")
    ctx = context.WithValue(ctx, "userID", "user-456")
    
    // Process order
    if err := processOrder(ctx, "order-789"); err != nil {
        zlog.Error().
            Context(ctx, []string{"requestID", "userID"}).
            Segment("orders", "process").
            Err(err).
            Messagef("Order processing failed: %s", "order-789")
    }
    
    // Success log
    zlog.Info().
        Context(ctx, []string{"requestID", "userID"}).
        Segment("orders", "process").
        Message("Order completed successfully")
}

func processOrder(ctx context.Context, orderID string) error {
    return errors.New("payment gateway timeout")
}
```

**Output:**
```json
{"time":"2024-03-07T10:00:00Z","level":"ERROR","source":"#main.main @ /app/main.go:25","callstack":["#main.main @ /app/main.go:25"],"app_ctx":{"requestID":"req-123","userID":"user-456"},"segment":"orders/process","error_msg":"payment gateway timeout","msg":"Order processing failed: order-789"}
{"time":"2024-03-07T10:00:00Z","level":"INFO","app_ctx":{"requestID":"req-123","userID":"user-456"},"segment":"orders/process","msg":"Order completed successfully"}
```

## 🔧 API Reference

### Log Levels
- `Debug()` - Detailed debugging information
- `Info()` - General operational messages
- `Warn()` - Warning messages
- `Error()` - Error messages

### Chaining Methods
- `Context(ctx, keys)` - Add context values
- `Segment(main, details...)` - Add hierarchical path
- `WithError(err)` / `Err(err)` - Add error message
- `WithSource()` - Add caller information
- `WithCallStack()` - Add full call stack
- `Alert()` - Mark as alert

### Terminal Methods
- `Message(msg)` / `Msg(msg)` - Emit log
- `Messagef(fmt, args...)` / `Msgf(fmt, args...)` - Emit formatted log
- `Fatal(msg)` / `Fatalf(fmt, args...)` - Log and exit(1)

### Global Functions
- `SetConfig(config)` - Configure automatic features
- `Configure(configs...)` - Create configuration
- `ConfigureFromJSONFile(path)` - Load configuration from JSON file
- `SetOutputWriter(writer)` - Set custom output destination (io.Writer)
- `AutoSourceConfig(level, enabled)` - Auto-add source
- `AutoCallStackConfig(level, enabled)` - Auto-add stack
- `MaxCallStackDepthConfig(level, depth)` - Set stack depth
- `Panic(msg)` / `Panicf(fmt, args...)` - Panic immediately

## 🎨 Configuration Patterns

### Production
```go
zlog.SetConfig(zlog.Configure(
    zlog.AutoSourceConfig(slog.LevelError, true),
    zlog.AutoCallStackConfig(slog.LevelError, true),
))
```

### Development
```go
zlog.SetConfig(zlog.Configure(
    zlog.AutoSourceConfig(slog.LevelDebug, true),
    zlog.AutoSourceConfig(slog.LevelInfo, true),
    zlog.AutoSourceConfig(slog.LevelWarn, true),
    zlog.AutoSourceConfig(slog.LevelError, true),
    zlog.AutoCallStackConfig(slog.LevelError, true),
    zlog.AutoCallStackConfig(slog.LevelDebug, true),
))
```

### Environment-Based with JSON
```go
// Load different configs based on environment
configFile := os.Getenv("LOG_CONFIG_PATH")
if configFile == "" {
    configFile = "log-config.json" // default
}
zlog.SetConfig(zlog.ConfigureFromJSONFile(configFile))
```

### File-Based Logging
```go
// Write logs to a file with rotation support
file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
if err != nil {
    log.Fatal(err)
}
defer file.Close()

// Send logs to both console and file
multiWriter := io.MultiWriter(os.Stdout, file)
zlog.SetOutputWriter(multiWriter)

zlog.SetConfig(zlog.Configure(
    zlog.AutoSourceConfig(slog.LevelError, true),
    zlog.AutoCallStackConfig(slog.LevelError, true),
))
```

### Performance-Critical
```go
// No automatic features - minimal overhead
// Only add source/stack when explicitly needed with WithSource()/WithCallStack()
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

Built on top of Go's excellent `log/slog` package.

---

Made with ❤️ by [Göksel Küçükşahin](https://github.com/GokselKUCUKSAHIN)
