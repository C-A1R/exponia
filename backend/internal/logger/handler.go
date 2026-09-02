package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
)

type Handler struct {
	output io.Writer
	level  slog.Level
	mutex  *sync.Mutex
	attrs  []slog.Attr
}

func NewHandler(output io.Writer, level slog.Level) *Handler {
	return &Handler{
		output: output,
		level:  level,
		mutex:  &sync.Mutex{},
	}
}

func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *Handler) Handle(_ context.Context, record slog.Record) error {
	var line strings.Builder
	line.WriteString(record.Time.Format("2006-01-02 15:04:05"))
	line.WriteString(" [")
	line.WriteString(record.Level.String())
	line.WriteString("] ")
	line.WriteString(record.Message)

	for _, attr := range h.attrs {
		appendAttr(&line, attr)
	}

	record.Attrs(func(attr slog.Attr) bool {
		appendAttr(&line, attr)
		return true
	})

	line.WriteByte('\n')

	h.mutex.Lock()
	defer h.mutex.Unlock()

	_, err := io.WriteString(h.output, line.String())
	return err
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	combined := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	combined = append(combined, h.attrs...)
	combined = append(combined, attrs...)

	return &Handler{
		output: h.output,
		level:  h.level,
		mutex:  h.mutex,
		attrs:  combined,
	}
}

func (h *Handler) WithGroup(_ string) slog.Handler {
	return h
}

func appendAttr(line *strings.Builder, attr slog.Attr) {
	if attr.Equal(slog.Attr{}) {
		return
	}

	attr.Value = attr.Value.Resolve()

	line.WriteByte(' ')
	line.WriteString(attr.Key)
	line.WriteByte(':')

	switch attr.Value.Kind() {
	case slog.KindString:
		line.WriteString(attr.Value.String())
	default:
		fmt.Fprint(line, attr.Value.Any())
	}
}
