package appinsights_test

import (
	"context"
	"log/slog"
	"maps"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/ApplicationInsights-Go/appinsights/contracts"
	"github.com/openclosed-dev/slogan/appinsights"
)

func TestLogMessage(t *testing.T) {

	server := newStubServer(8)
	defer server.Close()

	connectionString := server.connectionString()

	opts := appinsights.NewHandlerOptions(slog.LevelDebug)
	opts.Client = server.Client()

	ctx := context.Background()

	cases := []struct {
		name          string
		level         slog.Level
		message       string
		severityLevel contracts.SeverityLevel
	}{
		{"info", slog.LevelInfo, "info message", contracts.Information},
		{"warn", slog.LevelWarn, "warn message", contracts.Warning},
		{"error", slog.LevelError, "error message", contracts.Error},
		{"debug", slog.LevelDebug, "debug message", contracts.Verbose},
		//
		{"fatal", appinsights.LevelFatal, "fatal message", 4},
		{"critical", appinsights.LevelCritical, "critical message", 4},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			handler, err := appinsights.NewHandler(connectionString, opts)
			if err != nil {
				t.Fatalf("failed to create handler: %v", err)
			}

			logger := slog.New(handler)
			logger.Log(ctx, c.level, c.message)

			handler.Close()

			item := server.getTelemetry()

			if item.IKey != instrumentationKey {
				t.Errorf("incorrect instrumentation key: %s", item.IKey)
			}

			data := &item.Data.BaseData
			if data.Message != c.message {
				t.Errorf("incorrect message: %s", data.Message)
			}
			if data.SeverityLevel != int(c.severityLevel) {
				t.Errorf("incorrect level: %d", data.SeverityLevel)
			}
		})
	}
}

func TestLogLevel(t *testing.T) {

	server := newStubServer(8)
	defer server.Close()

	connectionString := server.connectionString()

	ctx := context.Background()

	allLevels := [5]contracts.SeverityLevel{
		contracts.Verbose,
		contracts.Information,
		contracts.Warning,
		contracts.Error,
		contracts.Critical,
	}

	cases := []struct {
		name     string
		minLevel slog.Leveler
		items    []contracts.SeverityLevel
	}{
		{
			"debug",
			slog.LevelDebug,
			allLevels[0:],
		},
		{
			"info",
			slog.LevelInfo,
			allLevels[1:],
		},
		{
			"warn",
			slog.LevelWarn,
			allLevels[2:],
		},
		{
			"error",
			slog.LevelError,
			allLevels[3:],
		},
		{
			"fatal",
			appinsights.LevelFatal,
			allLevels[4:],
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			opts := appinsights.NewHandlerOptions(c.minLevel)
			opts.Client = server.Client()

			handler, err := appinsights.NewHandler(connectionString, opts)
			if err != nil {
				t.Fatalf("failed to create handler: %v", err)
			}

			logger := slog.New(handler)
			logger.Log(ctx, slog.LevelDebug, "debug message")
			logger.Log(ctx, slog.LevelInfo, "info message")
			logger.Log(ctx, slog.LevelWarn, "warn message")
			logger.Log(ctx, slog.LevelError, "error message")
			logger.Log(ctx, appinsights.LevelFatal, "fatal message")

			handler.Close()

			items := server.telemetryItems()
			if len(items) != len(c.items) {
				t.Errorf("expected count is %d, but actual count was %d", len(c.items), len(items))
			}

			i := 0
			for _, item := range items {
				level := item.Data.BaseData.SeverityLevel
				if level != int(c.items[i]) {
					t.Errorf("expected level is %d, but actual level was %d", c.items[i], level)
				}
				i++
			}
		})
	}
}

func TestHandlerOptionsWithDefaultValues(t *testing.T) {

	server := newStubServer(8)
	defer server.Close()

	connectionString := server.connectionString()

	// HandlerOptions with default values
	opts := appinsights.HandlerOptions{}
	opts.Client = server.Client()

	handler, err := appinsights.NewHandler(connectionString, &opts)
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	logger := slog.New(handler)
	logger.Info("info message")

	handler.Close()

	items := server.telemetryItems()
	if len(items) != 1 {
		t.Errorf("too many telemetry items: %d", len(items))
	}

	data := items[0].Data.BaseData

	if data.SeverityLevel != int(contracts.Information) {
		t.Errorf("unexpected severity level: %d", data.SeverityLevel)
	}

	if data.Message != "info message" {
		t.Errorf("unexpected message: %s", data.Message)
	}
}

type attrMap map[string]string

type attrTestCase struct {
	name       string
	attrs      []slog.Attr
	properties attrMap
}

func (c *attrTestCase) args() []any {
	args := make([]any, 0, len(c.attrs)*2)
	for _, a := range c.attrs {
		args = append(args, a.Key, a.Value)
	}
	return args
}

func attrsCases() []attrTestCase {
	return []attrTestCase{
		{
			"string",
			[]slog.Attr{slog.String("key", "hello")},
			attrMap{
				"key": "hello",
			},
		},
		{
			"int",
			[]slog.Attr{slog.Int("key", 42)},
			attrMap{
				"key": "42",
			},
		},
		{
			"float",
			[]slog.Attr{slog.Float64("key", 3.14)},
			attrMap{
				"key": "3.14",
			},
		},
		{
			"true",
			[]slog.Attr{slog.Bool("key", true)},
			attrMap{
				"key": "true",
			},
		},
		{
			"false",
			[]slog.Attr{slog.Bool("key", false)},
			attrMap{
				"key": "false",
			},
		},
		{
			"time",
			[]slog.Attr{slog.Time("key", time.Unix(0, 0))},
			attrMap{
				"key": "1970-01-01T00:00:00Z",
			},
		},
		{
			"multiple",
			[]slog.Attr{slog.String("key1", "hello"), slog.Int("key2", 42)},
			attrMap{
				"key1": "hello",
				"key2": "42",
			},
		},
		{
			"no attribute",
			[]slog.Attr{},
			attrMap{},
		},
		{
			"empty attribute",
			[]slog.Attr{slog.Any("", nil)},
			attrMap{},
		},
		{
			"group",
			[]slog.Attr{
				slog.Group("group1", slog.String("key1", "hello"), slog.Int("key2", 42)),
			},
			attrMap{
				"group1.key1": "hello",
				"group1.key2": "42",
			},
		},
		{
			"group with no name",
			[]slog.Attr{
				slog.Group("", slog.String("key1", "hello")),
			},
			attrMap{
				"key1": "hello",
			},
		},
		{
			"empty group",
			[]slog.Attr{
				slog.Group("group1"),
			},
			attrMap{},
		},
	}
}

func TestLogMessageWithAttributes(t *testing.T) {

	server := newStubServer(8)
	defer server.Close()

	connectionString := server.connectionString()

	opts := appinsights.NewHandlerOptions(nil)
	opts.Client = server.Client()

	for _, c := range attrsCases() {
		t.Run(c.name, func(t *testing.T) {

			handler, err := appinsights.NewHandler(connectionString, opts)
			if err != nil {
				t.Fatalf("failed to create handler: %v", err)
			}

			logger := slog.New(handler)
			logger.Info("message", c.args()...)

			handler.Close()

			item := server.getTelemetry()

			actual := item.properties()
			if !maps.Equal(actual, c.properties) {
				t.Errorf("expected attributes are %v, but got %v", c.properties, actual)
			}
		})
	}
}

func TestWithAttrs(t *testing.T) {

	server := newStubServer(8)
	defer server.Close()

	connectionString := server.connectionString()

	opts := appinsights.NewHandlerOptions(nil)
	opts.Client = server.Client()

	for _, c := range attrsCases() {
		t.Run(c.name, func(t *testing.T) {

			handler, err := appinsights.NewHandler(connectionString, opts)
			if err != nil {
				t.Fatalf("failed to create handler: %v", err)
			}

			logger := slog.New(handler)
			logger = logger.With(c.args()...)
			logger.Info("message")

			handler.Close()

			item := server.getTelemetry()

			actual := item.properties()
			if !maps.Equal(actual, c.properties) {
				t.Errorf("expected attributes are %v, but got %v", c.properties, actual)
			}
		})
	}
}

func TestWithGroup(t *testing.T) {

	server := newStubServer(8)
	defer server.Close()

	opts := appinsights.NewHandlerOptions(nil)
	opts.Client = server.Client()

	handler, err := appinsights.NewHandler(server.connectionString(), opts)
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	logger := slog.New(handler)
	logger = logger.WithGroup("group1")

	logger.Info("message", "key1", "hello", "key2", 42)

	handler.Close()

	item := server.getTelemetry()

	expected := map[string]string{
		"group1.key1": "hello",
		"group1.key2": "42",
	}

	actual := item.properties()
	if !maps.Equal(actual, expected) {
		t.Errorf("expected attributes are %v, but got %v", expected, actual)
	}
}

func TestWithEmptyGroup(t *testing.T) {

	server := newStubServer(8)
	defer server.Close()

	opts := appinsights.NewHandlerOptions(nil)
	opts.Client = server.Client()

	handler, err := appinsights.NewHandler(server.connectionString(), opts)
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	logger := slog.New(handler)
	logger = logger.WithGroup("")

	logger.Info("message", "key1", "hello", "key2", 42)

	handler.Close()

	item := server.getTelemetry()

	expected := map[string]string{
		"key1": "hello",
		"key2": "42",
	}

	actual := item.properties()
	if !maps.Equal(actual, expected) {
		t.Errorf("expected attributes are %v, but got %v", expected, actual)
	}
}

func TestInvalidConnectionString(t *testing.T) {
	cases := []struct {
		name             string
		connectionString string
		message          string
	}{
		{"empty", "", "connection string is empty"},
		{"missing instrumentation key", "IngestionEndpoint=https://example.org/", "instrumentation key is missing"},
		{"missing ingestion endpoint", "InstrumentationKey=f81d4fae-7dec-11d0-a765-00a0c91e6bf6", "ingestion endpoint is missing"},
		{"invalid ingestion endpoint", "InstrumentationKey=f81d4fae-7dec-11d0-a765-00a0c91e6bf6;IngestionEndpoint=://example.org", "ingestion endpoint is not a valid URL"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := appinsights.NewHandler(c.connectionString, nil)

			if err == nil {
				t.Error("must be error")
			} else if !strings.HasPrefix(err.Error(), c.message) {
				t.Errorf("wrong error message: %s", err.Error())
			}
		})
	}
}
