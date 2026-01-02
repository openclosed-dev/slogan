/*
Package appinsights provides a slog Handler to submit log records to Application Insights.
See more at https://azure.microsoft.com/en-us/services/application-insights/

# Getting Started

Install this module using the following command.

	go get github.com/openclosed-dev/slogan

# Usage

The code below shows how to use the handler in this package.

	import (
		"fmt"
		"log/slog"
		"os"

		"github.com/openclosed-dev/slogan/appinsights"
	)

	func main() {

		// The connection string provided by your Application Insights resource.
		connectionString := os.Getenv("APPLICATIONINSIGHTS_CONNECTION_STRING")

		// Creates a handler for Application Insights using the default options.
		handler, err := appinsights.NewHandler(connectionString, nil)
		if err != nil {
			fmt.Print(err)
			return
		}
		// The handler should be closed to flush log records before the application exits.
		defer handler.Close()

		// Creates a slog.Logger with the handler.
		logger := slog.New(handler)
		// Makes the logger as the system-wide default if you prefer.
		slog.SetDefault(logger)

		slog.Info("hello", "count", 3)
	}
*/
package appinsights
