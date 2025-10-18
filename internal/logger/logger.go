package logger

import (
	"log/slog"
)

var Logger *slog.Logger

func Init(loggerLevel string) {
	Logger = slog.Default()
}
