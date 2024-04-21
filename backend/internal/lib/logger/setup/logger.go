package setup

import (
	"io"
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func SetupLogger(env, logPath string) *slog.Logger {
	var log *slog.Logger

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic("failed to open log file: " + err.Error())
	}
	defer logFile.Close()

	writer := io.Writer(logFile)

	switch env {
	case envLocal:
		log = SetupPrettySlog(writer)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(writer, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(writer, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}),
		)
	}
	return log
}
