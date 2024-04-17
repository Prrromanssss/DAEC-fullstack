package setup

import (
	"log/slog"
	"os"

	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/handlers/slogpretty"
)

func SetupPrettySlog(logFile *os.File) *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(logFile)

	return slog.New(handler)
}
