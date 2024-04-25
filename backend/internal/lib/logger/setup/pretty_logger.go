package setup

import (
	"log/slog"
	"os"

	"github.com/Prrromanssss/DAEC-fullstack/internal/lib/logger/handlers/slogpretty"
)

func SetupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout) // os.Stdout !!!

	return slog.New(handler)
}
