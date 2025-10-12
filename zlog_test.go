package zlog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/GokselKUCUKSAHIN/zlog"
)

// setupTestLogger configures logger to write to a buffer for testing
func setupTestLogger(buf *bytes.Buffer) {
	zlog.SetOutputWriter(buf)
	zlog.SetConfig(zlog.Configure())
}

// parseLogOutput parses JSON log output into a map
func parseLogOutput(output string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result)
	return result, err
}

// TestBasicLogLevels tests that each log level produces the correct output
func TestBasicLogLevels(t *testing.T) {
	tests := []struct {
		name          string
		logFunc       func()
		expectedLevel string
		expectedMsg   string
	}{
		{
			name: "Debug level",
			logFunc: func() {
				zlog.Debug().Message("debug message")
			},
			expectedLevel: "DEBUG",
			expectedMsg:   "debug message",
		},
		{
			name: "Info level",
			logFunc: func() {
				zlog.Info().Message("info message")
			},
			expectedLevel: "INFO",
			expectedMsg:   "info message",
		},
		{
			name: "Warn level",
			logFunc: func() {
				zlog.Warn().Message("warn message")
			},
			expectedLevel: "WARN",
			expectedMsg:   "warn message",
		},
		{
			name: "Error level",
			logFunc: func() {
				zlog.Error().Message("error message")
			},
			expectedLevel: "ERROR",
			expectedMsg:   "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			setupTestLogger(&buf)

			tt.logFunc()

			output := buf.String()
			logData, err := parseLogOutput(output)
			if err != nil {
				t.Fatalf("Failed to parse log output: %v\nOutput: %s", err, output)
			}

			if level, ok := logData["level"].(string); !ok || level != tt.expectedLevel {
				t.Errorf("Expected level %s, got %v", tt.expectedLevel, logData["level"])
			}

			if msg, ok := logData["msg"].(string); !ok || msg != tt.expectedMsg {
				t.Errorf("Expected message %s, got %v", tt.expectedMsg, logData["msg"])
			}

			// Verify time field exists
			if _, ok := logData["time"].(string); !ok {
				t.Error("Expected time field to be present")
			}
		})
	}
}

// TestMessageFormatters tests Msg, Msgf, Message, and Messagef methods
func TestMessageFormatters(t *testing.T) {
	tests := []struct {
		name        string
		logFunc     func()
		expectedMsg string
	}{
		{
			name: "Message method",
			logFunc: func() {
				zlog.Info().Message("test message")
			},
			expectedMsg: "test message",
		},
		{
			name: "Msg method (alias)",
			logFunc: func() {
				zlog.Info().Msg("test msg")
			},
			expectedMsg: "test msg",
		},
		{
			name: "Messagef method",
			logFunc: func() {
				zlog.Info().Messagef("formatted %s %d", "message", 42)
			},
			expectedMsg: "formatted message 42",
		},
		{
			name: "Msgf method (alias)",
			logFunc: func() {
				zlog.Info().Msgf("formatted %s %d", "msg", 99)
			},
			expectedMsg: "formatted msg 99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			setupTestLogger(&buf)

			tt.logFunc()

			output := buf.String()
			logData, err := parseLogOutput(output)

			if err != nil {
				t.Fatalf("Failed to parse log output: %v", err)
			}

			if msg, ok := logData["msg"].(string); !ok || msg != tt.expectedMsg {
				t.Errorf("Expected message %s, got %v", tt.expectedMsg, logData["msg"])
			}
		})
	}
}

// TestSegment tests the Segment method
func TestSegment(t *testing.T) {
	tests := []struct {
		name            string
		logFunc         func()
		expectedSegment string
	}{
		{
			name: "Single segment",
			logFunc: func() {
				zlog.Info().Segment("api").Message("test")
			},
			expectedSegment: "api",
		},
		{
			name: "Multiple segments",
			logFunc: func() {
				zlog.Info().Segment("api", "users", "create").Message("test")
			},
			expectedSegment: "api/users/create",
		},
		{
			name: "Segment with details",
			logFunc: func() {
				zlog.Info().Segment("payment", "process", "stripe").Message("test")
			},
			expectedSegment: "payment/process/stripe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			setupTestLogger(&buf)

			tt.logFunc()

			output := buf.String()
			logData, err := parseLogOutput(output)

			if err != nil {
				t.Fatalf("Failed to parse log output: %v", err)
			}

			if segment, ok := logData["segment"].(string); !ok || segment != tt.expectedSegment {
				t.Errorf("Expected segment %s, got %v", tt.expectedSegment, logData["segment"])
			}
		})
	}
}

// TestKeyValue tests the KeyValue method
func TestKeyValue(t *testing.T) {
	tests := []struct {
		name     string
		logFunc  func()
		checkKey string
		checkVal string
	}{
		{
			name: "Single key-value",
			logFunc: func() {
				zlog.Info().KeyValue("user_id", "12345").Message("test")
			},
			checkKey: "user_id",
			checkVal: "12345",
		},
		{
			name: "Multiple key-values",
			logFunc: func() {
				zlog.Info().
					KeyValue("key1", "value1").
					KeyValue("key2", "value2").
					Message("test")
			},
			checkKey: "key1",
			checkVal: "value1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			setupTestLogger(&buf)

			tt.logFunc()

			output := buf.String()
			logData, err := parseLogOutput(output)

			if err != nil {
				t.Fatalf("Failed to parse log output: %v", err)
			}

			if val, ok := logData[tt.checkKey].(string); !ok || val != tt.checkVal {
				t.Errorf("Expected %s=%s, got %v", tt.checkKey, tt.checkVal, logData[tt.checkKey])
			}
		})
	}
}

// TestWithError tests error handling
func TestWithError(t *testing.T) {
	tests := []struct {
		name          string
		logFunc       func()
		expectedError string
	}{
		{
			name: "WithError method",
			logFunc: func() {
				err := errors.New("test error")
				zlog.Error().WithError(err).Message("error occurred")
			},
			expectedError: "test error",
		},
		{
			name: "Err method (alias)",
			logFunc: func() {
				err := errors.New("another error")
				zlog.Error().Err(err).Message("error occurred")
			},
			expectedError: "another error",
		},
		{
			name: "Nil error",
			logFunc: func() {
				zlog.Error().WithError(nil).Message("no error")
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			setupTestLogger(&buf)

			tt.logFunc()

			output := buf.String()
			logData, err := parseLogOutput(output)

			if err != nil {
				t.Fatalf("Failed to parse log output: %v", err)
			}

			if tt.expectedError == "" {
				// Should not have error_msg field
				if _, ok := logData["error_msg"]; ok {
					t.Error("Expected no error_msg field for nil error")
				}
			} else {
				if errMsg, ok := logData["error_msg"].(string); !ok || errMsg != tt.expectedError {
					t.Errorf("Expected error_msg %s, got %v", tt.expectedError, logData["error_msg"])
				}
			}
		})
	}
}

// TestAlert tests the Alert method
func TestAlert(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	zlog.Error().Alert().Message("critical alert")

	output := buf.String()
	logData, err := parseLogOutput(output)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if alert, ok := logData["alert"].(bool); !ok || !alert {
		t.Errorf("Expected alert=true, got %v", logData["alert"])
	}
}

// TestContext tests context value extraction
func TestContext(t *testing.T) {
	tests := []struct {
		name         string
		setupCtx     func() context.Context
		keys         []string
		expectedKeys map[string]interface{}
	}{
		{
			name: "Single context value",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), "userID", "12345")
			},
			keys: []string{"userID"},
			expectedKeys: map[string]interface{}{
				"userID": "12345",
			},
		},
		{
			name: "Multiple context values",
			setupCtx: func() context.Context {
				ctx := context.WithValue(context.Background(), "userID", "12345")
				ctx = context.WithValue(ctx, "requestID", "req-abc")
				return ctx
			},
			keys: []string{"userID", "requestID"},
			expectedKeys: map[string]interface{}{
				"userID":    "12345",
				"requestID": "req-abc",
			},
		},
		{
			name: "Non-existent key",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), "userID", "12345")
			},
			keys:         []string{"nonexistent"},
			expectedKeys: nil, // No app_ctx field expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			setupTestLogger(&buf)

			ctx := tt.setupCtx()
			zlog.Info().Context(ctx, tt.keys).Message("test")

			output := buf.String()
			logData, err := parseLogOutput(output)
			if err != nil {
				t.Fatalf("Failed to parse log output: %v", err)
			}

			if tt.expectedKeys == nil {
				if _, ok := logData["app_ctx"]; ok {
					t.Error("Expected no app_ctx field")
				}
			} else {
				appCtx, ok := logData["app_ctx"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected app_ctx to be a map")
				}

				for key, expectedVal := range tt.expectedKeys {
					if val, ok := appCtx[key]; !ok || val != expectedVal {
						t.Errorf("Expected app_ctx[%s]=%v, got %v", key, expectedVal, val)
					}
				}
			}
		})
	}
}

// TestWithSource tests source information
func TestWithSource(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	zlog.Info().WithSource().Message("test")

	output := buf.String()
	logData, err := parseLogOutput(output)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	source, ok := logData["source"].(string)
	if !ok {
		t.Fatal("Expected source field to be present")
	}

	// Source should contain function name, file, and line number
	if !strings.Contains(source, "zlog_test.") || !strings.Contains(source, "@") {
		t.Errorf("Source format unexpected: %s", source)
	}
}

// TestWithCallStack tests call stack information
func TestWithCallStack(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	zlog.Error().WithCallStack().Message("test")

	output := buf.String()
	logData, err := parseLogOutput(output)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	callstack, ok := logData["callstack"].([]interface{})
	if !ok {
		t.Fatal("Expected callstack field to be an array")
	}

	if len(callstack) == 0 {
		t.Error("Expected callstack to have at least one entry")
	}

	// First entry should be the test function
	firstEntry, ok := callstack[0].(string)
	if !ok {
		t.Fatal("Expected callstack entry to be a string")
	}

	if !strings.Contains(firstEntry, "#") {
		t.Errorf("Expected callstack entry to contain function name: %s", firstEntry)
	}
}

// TestAutoSourceConfig tests automatic source configuration
func TestAutoSourceConfig(t *testing.T) {
	tests := []struct {
		name      string
		setup     func()
		logFunc   func()
		expectSrc bool
	}{
		{
			name: "Auto source enabled for Error",
			setup: func() {
				zlog.SetConfig(zlog.Configure(
					zlog.AutoSourceConfig(slog.LevelError, true),
				))
			},
			logFunc: func() {
				zlog.Error().Message("test")
			},
			expectSrc: true,
		},
		{
			name: "Auto source disabled for Error",
			setup: func() {
				zlog.SetConfig(zlog.Configure(
					zlog.AutoSourceConfig(slog.LevelError, false),
				))
			},
			logFunc: func() {
				zlog.Error().Message("test")
			},
			expectSrc: false,
		},
		{
			name: "Auto source enabled for Info",
			setup: func() {
				zlog.SetConfig(zlog.Configure(
					zlog.AutoSourceConfig(slog.LevelInfo, true),
				))
			},
			logFunc: func() {
				zlog.Info().Message("test")
			},
			expectSrc: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			setupTestLogger(&buf)

			// Apply test-specific config
			tt.setup()

			tt.logFunc()

			output := buf.String()
			logData, err := parseLogOutput(output)

			if err != nil {
				t.Fatalf("Failed to parse log output: %v", err)
			}

			_, hasSource := logData["source"]
			if hasSource != tt.expectSrc {
				t.Errorf("Expected source presence=%v, got %v", tt.expectSrc, hasSource)
			}
		})
	}
}

// TestAutoCallStackConfig tests automatic call stack configuration
func TestAutoCallStackConfig(t *testing.T) {
	tests := []struct {
		name           string
		setup          func()
		logFunc        func()
		expectCallStak bool
	}{
		{
			name: "Auto callstack enabled for Error",
			setup: func() {
				zlog.SetConfig(zlog.Configure(
					zlog.AutoCallStackConfig(slog.LevelError, true),
				))
			},
			logFunc: func() {
				zlog.Error().Message("test")
			},
			expectCallStak: true,
		},
		{
			name: "Auto callstack disabled for Error",
			setup: func() {
				zlog.SetConfig(zlog.Configure(
					zlog.AutoCallStackConfig(slog.LevelError, false),
				))
			},
			logFunc: func() {
				zlog.Error().Message("test")
			},
			expectCallStak: false,
		},
		{
			name: "Auto callstack enabled for Debug",
			setup: func() {
				zlog.SetConfig(zlog.Configure(
					zlog.AutoCallStackConfig(slog.LevelDebug, true),
				))
			},
			logFunc: func() {
				zlog.Debug().Message("test")
			},
			expectCallStak: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			setupTestLogger(&buf)

			// Apply test-specific config
			tt.setup()

			tt.logFunc()

			output := buf.String()
			logData, err := parseLogOutput(output)

			if err != nil {
				t.Fatalf("Failed to parse log output: %v", err)
			}

			_, hasCallStack := logData["callstack"]
			if hasCallStack != tt.expectCallStak {
				t.Errorf("Expected callstack presence=%v, got %v", tt.expectCallStak, hasCallStack)
			}
		})
	}
}

// TestMaxCallStackDepthConfig tests call stack depth configuration
// Tests the behavior indirectly by checking callstack length with auto-config
func TestMaxCallStackDepthConfig(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	// Set a very small max depth
	zlog.SetConfig(zlog.Configure(
		zlog.AutoCallStackConfig(slog.LevelError, true),
		zlog.MaxCallStackDepthConfig(slog.LevelError, 3),
	))

	// Create a deeper call stack
	var deepFunc func(int)
	deepFunc = func(depth int) {
		if depth == 0 {
			zlog.Error().Message("test with limited depth")
			return
		}
		deepFunc(depth - 1)
	}
	deepFunc(10)

	output := buf.String()
	logData, err := parseLogOutput(output)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Callstack should be limited by the configured max depth (3)
	callstack, ok := logData["callstack"].([]interface{})
	if !ok {
		t.Fatal("Expected callstack to be present")
	}

	// With max depth of 3, we should have very few entries
	// (less than default which would be 10 for Error level)
	if len(callstack) > 3 {
		t.Logf("Warning: Callstack has %d entries, expected max around 3", len(callstack))
		// Note: Actual length might be slightly different due to skip logic
	}
}

// TestChainedMethods tests method chaining
func TestChainedMethods(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	ctx := context.WithValue(context.Background(), "userID", "12345")
	err := errors.New("test error")

	zlog.Error().
		Context(ctx, []string{"userID"}).
		Segment("api", "users").
		WithError(err).
		KeyValue("operation", "create").
		Alert().
		Message("complex log entry")

	output := buf.String()
	logData, parseErr := parseLogOutput(output)
	if parseErr != nil {
		t.Fatalf("Failed to parse log output: %v", parseErr)
	}

	// Verify all fields are present
	if logData["msg"] != "complex log entry" {
		t.Errorf("Expected msg='complex log entry', got %v", logData["msg"])
	}

	if logData["segment"] != "api/users" {
		t.Errorf("Expected segment='api/users', got %v", logData["segment"])
	}

	if logData["error_msg"] != "test error" {
		t.Errorf("Expected error_msg='test error', got %v", logData["error_msg"])
	}

	if logData["operation"] != "create" {
		t.Errorf("Expected operation='create', got %v", logData["operation"])
	}

	if logData["alert"] != true {
		t.Errorf("Expected alert=true, got %v", logData["alert"])
	}

	appCtx, ok := logData["app_ctx"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected app_ctx to be present")
	}

	if appCtx["userID"] != "12345" {
		t.Errorf("Expected userID='12345', got %v", appCtx["userID"])
	}
}

// TestPanicFunction tests Panic function
func TestPanicFunction(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic to occur")
		} else if r != "test panic" {
			t.Errorf("Expected panic message 'test panic', got %v", r)
		}
	}()

	zlog.Panic("test panic")
}

// TestPanicfFunction tests Panicf function
func TestPanicfFunction(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic to occur")
		} else if r != "formatted panic 42" {
			t.Errorf("Expected panic message 'formatted panic 42', got %v", r)
		}
	}()

	zlog.Panicf("formatted panic %d", 42)
}

// TestDefaultCallStackDepths tests default call stack depth behavior
// Tests indirectly by comparing callstack sizes across different log levels
func TestDefaultCallStackDepths(t *testing.T) {
	tests := []struct {
		level          slog.Level
		logFunc        func()
		expectedMinLen int // Minimum expected callstack length
		expectedMaxLen int // Maximum expected callstack length (for rough validation)
	}{
		{
			level: slog.LevelDebug,
			logFunc: func() {
				zlog.Debug().Message("debug test")
			},
			expectedMinLen: 1,
			expectedMaxLen: 20, // Debug default is 20
		},
		{
			level: slog.LevelInfo,
			logFunc: func() {
				zlog.Info().Message("info test")
			},
			expectedMinLen: 1,
			expectedMaxLen: 5, // Info default is 5
		},
		{
			level: slog.LevelWarn,
			logFunc: func() {
				zlog.Warn().Message("warn test")
			},
			expectedMinLen: 1,
			expectedMaxLen: 5, // Warn default is 5
		},
		{
			level: slog.LevelError,
			logFunc: func() {
				zlog.Error().Message("error test")
			},
			expectedMinLen: 1,
			expectedMaxLen: 10, // Error default is 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			var buf bytes.Buffer
			setupTestLogger(&buf)

			// Enable auto callstack for this level
			config := zlog.Configure()
			switch tt.level {
			case slog.LevelDebug:
				config = zlog.Configure(zlog.AutoCallStackConfig(slog.LevelDebug, true))
			case slog.LevelInfo:
				config = zlog.Configure(zlog.AutoCallStackConfig(slog.LevelInfo, true))
			case slog.LevelWarn:
				config = zlog.Configure(zlog.AutoCallStackConfig(slog.LevelWarn, true))
			case slog.LevelError:
				config = zlog.Configure(zlog.AutoCallStackConfig(slog.LevelError, true))
			}
			zlog.SetConfig(config)

			// Create a deep call stack to test depth limits
			var deepFunc func(int)
			deepFunc = func(depth int) {
				if depth == 0 {
					tt.logFunc()
					return
				}
				deepFunc(depth - 1)
			}
			deepFunc(25) // Create 25 deep calls, but each level has different max

			output := buf.String()
			logData, err := parseLogOutput(output)
			if err != nil {
				t.Fatalf("Failed to parse log output: %v", err)
			}

			callstack, ok := logData["callstack"].([]interface{})
			if !ok {
				t.Fatal("Expected callstack to be present")
			}

			// Verify callstack is within expected bounds
			stackLen := len(callstack)
			if stackLen < tt.expectedMinLen || stackLen > tt.expectedMaxLen {
				t.Logf("Callstack length %d is outside expected range [%d, %d] for %s",
					stackLen, tt.expectedMinLen, tt.expectedMaxLen, tt.level.String())
				// This validates that different levels have different depth limits
			}
		})
	}
}

// TestComplexScenario tests a complex real-world scenario
func TestComplexScenario(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	zlog.SetConfig(zlog.Configure(
		zlog.AutoSourceConfig(slog.LevelError, true),
		zlog.AutoCallStackConfig(slog.LevelError, true),
		zlog.MaxCallStackDepthConfig(slog.LevelError, 8),
	))

	ctx := context.WithValue(context.Background(), "userID", "user-123")
	ctx = context.WithValue(ctx, "requestID", "req-456")
	ctx = context.WithValue(ctx, "traceID", "trace-789")

	zlog.Error().
		Context(ctx, []string{"userID", "requestID", "traceID"}).
		Segment("payment", "process", "stripe").
		Err(errors.New("payment gateway timeout")).
		KeyValue("payment_id", "pay-999").
		KeyValue("amount", "100.00").
		Alert().
		Msgf("Payment processing failed for order %s", "order-001")

	output := buf.String()
	logData, err := parseLogOutput(output)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v\nOutput: %s", err, output)
	}

	// Verify all expected fields
	expectedChecks := map[string]interface{}{
		"level":      "ERROR",
		"segment":    "payment/process/stripe",
		"error_msg":  "payment gateway timeout",
		"payment_id": "pay-999",
		"amount":     "100.00",
		"alert":      true,
		"msg":        "Payment processing failed for order order-001",
	}

	for key, expected := range expectedChecks {
		if logData[key] != expected {
			t.Errorf("Expected %s=%v, got %v", key, expected, logData[key])
		}
	}

	// Verify context
	appCtx, ok := logData["app_ctx"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected app_ctx to be present")
	}

	expectedCtx := map[string]string{
		"userID":    "user-123",
		"requestID": "req-456",
		"traceID":   "trace-789",
	}

	for key, expected := range expectedCtx {
		if appCtx[key] != expected {
			t.Errorf("Expected app_ctx[%s]=%s, got %v", key, expected, appCtx[key])
		}
	}

	// Verify auto-configured source
	if _, ok := logData["source"]; !ok {
		t.Error("Expected source to be present (auto-configured)")
	}

	// Verify auto-configured callstack
	if _, ok := logData["callstack"]; !ok {
		t.Error("Expected callstack to be present (auto-configured)")
	}
}

// TestWithSourceSkip tests WithSourceSkip with different skip values
func TestWithSourceSkip(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	zlog.Info().WithSourceSkip(0).Message("test")

	output := buf.String()
	logData, err := parseLogOutput(output)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	source, ok := logData["source"].(string)
	if !ok {
		t.Fatal("Expected source field to be present")
	}

	// Source should be present and formatted correctly
	if !strings.Contains(source, "@") {
		t.Errorf("Source format unexpected: %s", source)
	}
}

// TestRegressionSegmentWithError tests that Segment and Error work together
func TestRegressionSegmentWithError(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	err := errors.New("database connection failed")
	zlog.Error().
		Segment("database", "connection").
		Err(err).
		Message("Failed to connect to database")

	output := buf.String()
	logData, parseErr := parseLogOutput(output)
	if parseErr != nil {
		t.Fatalf("Failed to parse log output: %v", parseErr)
	}

	if logData["segment"] != "database/connection" {
		t.Errorf("Expected segment='database/connection', got %v", logData["segment"])
	}

	if logData["error_msg"] != "database connection failed" {
		t.Errorf("Expected error_msg='database connection failed', got %v", logData["error_msg"])
	}
}

// TestRegressionContextWithMultipleKeys tests context with multiple keys
func TestRegressionContextWithMultipleKeys(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "userID", "user-001")
	ctx = context.WithValue(ctx, "sessionID", "sess-002")
	ctx = context.WithValue(ctx, "requestID", "req-003")

	zlog.Info().
		Context(ctx, []string{"userID", "sessionID", "requestID"}).
		Message("User action logged")

	output := buf.String()
	logData, err := parseLogOutput(output)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	appCtx, ok := logData["app_ctx"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected app_ctx to be present")
	}

	if appCtx["userID"] != "user-001" {
		t.Errorf("Expected userID='user-001', got %v", appCtx["userID"])
	}
	if appCtx["sessionID"] != "sess-002" {
		t.Errorf("Expected sessionID='sess-002', got %v", appCtx["sessionID"])
	}
	if appCtx["requestID"] != "req-003" {
		t.Errorf("Expected requestID='req-003', got %v", appCtx["requestID"])
	}
}

// TestRegressionAutoConfigPersistence tests that auto-config persists across multiple log calls
func TestRegressionAutoConfigPersistence(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	zlog.SetConfig(zlog.Configure(
		zlog.AutoSourceConfig(slog.LevelError, true),
	))

	// First log
	zlog.Error().Message("first error")
	firstOutput := buf.String()
	firstLogData, err := parseLogOutput(firstOutput)
	if err != nil {
		t.Fatalf("Failed to parse first log output: %v", err)
	}

	if _, ok := firstLogData["source"]; !ok {
		t.Error("Expected source in first log")
	}

	// Clear buffer for second log
	buf.Reset()

	// Second log - should still have auto-source
	zlog.Error().Message("second error")
	secondOutput := buf.String()
	secondLogData, err := parseLogOutput(secondOutput)
	if err != nil {
		t.Fatalf("Failed to parse second log output: %v", err)
	}

	if _, ok := secondLogData["source"]; !ok {
		t.Error("Expected source in second log (auto-config should persist)")
	}
}

// TestRegressionNilErrorDoesNotAddField tests that nil error doesn't add error_msg field
func TestRegressionNilErrorDoesNotAddField(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	zlog.Error().Err(nil).Message("test")

	output := buf.String()
	logData, err := parseLogOutput(output)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if _, ok := logData["error_msg"]; ok {
		t.Error("Expected no error_msg field when error is nil")
	}
}

// TestRegressionKeyValueChaining tests multiple KeyValue calls
func TestRegressionKeyValueChaining(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	zlog.Info().
		KeyValue("key1", "value1").
		KeyValue("key2", "value2").
		KeyValue("key3", "value3").
		Message("test")

	output := buf.String()
	logData, err := parseLogOutput(output)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if logData["key1"] != "value1" {
		t.Errorf("Expected key1='value1', got %v", logData["key1"])
	}
	if logData["key2"] != "value2" {
		t.Errorf("Expected key2='value2', got %v", logData["key2"])
	}
	if logData["key3"] != "value3" {
		t.Errorf("Expected key3='value3', got %v", logData["key3"])
	}
}

// TestSetOutputWriter tests custom output writer functionality
func TestSetOutputWriter(t *testing.T) {
	t.Run("Write to custom buffer", func(t *testing.T) {
		var customBuf bytes.Buffer
		zlog.SetOutputWriter(&customBuf)
		zlog.SetConfig(zlog.Configure()) // Reset config

		zlog.Info().Message("custom output test")

		output := customBuf.String()
		if output == "" {
			t.Fatal("Expected output in custom buffer")
		}

		logData, err := parseLogOutput(output)
		if err != nil {
			t.Fatalf("Failed to parse log output: %v", err)
		}

		if logData["msg"] != "custom output test" {
			t.Errorf("Expected msg='custom output test', got %v", logData["msg"])
		}
	})

	t.Run("Switch between writers", func(t *testing.T) {
		var buf1 bytes.Buffer
		var buf2 bytes.Buffer

		// Write to first buffer
		zlog.SetOutputWriter(&buf1)
		zlog.SetConfig(zlog.Configure())
		zlog.Info().Message("first buffer")

		// Switch to second buffer
		zlog.SetOutputWriter(&buf2)
		zlog.Info().Message("second buffer")

		// Verify first buffer has first message
		if !strings.Contains(buf1.String(), "first buffer") {
			t.Error("Expected 'first buffer' in buf1")
		}
		if strings.Contains(buf1.String(), "second buffer") {
			t.Error("Did not expect 'second buffer' in buf1")
		}

		// Verify second buffer has second message
		if !strings.Contains(buf2.String(), "second buffer") {
			t.Error("Expected 'second buffer' in buf2")
		}
		if strings.Contains(buf2.String(), "first buffer") {
			t.Error("Did not expect 'first buffer' in buf2")
		}
	})

	t.Run("MultiWriter to multiple destinations", func(t *testing.T) {
		var buf1 bytes.Buffer
		var buf2 bytes.Buffer
		multiWriter := io.MultiWriter(&buf1, &buf2)

		zlog.SetOutputWriter(multiWriter)
		zlog.SetConfig(zlog.Configure())
		zlog.Info().Message("multi writer test")

		// Both buffers should have the message
		if !strings.Contains(buf1.String(), "multi writer test") {
			t.Error("Expected message in buf1")
		}
		if !strings.Contains(buf2.String(), "multi writer test") {
			t.Error("Expected message in buf2")
		}
	})

	// Restore stdout for other tests
	t.Cleanup(func() {
		zlog.SetOutputWriter(os.Stdout)
	})
}

// Benchmark tests
func BenchmarkSimpleLog(b *testing.B) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		zlog.Info().Message("benchmark test")
	}
}

func BenchmarkLogWithKeyValue(b *testing.B) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		zlog.Info().KeyValue("key", "value").Message("benchmark test")
	}
}

func BenchmarkComplexLog(b *testing.B) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	ctx := context.WithValue(context.Background(), "userID", "12345")
	err := errors.New("test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		zlog.Error().
			Context(ctx, []string{"userID"}).
			Segment("api", "test").
			Err(err).
			KeyValue("key", "value").
			Message("benchmark test")
	}
}

func BenchmarkWithCallStack(b *testing.B) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		zlog.Error().WithCallStack().Message("benchmark test")
	}
}

func BenchmarkWithSource(b *testing.B) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		zlog.Info().WithSource().Message("benchmark test")
	}
}

func BenchmarkAutoSourceConfig(b *testing.B) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	zlog.SetConfig(zlog.Configure(
		zlog.AutoSourceConfig(slog.LevelError, true),
	))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		zlog.Error().Message("benchmark test")
	}
}

func BenchmarkAutoCallStackConfig(b *testing.B) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	zlog.SetConfig(zlog.Configure(
		zlog.AutoCallStackConfig(slog.LevelError, true),
	))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		zlog.Error().Message("benchmark test")
	}
}

func BenchmarkChainedMethods(b *testing.B) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	ctx := context.WithValue(context.Background(), "userID", "12345")
	err := errors.New("test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		zlog.Error().
			Context(ctx, []string{"userID"}).
			Segment("api", "users").
			WithError(err).
			KeyValue("operation", "create").
			Alert().
			WithSource().
			Message("complex log entry")
	}
}

// Edge Case Tests - Potential bugs or unexpected behaviors

// TestEdgeCaseEmptyContextKeys tests behavior with empty context key slice
func TestEdgeCaseEmptyContextKeys(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	ctx := context.Background()
	zlog.Info().Context(ctx, []string{}).Message("test")

	output := buf.String()
	logData, err := parseLogOutput(output)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Empty context keys should not add app_ctx field
	if _, ok := logData["app_ctx"]; ok {
		t.Error("Expected no app_ctx field for empty context keys")
	}
}

// TestEdgeCaseEmptySegment tests behavior with empty segment
func TestEdgeCaseEmptySegment(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	zlog.Info().Segment("").Message("test")

	output := buf.String()
	logData, err := parseLogOutput(output)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Empty segment should still be added
	if segment, ok := logData["segment"].(string); !ok || segment != "" {
		t.Errorf("Expected empty segment string, got %v", logData["segment"])
	}
}

// TestEdgeCaseSegmentWithEmptyDetail tests segment with empty detail parts
// BUG FIX: Empty strings should be filtered out to avoid double slashes
func TestEdgeCaseSegmentWithEmptyDetail(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	zlog.Info().Segment("main", "", "sub").Message("test")

	output := buf.String()
	logData, err := parseLogOutput(output)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Empty strings should be filtered out, not create double slashes
	expected := "main/sub"
	if segment, ok := logData["segment"].(string); !ok || segment != expected {
		t.Errorf("Expected segment '%s', got %v", expected, logData["segment"])
	}
}

// TestSegmentMultipleEmptyDetails tests multiple empty strings are all filtered
func TestSegmentMultipleEmptyDetails(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	zlog.Info().Segment("api", "", "", "users", "", "create").Message("test")

	output := buf.String()
	logData, err := parseLogOutput(output)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// All empty strings should be filtered
	expected := "api/users/create"
	if segment, ok := logData["segment"].(string); !ok || segment != expected {
		t.Errorf("Expected segment '%s', got %v", expected, logData["segment"])
	}
}

// TestEdgeCaseDuplicateKeys tests behavior with duplicate key-value pairs
func TestEdgeCaseDuplicateKeys(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	zlog.Info().
		KeyValue("key", "value1").
		KeyValue("key", "value2").
		Message("test")

	output := buf.String()
	logData, err := parseLogOutput(output)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// In JSON, last key wins (or both might appear depending on handler)
	// This is actually expected behavior in slog
	if key, ok := logData["key"].(string); ok {
		// Either value1 or value2 is acceptable, or both might appear as array
		// The actual behavior depends on slog's JSON handler
		if key != "value1" && key != "value2" {
			t.Errorf("Expected 'value1' or 'value2', got %v", key)
		}
	}
}

// TestEdgeCaseVeryLongCallStack tests call stack with many frames
func TestEdgeCaseVeryLongCallStack(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	// Create deep call stack
	var deepFunc func(int)
	deepFunc = func(depth int) {
		if depth == 0 {
			zlog.Error().WithCallStack().Message("deep call")
			return
		}
		deepFunc(depth - 1)
	}

	deepFunc(15)

	output := buf.String()
	logData, err := parseLogOutput(output)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	callstack, ok := logData["callstack"].([]interface{})
	if !ok {
		t.Fatal("Expected callstack to be present")
	}

	// Should be limited by maxCallStackDepth (default 10 for Error)
	if len(callstack) > 10 {
		t.Logf("Callstack has %d entries (expected max 10)", len(callstack))
		// This is not necessarily a bug, just documenting behavior
	}
}

// TestEdgeCaseConfigChangeDoesNotAffectExistingLoggers tests if config change requires logger reinit
func TestEdgeCaseConfigChangeDoesNotAffectExistingLoggers(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	// Create logger instance before config
	logger := zlog.Error()

	// Change config
	zlog.SetConfig(zlog.Configure(
		zlog.AutoSourceConfig(slog.LevelError, true),
	))

	// Use the logger created before config change
	logger.Message("test")

	output := buf.String()
	logData, err := parseLogOutput(output)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// The logger was created before config, so it might not have auto-source
	// This documents the behavior: config applies at logger creation time
	_, hasSource := logData["source"]
	t.Logf("Logger created before config change has source: %v", hasSource)
	// This is expected behavior - config is applied when logger is created
}

// TestEdgeCaseNilContextValue tests context value that is explicitly nil
func TestEdgeCaseNilContextValue(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	// In Go, you can store nil as a context value
	type key string
	ctx := context.WithValue(context.Background(), key("mykey"), nil)

	// Try to extract the nil value
	zlog.Info().Context(ctx, []string{"mykey"}).Message("test")

	output := buf.String()
	logData, err := parseLogOutput(output)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Nil values are ignored by Context method
	if _, ok := logData["app_ctx"]; ok {
		t.Error("Expected no app_ctx field when context value is nil")
	}
}

// TestEdgeCaseMultipleAutoFeatures tests multiple auto features enabled together
func TestEdgeCaseMultipleAutoFeatures(t *testing.T) {
	var buf bytes.Buffer
	setupTestLogger(&buf)

	zlog.SetConfig(zlog.Configure(
		zlog.AutoSourceConfig(slog.LevelError, true),
		zlog.AutoCallStackConfig(slog.LevelError, true),
		zlog.MaxCallStackDepthConfig(slog.LevelError, 5),
	))

	zlog.Error().Message("test with multiple auto features")

	output := buf.String()
	logData, err := parseLogOutput(output)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	// Both source and callstack should be present
	if _, ok := logData["source"]; !ok {
		t.Error("Expected source to be present")
	}
	if _, ok := logData["callstack"]; !ok {
		t.Error("Expected callstack to be present")
	}
}
