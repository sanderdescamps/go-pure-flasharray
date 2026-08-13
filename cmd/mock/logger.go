package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
)

// CobraLogHandler is a custom slog.Handler that outputs logs in the format:
// "yyyy/mm/dd HH:MM:SS level message"
type CobraLogHandler struct {
	writer   io.Writer
	minLevel slog.Level
	mu       sync.Mutex
	attrs    []slog.Attr
	group    string
}

// NewCobraLogHandler creates a new CobraLogHandler with a specified minimum log level.
// Logs below this level will be skipped.
func NewCobraLogHandler(w io.Writer, minLevel slog.Level) *CobraLogHandler {
	return &CobraLogHandler{
		writer:   w,
		minLevel: minLevel,
		attrs:    []slog.Attr{},
	}
}

// Enabled reports whether the handler handles records at the given level.
// Records with a level below minLevel are skipped.
func (h *CobraLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

// Handle formats and writes the record to the writer.
func (h *CobraLogHandler) Handle(ctx context.Context, record slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Format: yyyy/mm/dd HH:MM:SS level message
	timestamp := record.Time.Format("2006/01/02 15:04:05")
	level := record.Level.String()
	message := record.Message

	// Collect all attributes (from handler and record)
	var attrs []slog.Attr
	attrs = append(attrs, h.attrs...)

	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr)
		return true
	})

	// Format: timestamp level message [attrs]
	output := fmt.Sprintf("%s %s %s", timestamp, level, message)

	// Append attributes if present
	if len(attrs) > 0 {
		output += " ["
		for i, attr := range attrs {
			if i > 0 {
				output += ", "
			}
			output += fmt.Sprintf("%s=%v", attr.Key, attr.Value.Any())
		}
		output += "]"
	}

	output += "\n"

	_, err := h.writer.Write([]byte(output))
	return err
}

// WithAttrs returns a new handler with the given attributes attached.
func (h *CobraLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CobraLogHandler{
		writer:   h.writer,
		minLevel: h.minLevel,
		attrs:    append(append([]slog.Attr{}, h.attrs...), attrs...),
		group:    h.group,
	}
}

// WithGroup returns a new handler with the given group name.
func (h *CobraLogHandler) WithGroup(name string) slog.Handler {
	return &CobraLogHandler{
		writer:   h.writer,
		minLevel: h.minLevel,
		attrs:    h.attrs,
		group:    name,
	}
}
