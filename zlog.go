package zlog

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type ZLogger interface {
	Context(ctx context.Context, keys []string) ZLogger
	Segment(mainSegment string, detail ...string) ZLogger
	WithError(err error) ZLogger
	Err(err error) ZLogger
	Alert() ZLogger
	WithSource() ZLogger
	WithSourceSkip(skip int) ZLogger
	WithCallStack() ZLogger
	Message(message string)
	Msg(message string)
	Messagef(format string, args ...any)
	Msgf(format string, args ...any)
	Fatal(message string)
	Fatalf(format string, args ...any)
}

type zlogImpl struct {
	logger            *slog.Logger
	attrs             []any
	maxCallStackDepth int
}

// levelConfig holds configuration for a specific log level
type levelConfig struct {
	AutoSource        bool // Automatically add source information
	AutoCallStack     bool // Automatically add call stack information
	MaxCallStackDepth int  // Max call stack depth (0 = use default)
}

// logConfig holds global configuration for automatic features
type logConfig struct {
	Debug levelConfig // Configuration for Debug level (default MaxCallStackDepth: 20)
	Info  levelConfig // Configuration for Info level (default MaxCallStackDepth: 5)
	Warn  levelConfig // Configuration for Warn level (default MaxCallStackDepth: 5)
	Error levelConfig // Configuration for Error level (default MaxCallStackDepth: 10)
}

type Configurable = func(config *logConfig)

func Configure(configs ...Configurable) logConfig {
	conf := logConfig{}
	for _, configFunc := range configs {
		configFunc(&conf)
	}
	return conf
}

func AutoSourceConfig(level slog.Level, autoSource bool) Configurable {
	return func(config *logConfig) {
		switch level {
		case slog.LevelDebug:
			config.Debug.AutoSource = autoSource
		case slog.LevelInfo:
			config.Info.AutoSource = autoSource
		case slog.LevelWarn:
			config.Warn.AutoSource = autoSource
		case slog.LevelError:
			config.Error.AutoSource = autoSource
		}
	}
}

func AutoCallStackConfig(level slog.Level, autoCallStack bool) Configurable {
	return func(config *logConfig) {
		switch level {
		case slog.LevelDebug:
			config.Debug.AutoCallStack = autoCallStack
		case slog.LevelInfo:
			config.Info.AutoCallStack = autoCallStack
		case slog.LevelWarn:
			config.Warn.AutoCallStack = autoCallStack
		case slog.LevelError:
			config.Error.AutoCallStack = autoCallStack
		}
	}
}

func MaxCallStackDepthConfig(level slog.Level, maxDepth int) Configurable {
	return func(config *logConfig) {
		switch level {
		case slog.LevelDebug:
			config.Debug.MaxCallStackDepth = maxDepth
		case slog.LevelInfo:
			config.Info.MaxCallStackDepth = maxDepth
		case slog.LevelWarn:
			config.Warn.MaxCallStackDepth = maxDepth
		case slog.LevelError:
			config.Error.MaxCallStackDepth = maxDepth
		}
	}
}

var (
	debugLogger  *slog.Logger
	infoLogger   *slog.Logger
	warnLogger   *slog.Logger
	errorLogger  *slog.Logger
	globalConfig logConfig

	// Default call stack depths for each log level
	defaultCallStackDepths = map[slog.Level]int{
		slog.LevelDebug: 20,
		slog.LevelInfo:  5,
		slog.LevelWarn:  5,
		slog.LevelError: 10,
	}
)

func init() {
	debugLogger = initNewSlog(slog.LevelDebug)
	infoLogger = initNewSlog(slog.LevelInfo)
	warnLogger = initNewSlog(slog.LevelWarn)
	errorLogger = initNewSlog(slog.LevelError)
}

func initNewSlog(customLevel slog.Level) *slog.Logger {
	replaceAttr := func(groups []string, attr slog.Attr) slog.Attr {
		switch attr.Key {
		case "time":
			return slog.String("time", attr.Value.Time().Format(time.RFC3339))
		case "level":
			return slog.String("level", customLevel.String())
		}
		return attr
	}
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   false,
		ReplaceAttr: replaceAttr,
	})
	return slog.New(jsonHandler)
}

// SetConfig configures global auto-features for all loggers.
// This should be called once in main() to set the desired automatic behaviors.
//
// Example:
//
// zlog.SetConfig(zlog.Configure(
//
//	zlog.AutoSourceConfig(slog.LevelError, true),
//	zlog.AutoCallStackConfig(slog.LevelError, true),
//	zlog.MaxCallStackDepthConfig(slog.LevelError, 8),
//	zlog.AutoSourceConfig(slog.LevelWarn, true),
//	zlog.AutoSourceConfig(slog.LevelInfo, true),
//	zlog.AutoSourceConfig(slog.LevelDebug, true),
//	zlog.AutoCallStackConfig(slog.LevelDebug, true),
//	zlog.MaxCallStackDepthConfig(slog.LevelDebug, 12),
//
// ))
func SetConfig(config logConfig) {
	globalConfig = config
}

// Debug returns a new logger instance at Debug level.
// Debug level is used for detailed troubleshooting and development information.
// The call stack depth is configurable (default: 20) for comprehensive debugging.
//
// Example:
//
//	Debug().Message("Processing item details")
//	// Output: {"level":"debug","time":"2024-03-07T10:00:00Z","message":"Processing item details"}
func Debug() ZLogger {
	level := slog.LevelDebug
	z := &zlogImpl{
		logger:            debugLogger,
		maxCallStackDepth: getMaxCallStackDepth(level),
	}
	return z.applyAutoFeatures(level)
}

// Info returns a new logger instance at Info level.
// Info level is used for general operational entries about what's going on inside the application.
// The call stack depth is configurable (default: 5) for basic tracing.
//
// Example:
//
//	Info().Message("Application started successfully")
//	// Output: {"level":"info","time":"2024-03-07T10:00:00Z","message":"Application started successfully"}
func Info() ZLogger {
	level := slog.LevelInfo
	z := &zlogImpl{
		logger:            infoLogger,
		maxCallStackDepth: getMaxCallStackDepth(level),
	}
	return z.applyAutoFeatures(level)
}

// Warn returns a new logger instance at Warn level.
// Warn level is used for potentially harmful situations and recoverable errors.
// The call stack depth is configurable (default: 5) for basic tracing.
//
// Example:
//
//	Warn().Message("High memory usage detected")
//	// Output: {"level":"warn","time":"2024-03-07T10:00:00Z","message":"High memory usage detected"}
func Warn() ZLogger {
	level := slog.LevelWarn
	z := &zlogImpl{
		logger:            warnLogger,
		maxCallStackDepth: getMaxCallStackDepth(level),
	}
	return z.applyAutoFeatures(level)
}

// Error returns a new logger instance at Error level.
// Error level is used for errors that should be investigated.
// The call stack depth is configurable (default: 10) for detailed error tracing.
//
// Example:
//
//	Error().Error(err).Message("Failed to process request")
//	// Output: {"level":"error","time":"2024-03-07T10:00:00Z","error_msg":"connection refused","message":"Failed to process request"}
func Error() ZLogger {
	level := slog.LevelError
	z := &zlogImpl{
		logger:            errorLogger,
		maxCallStackDepth: getMaxCallStackDepth(level),
	}
	return z.applyAutoFeatures(level)
}

// Panic immediately panics with the given message.
// This should be used only in unrecoverable situations where the application must stop immediately.
//
// Example:
//
//	Panic("Critical configuration missing")
//	// Panics with message: "Critical configuration missing"
func Panic(message string) {
	panic(message)
}

// Panicf immediately panics with the formatted message.
// This should be used only in unrecoverable situations where the application must stop immediately.
//
// Example:
//
//	Panicf("Critical configuration missing: %s", "database credentials")
//	// Panics with message: "Critical configuration missing: database credentials"
func Panicf(format string, args ...any) {
	panic(fmt.Sprintf(format, args...))
}

// Context adds context key-value pairs to the log entry.
// It extracts values from the provided context using the specified keys.
// If a key doesn't exist in the context, it's ignored and the log entry remains unchanged.
//
// Example:
//
//	ctx := context.WithValue(context.Background(), "userID", "12345")
//	ctx = context.WithValue(ctx, "requestID", "req-abc-123")
//	Info().Context(ctx, []string{"userID", "requestID", "nonexistent"}).Message("User action")
//	// Output: {"level":"info","time":"2024-03-07T10:00:00Z","app_ctx":{"userID":"12345","requestID":"req-abc-123"},"message":"User action"}
func (z *zlogImpl) Context(ctx context.Context, keys []string) ZLogger {
	contextMap := make(map[string]any, len(keys))
	for _, key := range keys {
		value := ctx.Value(key)
		if value != nil {
			contextMap[key] = value
		}
	}
	if len(contextMap) == 0 {
		return z
	}
	return z.appendAttr(slog.Any("app_ctx", contextMap))
}

// KeyValue adds a custom key-value pair to the log entry.
// This method is useful for adding arbitrary string metadata to the log.
//
// Example:
//
//	Info().KeyValue("server", "prod-1").KeyValue("region", "eu-west").Message("Server status")
//	// Output: {"level":"info","time":"2024-03-07T10:00:00Z","server":"prod-1","region":"eu-west","message":"Server status"}
func (z *zlogImpl) KeyValue(key, value string) ZLogger {
	return z.appendAttr(slog.String(key, value))
}

// Segment adds a hierarchical path to the log entry.
// Paths help categorize logs by application area, component, or processing stage.
// Multiple detail segments are joined with "/" to create a hierarchical path structure.
//
// Example:
//
//	Info().Segment("api", "users", "create").Message("New user registration")
//	// Output: {"level":"info","time":"2024-03-07T10:00:00Z","segment":"api/users/create","message":"New user registration"}
func (z *zlogImpl) Segment(mainSegment string, detail ...string) ZLogger {
	if len(detail) > 0 {
		mainSegment += "/" + strings.Join(detail, "/")
	}
	return z.appendAttr(slog.String("segment", mainSegment))
}

// WithError adds error information to the log entry.
// It extracts the error message and adds it as 'error_msg' field.
// If the error implements additional interfaces (like stack traces),
// only the Error() string is captured.
//
// Example:
//
//	err := errors.New("connection timeout")
//	Error().WithError(err).Message("Database operation failed")
//	// Output: {"level":"error","time":"2024-03-07T10:00:00Z","error_msg":"connection timeout","message":"Database operation failed"}
func (z *zlogImpl) WithError(err error) ZLogger {
	return z.appendAttr(slog.String("error_msg", err.Error()))
}

// Err is an alias for WithError.
// It provides a shorter method name for convenience.
//
// Example:
//
//	err := errors.New("connection timeout")
//	Error().Err(err).Message("Database operation failed")
//	// Output: {"level":"error","time":"2024-03-07T10:00:00Z","error_msg":"connection timeout","message":"Database operation failed"}
func (z *zlogImpl) Err(err error) ZLogger {
	return z.WithError(err)
}

// WithSource adds the caller's information to the log entry.
// It includes the calling function's name, file path, and line number.
// This is useful for debugging and tracing the exact origin of log messages.
//
// Example:
//
//	Info().WithSource().Message("Processing payment")
//	// Output: {"level":"info","time":"2024-03-07T10:00:00Z","source":"payment.ProcessTransaction @ /app/payment.go:42","message":"Processing payment"}
func (z *zlogImpl) WithSource() ZLogger {
	source, ok := getSourceString(2)
	if !ok {
		return z
	}
	return z.appendAttr(slog.String("source", source))
}

// WithSourceSkip adds the caller's information to the log entry.
// It includes the calling function's name, file path, and line number.
// This is useful for debugging and tracing the exact origin of log messages.
//
// Example:
//
//	Info().WithSourceSkip(3).Message("Processing payment")
//	// Output: {"level":"info","time":"2024-03-07T10:00:00Z","source":"payment.ProcessTransaction @ /app/payment.go:42","message":"Processing payment"}
func (z *zlogImpl) WithSourceSkip(skip int) ZLogger {
	source, ok := getSourceString(2 + skip)
	if !ok {
		return z
	}
	return z.appendAttr(slog.String("source", source))
}

// WithCallStack adds the call stack information to the log entry.
// The depth of the stack trace depends on the logger's maxCallStackDepth setting
// which can be configured globally (defaults: Debug=20, Error=10, Info/Warn=5).
// The trace stops when it reaches the main function or the maximum depth.
//
// Example:
//
//	Error().WithCallStack().Message("Unexpected error")
//	// Output: {"level":"error","time":"2024-03-07T10:00:00Z","callstack":["app.ProcessOrder @ /app/order.go:42","app.HandleRequest @ /app/handler.go:123","main.main @ /app/main.go:15"],"message":"Unexpected error"}
func (z *zlogImpl) WithCallStack() ZLogger {
	callStack := make([]string, 0)
	for skip := 2; skip < z.maxCallStackDepth; skip++ {
		current, ok := getSourceString(skip)
		if !ok {
			continue
		}
		callStack = append(callStack, current)
		if strings.HasPrefix(current, "#main.main") {
			break
		}
	}
	return z.appendAttr(slog.Any("callstack", callStack))
}

// Alert marks the log entry as requiring immediate attention.
// This adds an 'alert' boolean field that can be used for filtering
// or triggering notifications in log management systems.
//
// Example:
//
//	Error().Alert().Message("System running out of disk space")
//	// Output: {"level":"error","time":"2024-03-07T10:00:00Z","alert":true,"message":"System running out of disk space"}
func (z *zlogImpl) Alert() ZLogger {
	return z.appendAttr(slog.Bool("alert", true))
}

// Message emits the log entry with the given message.
// This is a terminal operation that writes the log entry with all accumulated attributes.
// After calling Message, the logger instance should not be reused.
//
// Example:
//
//	Info().KeyValue("status", "healthy").Message("Health check completed")
//	// Output: {"level":"info","time":"2024-03-07T10:00:00Z","status":"healthy","message":"Health check completed"}
func (z *zlogImpl) Message(message string) {
	z.logger.Info(message, z.attrs...)
}

// Msg is an alias for Message.
// It provides a shorter method name for convenience while maintaining the same functionality.
//
// Example:
//
//	Info().KeyValue("status", "healthy").Msg("Health check completed")
//	// Output: {"level":"info","time":"2024-03-07T10:00:00Z","status":"healthy","message":"Health check completed"}
func (z *zlogImpl) Msg(message string) {
	z.logger.Info(message, z.attrs...)
}

// Messagef emits the log entry with a formatted message.
// This is a terminal operation that formats the message using fmt.Sprintf
// and writes the log entry with all accumulated attributes.
//
// Example:
//
//	Info().Messagef("Processed %d items in %v", 100, time.Second*2)
//	// Output: {"level":"info","time":"2024-03-07T10:00:00Z","message":"Processed 100 items in 2s"}
func (z *zlogImpl) Messagef(format string, args ...any) {
	z.logger.Info(fmt.Sprintf(format, args...), z.attrs...)
}

// Msgf is an alias for Messagef.
// It provides a shorter method name for convenience while maintaining the same functionality.
//
// Example:
//
//	Info().Msgf("Processed %d items in %v", 100, time.Second*2)
//	// Output: {"level":"info","time":"2024-03-07T10:00:00Z","message":"Processed 100 items in 2s"}
func (z *zlogImpl) Msgf(format string, args ...any) {
	z.logger.Info(fmt.Sprintf(format, args...), z.attrs...)
}

// Fatal logs the message at error level and then terminates the program with exit code 1.
// This is a terminal operation that should be used only when the application cannot continue running.
// After calling Fatal, the program will exit immediately.
// The method ensures that the log message is written to the output before exiting.
//
// Example:
//
//	Error().Fatal("Failed to initialize database connection")
//	// Output: {"level":"error","time":"2024-03-07T10:00:00Z","message":"Failed to initialize database connection"}
//	// Then exits with status 1
func (z *zlogImpl) Fatal(message string) {
	z.Message(message)
	// Ensure logs are written before exit
	if handler, ok := z.logger.Handler().(interface{ Sync() error }); ok {
		_ = handler.Sync()
	}
	os.Stdout.Sync()
	os.Exit(1)
}

// Fatalf logs the formatted message at error level and then terminates the program with exit code 1.
// This is a terminal operation that should be used only when the application cannot continue running.
// After calling Fatalf, the program will exit immediately.
// The method ensures that the log message is written to the output before exiting.
//
// Example:
//
//	Error().Fatalf("Failed to initialize %s connection", "database")
//	// Output: {"level":"error","time":"2024-03-07T10:00:00Z","message":"Failed to initialize database connection"}
//	// Then exits with status 1
func (z *zlogImpl) Fatalf(format string, args ...any) {
	z.Messagef(format, args...)
	// Ensure logs are written before exit
	if handler, ok := z.logger.Handler().(interface{ Sync() error }); ok {
		_ = handler.Sync()
	}
	os.Stdout.Sync()
	os.Exit(1)
}

func (z *zlogImpl) appendAttr(attr slog.Attr) ZLogger {
	z.attrs = append(z.attrs, attr)
	return z
}

func (z *zlogImpl) appendAttrs(attrs ...any) ZLogger {
	z.attrs = append(z.attrs, attrs...)
	return z
}

// applyAutoFeatures applies automatic features based on global config
func (z *zlogImpl) applyAutoFeatures(level slog.Level) ZLogger {
	var autoSource, autoCallStack bool

	switch level {
	case slog.LevelDebug:
		autoSource = globalConfig.Debug.AutoSource
		autoCallStack = globalConfig.Debug.AutoCallStack
	case slog.LevelInfo:
		autoSource = globalConfig.Info.AutoSource
		autoCallStack = globalConfig.Info.AutoCallStack
	case slog.LevelWarn:
		autoSource = globalConfig.Warn.AutoSource
		autoCallStack = globalConfig.Warn.AutoCallStack
	case slog.LevelError:
		autoSource = globalConfig.Error.AutoSource
		autoCallStack = globalConfig.Error.AutoCallStack
	}

	if autoSource {
		if source, ok := getSourceString(3); ok {
			z.appendAttr(slog.String("source", source))
		}
	}

	if autoCallStack {
		callStack := make([]string, 0)
		for skip := 3; skip < z.maxCallStackDepth; skip++ {
			current, ok := getSourceString(skip)
			if !ok {
				continue
			}
			callStack = append(callStack, current)
			if strings.HasPrefix(current, "#main.main") {
				break
			}
		}
		z.appendAttr(slog.Any("callstack", callStack))
	}

	return z
}

// getMaxCallStackDepth returns the max call stack depth for the given level
// If config value is 0, returns the default value
func getMaxCallStackDepth(level slog.Level) int {
	switch level {
	case slog.LevelDebug:
		if globalConfig.Debug.MaxCallStackDepth > 0 {
			return globalConfig.Debug.MaxCallStackDepth
		}
		return defaultCallStackDepths[level]
	case slog.LevelInfo:
		if globalConfig.Info.MaxCallStackDepth > 0 {
			return globalConfig.Info.MaxCallStackDepth
		}
		return defaultCallStackDepths[level]
	case slog.LevelWarn:
		if globalConfig.Warn.MaxCallStackDepth > 0 {
			return globalConfig.Warn.MaxCallStackDepth
		}
		return defaultCallStackDepths[level]
	case slog.LevelError:
		if globalConfig.Error.MaxCallStackDepth > 0 {
			return globalConfig.Error.MaxCallStackDepth
		}
		return defaultCallStackDepths[level]
	default:
		return 5
	}
}

func getSourceString(skip int) (string, bool) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "", false
	}

	var funcName string
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		funcName = "?"
	} else {
		funcName = fn.Name()
		moduleSeparator := strings.LastIndex(funcName, "/")
		if moduleSeparator != -1 {
			funcName = funcName[moduleSeparator+1:]
		}
	}
	var b strings.Builder
	b.WriteByte('#')
	b.WriteString(funcName)
	b.WriteString(" @ ")
	b.WriteString(file)
	b.WriteByte(':')
	b.WriteString(strconv.FormatInt(int64(line), 10))
	return b.String(), true
}
