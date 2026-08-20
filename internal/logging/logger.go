package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"
)

const (
	reset = "\033[0m"

	black        = 30
	red          = 31
	green        = 32
	yellow       = 33
	blue         = 34
	magenta      = 35
	cyan         = 36
	lightGray    = 37
	darkGray     = 90
	lightRed     = 91
	lightGreen   = 92
	lightYellow  = 93
	lightBlue    = 94
	lightMagenta = 95
	lightCyan    = 96
	white        = 97
)

// NewHandler creates the colorized terminal handler used by normal process
// logging. Diagnostic frontends should use an observer constructor instead of
// scraping this human-oriented output.
func NewHandler(opts *slog.HandlerOptions) *Handler {
	return NewHandlerWithObserver(opts, nil)
}

// Record is the structured log representation delivered to shell and other
// diagnostic observers after handler attributes and groups are resolved.
type Record struct {
	At         time.Time
	Level      slog.Level
	Message    string
	Attributes map[string]any
}

// NewHandlerWithObserver preserves terminal logging while also publishing
// structured records to an optional in-process diagnostic consumer.
func NewHandlerWithObserver(opts *slog.HandlerOptions, observer func(Record)) *Handler {
	return newHandler(opts, observer, true)
}

// NewObserverHandler captures structured records without writing directly to
// stdout, which would corrupt a full-screen terminal UI.
func NewObserverHandler(opts *slog.HandlerOptions, observer func(Record)) *Handler {
	return newHandler(opts, observer, false)
}

// newHandler constructs the shared JSON-backed attribute resolver used by both terminal and observer handlers. Using
// slog's own handler here preserves grouping and ReplaceAttr semantics instead of reimplementing them incompletely.
func newHandler(opts *slog.HandlerOptions, observer func(Record), output bool) *Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	b := &bytes.Buffer{}

	return &Handler{
		b: b,
		h: slog.NewJSONHandler(b, &slog.HandlerOptions{
			Level:       opts.Level,
			AddSource:   opts.AddSource,
			ReplaceAttr: suppressDefaults(opts.ReplaceAttr),
		}),
		m:        &sync.Mutex{},
		observer: observer,
		output:   output,
	}
}

// colorize wraps one terminal fragment with an ANSI color and an immediate reset. Resetting each fragment prevents a
// caller's later output from inheriting the logger's color.
func colorize(colorCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, reset)
}

// Handler preserves slog's immutable WithAttrs/WithGroup semantics while
// serializing terminal output and optional observer delivery. Derived handlers
// share the same mutex because they still write to one process stream.
type Handler struct {
	h        slog.Handler
	b        *bytes.Buffer
	m        *sync.Mutex
	observer func(Record)
	output   bool
}

// Enabled delegates level filtering to the inner JSON handler so derived handlers and the terminal facade always make
// the same admission decision.
func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.h.Enabled(ctx, level)
}

// WithAttrs returns an immutable derived handler while retaining the shared buffer lock. Without the shared lock,
// sibling loggers could interleave JSON in the same scratch buffer.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{h: h.h.WithAttrs(attrs), b: h.b, m: h.m, observer: h.observer, output: h.output}
}

// WithGroup preserves slog's group nesting in a derived handler and shares synchronization with its parent.
func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{h: h.h.WithGroup(name), b: h.b, m: h.m, observer: h.observer, output: h.output}
}

const (
	timeFormat = "[15:04:05.000]"
)

// Handle resolves the complete slog record once, sends an isolated attribute map to the observer, and optionally emits
// the human-oriented terminal form. Observer delivery comes first so disabling terminal output changes no diagnostics.
func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	attrs, err := h.computeAttrs(ctx, r)
	if err != nil {
		return err
	}

	h.notifyObserver(r, attrs)

	if !h.output {
		return nil
	}

	return writeTerminalRecord(r, attrs)
}

// notifyObserver clones attributes before invoking user code so the terminal formatting phase may remove its service
// label without mutating the structured record retained by a diagnostic consumer.
func (h *Handler) notifyObserver(r slog.Record, attrs map[string]any) {
	if h.observer == nil {
		return
	}

	h.observer(Record{
		At:         r.Time,
		Level:      r.Level,
		Message:    r.Message,
		Attributes: cloneAttrs(attrs),
	})
}

// writeTerminalRecord renders the compatibility-oriented console line. The service attribute becomes a label rather
// than duplicated JSON, while every other resolved attribute remains machine-readable at the end of the line.
func writeTerminalRecord(r slog.Record, attrs map[string]any) error {
	service := attrs["service"]
	if service == nil {
		service = "Dark Magic"
	} else {
		delete(attrs, "service")
	}

	encodedAttrs, err := json.Marshal(attrs)
	if err != nil {
		return fmt.Errorf("error when marshaling attrs: %w", err)
	}

	if string(encodedAttrs) == "{}" {
		encodedAttrs = []byte{}
	}

	fmt.Println(
		colorize(lightGray, r.Time.Format(timeFormat)),
		terminalLevelLabel(r.Level),
		colorize(white, fmt.Sprintf("%s:", service)),
		colorize(white, r.Message),
		colorize(darkGray, string(encodedAttrs)),
	)

	return nil
}

// terminalLevelLabel maps slog's ordered levels to the small terminal palette. Custom levels below debug intentionally
// share debug's muted color so high-frequency trace output does not overpower warnings and errors.
func terminalLevelLabel(level slog.Level) string {
	label := level.String() + ":"

	switch {
	case level <= slog.LevelDebug:
		return colorize(darkGray, label)
	case level == slog.LevelInfo:
		return colorize(cyan, label)
	case level == slog.LevelWarn:
		return colorize(lightYellow, label)
	case level >= slog.LevelError:
		return colorize(lightRed, label)
	default:
		return label
	}
}

// cloneAttrs gives observers ownership of the delivered map. Values remain shallow because slog has already resolved
// them and the handler never mutates nested values.
func cloneAttrs(attributes map[string]any) map[string]any {
	result := make(map[string]any, len(attributes))
	for key, value := range attributes {
		result[key] = value
	}

	return result
}

// computeAttrs asks slog's JSON handler to perform attribute replacement and grouping, then decodes only the resolved
// attributes. The mutex covers both buffer use and reset because every derived Handler shares this scratch buffer.
func (h *Handler) computeAttrs(
	ctx context.Context,
	r slog.Record,
) (map[string]any, error) {
	h.m.Lock()
	defer func() {
		h.b.Reset()
		h.m.Unlock()
	}()

	if err := h.h.Handle(ctx, r); err != nil {
		return nil, fmt.Errorf("error when calling inner handler's Handle: %w", err)
	}

	var attrs map[string]any

	err := json.Unmarshal(h.b.Bytes(), &attrs)
	if err != nil {
		return nil, fmt.Errorf("error when unmarshaling inner handler's Handle result: %w", err)
	}

	return attrs, nil
}

// suppressDefaults removes the JSON handler's standard envelope fields because the terminal layout renders them
// separately. A caller-provided ReplaceAttr still receives all application attributes in their original groups.
func suppressDefaults(
	next func([]string, slog.Attr) slog.Attr,
) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey ||
			a.Key == slog.LevelKey ||
			a.Key == slog.MessageKey {
			return slog.Attr{}
		}

		if next == nil {
			return a
		}

		return next(groups, a)
	}
}
