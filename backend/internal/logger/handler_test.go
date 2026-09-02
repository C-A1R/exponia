package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestHandlerWritesAttributeNamesAndValues(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(NewHandler(&output, slog.LevelInfo))

	logger.Info(
		"application starting",
		slog.String("version", "0.1.0"),
		slog.String("commit", "5074874"),
		slog.String("build_time", "2026-09-02T09:40:27Z"),
	)

	line := output.String()
	for _, expected := range []string{
		"[INFO] application starting",
		"version:0.1.0",
		"commit:5074874",
		"build_time:2026-09-02T09:40:27Z",
	} {
		if !strings.Contains(line, expected) {
			t.Fatalf("expected log to contain %q, got %q", expected, line)
		}
	}
}
