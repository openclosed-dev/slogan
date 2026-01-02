package appinsights_test

import (
	"context"
	"log/slog"
	"maps"
	"strings"
	"testing"
	"time"

	"github.com/openclosed-dev/slogan/appinsights"
)

func TestLogMessage(t *testing.T) {

	server := newStubServer()
	defer server.Close()

	connectionString := server.connectionString()

	opts := appinsights.NewHandlerOptions(slog.LevelDebug)
	opts.Client = server.Client()

	ctx := context.Background()

	cases := []struct {
		name          string
		level         slog.Level
		message       string
		severityLevel int
	}{
		{"info", slog.LevelInfo, "info message", 1},
		{"warn", slog.LevelWarn, "warn message", 2},
		{"error", slog.LevelError, "error message", 3},
		{"debug", slog.LevelDebug, "debug message", 0},
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
			if data.SeverityLevel != c.severityLevel {
				t.Errorf("incorrect level: %d", data.SeverityLevel)
			}
		})
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

	server := newStubServer()
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

	server := newStubServer()
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

	server := newStubServer()
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

	server := newStubServer()
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
