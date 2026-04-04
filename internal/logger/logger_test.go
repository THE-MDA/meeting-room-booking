package logger

import (
    "log/slog"
    "testing"
)

func TestInit(t *testing.T) {
    Init(slog.LevelDebug)
    
    Init(slog.LevelInfo)
    
    Init(slog.LevelWarn)
    
    Init(slog.LevelError)
}

func TestDebug(t *testing.T) {
    Debug("test message", "key", "value")
}

func TestInfo(t *testing.T) {
    Info("test message", "key", "value")
}

func TestWarn(t *testing.T) {
    Warn("test message", "key", "value")
}

func TestError(t *testing.T) {
    Error("test message", "key", "value")
}