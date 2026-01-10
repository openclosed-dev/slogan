package appinsights_test

import (
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"testing/slogtest"

	"github.com/openclosed-dev/slogan/appinsights"
)

func TestHandlerWithTestSuite(t *testing.T) {

	server := newStubServer(256)
	defer server.Close()

	opts := appinsights.NewHandlerOptions(slog.LevelInfo)
	opts.Client = server.Client()

	var handler *appinsights.Handler

	newHandler := func(t *testing.T) slog.Handler {
		var err error
		handler, err = appinsights.NewHandler(server.connectionString(), opts)
		if err != nil {
			t.Fatalf("failed to create handler: %v", err)
		}
		return handler
	}

	result := func(t *testing.T) map[string]any {
		handler.Close()
		items := server.telemetryItems()
		resultMap := convertTelemetryToMap(items[0])
		caseName := strings.Split(t.Name(), "/")[1]
		// For this specific case
		// we need to remove the time from the converted map.
		if caseName == "zero-time" {
			delete(resultMap, slog.TimeKey)
		}
		return resultMap
	}

	slogtest.Run(t, newHandler, result)
}

func convertTelemetryToMap(t *telemetry) map[string]any {
	resultMap := make(map[string]any)

	resultMap[slog.TimeKey] = t.Time
	resultMap[slog.MessageKey] = t.Data.BaseData.Message
	resultMap[slog.LevelKey] = convertSeverity(t.Data.BaseData.SeverityLevel)

	props := t.Data.BaseData.Properties
	for key, value := range props {
		names := strings.Split(key, ".")
		if len(names) <= 1 {
			resultMap[names[0]] = value
		} else {
			groupMap := resultMap
			for i, name := range names {
				if i == len(names)-1 {
					groupMap[name] = value
				} else {
					found, ok := groupMap[name]
					if !ok {
						found = make(map[string]any)
						groupMap[name] = found
					}
					groupMap = found.(map[string]any)
				}
			}
		}
	}

	return resultMap
}

func convertSeverity(severity int) slog.Leveler {
	switch severity {
	case 0:
		return slog.LevelDebug
	case 1:
		return slog.LevelInfo
	case 2:
		return slog.LevelWarn
	case 3:
		return slog.LevelError
	default:
		panic(fmt.Errorf("Unexpected severity value: %d", severity))
	}
}
