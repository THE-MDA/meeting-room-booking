package logger

import (
    "log/slog"
    "os"
)

func Init(level slog.Level) {
    var handler slog.Handler
    
    if os.Getenv("ENVIRONMENT") == "production" {
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level:     level,
            AddSource: true,
        })
    } else {
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level:     level,
            AddSource: true,
        })
    }
    
    slog.SetDefault(slog.New(handler))
}

func Debug(msg string, args ...any) {
    slog.Debug(msg, args...)
}

func Info(msg string, args ...any) {
    slog.Info(msg, args...)
}

func Warn(msg string, args ...any) {
    slog.Warn(msg, args...)
}

func Error(msg string, args ...any) {
    slog.Error(msg, args...)
}