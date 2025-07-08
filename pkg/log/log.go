package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

type logStruct struct {
	sourceId string
	slog.Handler
}

func (h *logStruct) Handle(ctx context.Context, r slog.Record) error {
	r.Message = fmt.Sprintf("ID:%s -> %s", h.sourceId, r.Message)
	return h.Handler.Handle(ctx, r)
}

var defaultLogger *slog.Logger

func init() {
	defaultLogger = slog.New(
		&logStruct{
			Handler: slog.NewTextHandler(os.Stdout, nil),
			sourceId: "1",
		})
}

func SetId(s string) {
	defaultLogger = slog.New(
		&logStruct{
			Handler: slog.NewTextHandler(os.Stdout, nil),
			sourceId: s,
		})
}

func Infof(s string, args ...any) {
	defaultLogger.Info(s, args...)
}

func Errorf(s string, args ...any) {
	defaultLogger.Error(s, args...)
}

func Fatalf(s string, args ...any) {
	defaultLogger.Error(s, args...)
}

func Debugf(s string, args ...any) {
	defaultLogger.Debug(s, args...)
}

func Warnf(s string, args ...any) {
	defaultLogger.Warn(s, args...)
}
