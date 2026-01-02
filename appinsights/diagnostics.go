package appinsights

import (
	"fmt"
	"time"

	"github.com/microsoft/ApplicationInsights-Go/appinsights"
)

// EnableDiagnostics enables telemetry printing for debugging purpose.
func EnableDiagnostics() {
	appinsights.NewDiagnosticsMessageListener(printDiagnosticsMessage)
}

func printDiagnosticsMessage(message string) error {
	fmt.Printf("[%s] %s\n", time.Now().Format(time.RFC3339), message)
	return nil
}
