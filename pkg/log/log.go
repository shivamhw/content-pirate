package log

import (
	"log/slog"
	"os"
)

var defaultLogger *slog.Logger


func init() {
	defaultLogger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}


func Infof(s string, args... any){
	defaultLogger.Info(s, args...)
}


func Errorf(s string, args... any){
	defaultLogger.Error(s, args...)
}

func Fatalf(s string, args... any){
	defaultLogger.Error(s, args...)
}

func Debugf(s string, args... any){
	defaultLogger.Debug(s, args...)
}


func Warnf(s string, args... any){
	defaultLogger.Warn(s, args...)
}