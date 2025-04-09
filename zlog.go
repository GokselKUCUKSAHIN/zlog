package zlog

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

type ZLogger interface {
	Context(ctx context.Context, keys []string) ZLogger
	KeyValue(key, value string) ZLogger
	Segment(mainSegment string, detail ...string) ZLogger
	Error(err error) ZLogger
	Message(message string)
	Messagef(format string, args ...any)
}

type zlogImpl struct {
	logger *slog.Logger
	attrs  []any
}

var (
	debugLogger      *slog.Logger
	infoLogger       *slog.Logger
	warnLogger       *slog.Logger
	errorLogger      *slog.Logger
	loggerMap        map[slog.Level]*slog.Logger
	sensitiveHeaders = map[string]struct{}{
		"authorization": {},
		"x-api-key":     {},
		"api-key":       {},
		"jwt":           {},
		"token":         {},
		"password":      {},
		"credit-card":   {},
		"tckn":          {},
		"wasmjwt":       {},
		"cookie":        {},
	}
)

func init() {
	debugLogger = initNewSlog(slog.LevelDebug)
	infoLogger = initNewSlog(slog.LevelInfo)
	warnLogger = initNewSlog(slog.LevelWarn)
	errorLogger = initNewSlog(slog.LevelError)

	loggerMap = map[slog.Level]*slog.Logger{
		slog.LevelDebug: debugLogger,
		slog.LevelInfo:  infoLogger,
		slog.LevelWarn:  warnLogger,
		slog.LevelError: errorLogger,
	}
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
		ReplaceAttr: replaceAttr,
	})
	return slog.New(jsonHandler)
}

func Debug() ZLogger {
	return &zlogImpl{
		logger: debugLogger,
	}
}

func Info() ZLogger {
	return &zlogImpl{
		logger: infoLogger,
	}
}

func Warn() ZLogger {
	return &zlogImpl{
		logger: warnLogger,
	}
}

func Error() ZLogger {
	return &zlogImpl{
		logger: errorLogger,
	}
}

//func AutoLevel[T any](cmp T, leveler func(T) slog.Level) ZLogger {
//	logger, exists := loggerMap[leveler(cmp)]
//	if !exists {
//		logger = errorLogger.With(
//			slog.String("zlog_warning", "AutoLevel failed. used default Error Logger as fail safe"),
//		)
//	}
//	return &zlogImpl{
//		logger: logger,
//	}
//}

func (z *zlogImpl) Context(ctx context.Context, keys []string) ZLogger {
	contextMap := make(map[string]any, len(keys))
	for _, key := range keys {
		//< might be overkill maybe seperate as ContextSecure
		_, exists := sensitiveHeaders[strings.ToLower(key)]
		if exists {
			contextMap[key] = "[REDACTED]"
			continue
		}
		//>
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
