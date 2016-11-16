package log

import (
	"os"

	"golang.org/x/net/context"
)

type loggerCtxKeyType int

const loggerCtxKey loggerCtxKeyType = 0

func SetLoggerToContext(ctx context.Context, logger ILogger) context.Context {
	return context.WithValue(ctx, loggerCtxKey, logger)
}

func GetLoggerFromContext(ctx context.Context) ILogger {
	if logger, ok := ctx.Value(loggerCtxKey).(ILogger); ok {
		return logger
	}
	return New(os.Stderr, "", ERROR)
}
