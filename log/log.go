package log

import (
	"context"

	"github.com/sirupsen/logrus"
)

type logKey string

const (
	key logKey = "ske-logger"
)

type logger interface {
	Infof(msg string, args ...interface{})
	Warnf(msg string, args ...interface{})
}

func SetLogger(ctx context.Context, logger logger) context.Context {
	return context.WithValue(ctx, key, logger)
}

func getLogger(ctx context.Context) logger {
	logger, _ := ctx.Value(key).(logger)
	if logger == nil {
		return logrus.StandardLogger()
	}
	return logger
}

func Infof(ctx context.Context, msg string, args ...interface{}) {
	getLogger(ctx).Infof(msg, args...)

}

func Warnf(ctx context.Context, msg string, args ...interface{}) {
	getLogger(ctx).Warnf(msg, args...)
}
