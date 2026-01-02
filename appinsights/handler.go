package appinsights

import (
	"context"
	"log/slog"
	"maps"
	"net/http"
	"time"

	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	"github.com/microsoft/ApplicationInsights-Go/appinsights/contracts"
)

const (
	// LevelCritical is the log level corresponding to the Critical level
	// in Application Insights
	LevelCritical = slog.Level(12)
	// LevelFatal is an alias of [LevelCritical]
	LevelFatal = LevelCritical
)

const (
	defaultLogLevel       = slog.LevelInfo
	ingestionEndpointPath = "/v2/track"
)

// HandlerOptions are options for a [Handler].
type HandlerOptions struct {
	// Level reports the minimum record level that will be logged.
	Level slog.Leveler
	// MaxBatchSize is the maximum number of log records.
	// that can be submitted in a request.
	MaxBatchSize int
	// MaxBatchInterval is the maximum time to wait before sending a batch of log records.
	MaxBatchInterval time.Duration
	// Client is a customized HTTP client.
	Client *http.Client
}

// Handler is a [slog.Handler] that submits log records to
// Azure Application Insights.
type Handler struct {
	opts   HandlerOptions
	client appinsights.TelemetryClient
	level  slog.Leveler
	// keyPrefix is empty or otherwise ends with period.
	keyPrefix  string
	attributes map[string]string
}

// NewHandlerOptions creates a [HandlerOptions]
// that contains reasonable default values.
func NewHandlerOptions(level slog.Leveler) *HandlerOptions {
	if level == nil {
		level = defaultLogLevel
	}
	return &HandlerOptions{
		Level:            level,
		MaxBatchSize:     1024,
		MaxBatchInterval: time.Duration(10) * time.Second,
	}
}

// NewHandler creates a [Handler] that submits log records to
// Azure Application Insights resource.
// The argument connectionString must be a valid connection string
// given by the target Application Insights resource.
// If opts is nil, the default options are used.
func NewHandler(connectionString string, opts *HandlerOptions) (*Handler, error) {
	if opts == nil {
		opts = NewHandlerOptions(defaultLogLevel)
	}

	var params, err = parseConnectionString(connectionString)
	if err != nil {
		return nil, err
	}

	return &Handler{
		opts:       *opts,
		client:     newTelemetryClient(params, opts),
		level:      opts.Level,
		attributes: make(map[string]string),
	}, nil
}

// Enabled reports whether the handler handles records at the given level.
// The handler ignores records whose level is lower.
func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

// Handle handles the log Record.
func (h *Handler) Handle(_ context.Context, r slog.Record) error {

	item := appinsights.NewTraceTelemetry(r.Message, mapLogLevel(r.Level))

	if !r.Time.IsZero() {
		item.Timestamp = r.Time
	}

	maps.Copy(item.Properties, h.attributes)

	r.Attrs(func(a slog.Attr) bool {
		addAttributeToMap(item.Properties, h.keyPrefix, a)
		return true
	})

	h.client.Track(item)

	return nil
}

// WithAttrs returns a new [Handler] whose attributes consists
// of h's attributes followed by attrs.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		// Note: slog.Logger does not pass in empty slice
		return h
	}

	newSize := len(h.attributes) + len(attrs)
	newAttributes := make(map[string]string, newSize)
	maps.Copy(newAttributes, h.attributes)
	for _, a := range attrs {
		addAttributeToMap(newAttributes, h.keyPrefix, a)
	}

	return &Handler{
		opts:       h.opts,
		client:     h.client,
		level:      h.level,
		keyPrefix:  h.keyPrefix,
		attributes: newAttributes,
	}
}

// WithGroup returns a new [Handler] with the given group appended to
// the receiver's existing groups.
// The keys of all subsequent attributes, whether added by With or in a
// Record, will be qualified by the sequence of group names.
func (h *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		// Note: log.Logger does not pass in empty name
		return h
	}

	newKeyPrefix := h.keyPrefix + name + "."

	return &Handler{
		opts:       h.opts,
		client:     h.client,
		level:      h.level,
		keyPrefix:  newKeyPrefix,
		attributes: maps.Clone(h.attributes),
	}
}

// Close flushes the buffered log records
// and waits until the transmission is complete.
func (h *Handler) Close() {
	if client := h.client; client != nil {
		select {
		case <-client.Channel().Close(10 * time.Second):
		case <-time.After(30 * time.Second):
		}
	}
}

func newTelemetryClient(params *connectionParams, opts *HandlerOptions) appinsights.TelemetryClient {

	var endpointUrl = *params.ingestionEndpoint
	endpointUrl.Path = ingestionEndpointPath

	var config = appinsights.TelemetryConfiguration{
		InstrumentationKey: params.instrumentationKey,
		EndpointUrl:        endpointUrl.String(),
		MaxBatchSize:       opts.MaxBatchSize,
		MaxBatchInterval:   opts.MaxBatchInterval,
		Client:             opts.Client,
	}

	return appinsights.NewTelemetryClientFromConfig(&config)
}

func mapLogLevel(level slog.Level) contracts.SeverityLevel {
	switch {
	case level <= slog.LevelDebug:
		return appinsights.Verbose
	case level >= LevelCritical:
		return appinsights.Critical
	case level >= slog.LevelError:
		return appinsights.Error
	case level >= slog.LevelWarn:
		return appinsights.Warning
	default:
		return appinsights.Information
	}
}

func addAttributeToMap(m map[string]string, keyPrefix string, a slog.Attr) {
	if a.Equal(slog.Attr{}) {
		return
	}

	var value string

	a.Value = a.Value.Resolve()

	switch a.Value.Kind() {
	case slog.KindTime:
		value = a.Value.Time().UTC().Format(time.RFC3339Nano)
	case slog.KindGroup:
		addAttributeGroupToMap(m, keyPrefix, a)
		return
	default:
		value = a.Value.String()
	}

	if value != "" {
		m[keyPrefix+a.Key] = value
	}
}

func addAttributeGroupToMap(m map[string]string, keyPrefix string, g slog.Attr) {
	attrs := g.Value.Group()
	if len(attrs) == 0 {
		return
	}

	if g.Key != "" {
		keyPrefix += g.Key + "."
	}

	for _, a := range attrs {
		addAttributeToMap(m, keyPrefix, a)
	}
}
