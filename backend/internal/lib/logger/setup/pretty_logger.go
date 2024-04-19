package setup

import (
	"io"
	"log/slog"
	"os"

	"github.com/Prrromanssss/DAEE-fullstack/internal/lib/logger/handlers/slogpretty"
)

func SetupPrettySlog(logFile io.Writer) *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout) // os.Stdout !!!

	return slog.New(handler)
}
