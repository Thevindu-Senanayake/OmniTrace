package logging

import (
	"log/slog"
	"os"
)

// New returns a JSON slog logger tagged with the service name.
func New(serviceName string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", serviceName)
}
