package zlog

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"
)

type ZLogger interface {
	Context(ctx context.Context, keys []string) ZLogger
	Segment(mainSegment string, detail ...string) ZLogger
	Error(err error) ZLogger
	AddSource() ZLogger
	AddCallStack() ZLogger
	Message(message string)
	Messagef(format string, args ...any)
}

type zlogImpl struct {
	logger            *slog.Logger
	attrs             []any
	maxCallStackDepth int
}

var (
	debugLogger *slog.Logger
	infoLogger  *slog.Logger
	warnLogger  *slog.Logger
	errorLogger *slog.Logger
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

func Debug() ZLogger {
	return &zlogImpl{
		logger:            debugLogger,
		maxCallStackDepth: 20,
	}
}

func Info() ZLogger {
	return &zlogImpl{
		logger:            infoLogger,
		maxCallStackDepth: 5,
	}
}

func Warn() ZLogger {
	return &zlogImpl{
		logger:            warnLogger,
		maxCallStackDepth: 5,
	}
}

func Error() ZLogger {
	return &zlogImpl{
		logger:            errorLogger,
		maxCallStackDepth: 10,
	}
}

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

func (z *zlogImpl) KeyValue(key, value string) ZLogger {
	return z.appendAttr(slog.String(key, value))
}

func (z *zlogImpl) Segment(mainSegment string, detail ...string) ZLogger {
	if len(detail) > 0 {
		mainSegment += "/" + strings.Join(detail, "/")
	}
	return z.appendAttr(slog.String("segment", mainSegment))
}

func (z *zlogImpl) Error(err error) ZLogger {
	return z.appendAttr(slog.String("error_msg", err.Error()))
}

func (z *zlogImpl) AddSource() ZLogger {
	return z.appendAttr(slog.String("source", getSourceString(2)))
}

func (z *zlogImpl) AddCallStack() ZLogger {
	callStack := make([]string, 0)
	for skip := 2; skip < z.maxCallStackDepth; skip++ {
		current := getSourceString(skip)
		if current == "# @ :0" {
			continue
		}
		callStack = append(callStack, current)
		if strings.HasPrefix(current, "#main.main") {
			break
		}
	}
	return z.appendAttr(slog.Any("callstack", callStack))
}

func (z *zlogImpl) Message(message string) {
	z.logger.Info(message, z.attrs...)
}

func (z *zlogImpl) Messagef(format string, args ...any) {
	z.logger.Info(fmt.Sprintf(format, args...), z.attrs...)
}

func (z *zlogImpl) appendAttr(attr slog.Attr) ZLogger {
	z.attrs = append(z.attrs, attr)
	return z
}

func (z *zlogImpl) appendAttrs(attrs ...any) ZLogger {
	z.attrs = append(z.attrs, attrs...)
	return z
}

func getSourceString(skip int) string {
	pc, file, line, _ := runtime.Caller(skip)
	fn := runtime.FuncForPC(pc)
	funcName := fn.Name()
	moduleSeparator := strings.LastIndex(funcName, "/")
	if moduleSeparator != -1 {
		funcName = funcName[moduleSeparator+1:]
	}
	return fmt.Sprintf("#%s @ %s:%d", funcName, file, line)
}
