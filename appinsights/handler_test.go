package appinsights

import (
	"log/slog"
	"testing"
)

const (
	connectionString = "InstrumentationKey=f81d4fae-7dec-11d0-a765-00a0c91e6bf6;IngestionEndpoint=https://www.example.org/"
)

func TestWithEmptyAttrs(t *testing.T) {

	handler, err := NewHandler(connectionString, nil)
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	attrs := make([]slog.Attr, 0)
	var handler2 *Handler = handler.withAttrs(attrs)
	if handler2 != handler {
		t.Errorf("new group was created")
	}
}

func TestWithEmptyGroup(t *testing.T) {

	handler, err := NewHandler(connectionString, nil)
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	var handler2 *Handler = handler.withGroup("")
	if handler2 != handler {
		t.Errorf("new group was created")
	}
}
