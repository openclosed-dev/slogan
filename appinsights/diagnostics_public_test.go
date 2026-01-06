package appinsights_test

import (
	"log/slog"
	"testing"

	"github.com/openclosed-dev/slogan/appinsights"
)

func TestEnableDiagnostics(t *testing.T) {

	appinsights.EnableDiagnostics()

	server := newStubServer(8)
	defer server.Close()

	opts := appinsights.NewHandlerOptions(nil)
	opts.Client = server.Client()

	handler, err := appinsights.NewHandler(server.connectionString(), opts)
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}
	defer handler.Close()

	logger := slog.New(handler)
	logger.Info("a message")
}
