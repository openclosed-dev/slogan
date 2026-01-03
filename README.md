# slogan

[![Documentation](https://godoc.org/github.com/openclosed-dev/slogan?status.svg)](https://godoc.org/github.com/openclosed-dev/slogan/appinsights) [![Release](https://img.shields.io/github/release/openclosed-dev/slogan/all.svg)](https://github.com/openclosed-dev/slogan/releases)

slogan is a Go module that provides a [slog](https://pkg.go.dev/log/slog) Handler implementation for submitting log records to [Azure Application Insights](https://learn.microsoft.com/en-us/azure/azure-monitor/app/app-insights-overview).

## Prerequisites

This module requires [Go](https://go.dev) version [1.23](https://go.dev/doc/devel/release#go1.23.0) or higher.

## Getting Started

Install this module using the following command.

```
go get github.com/openclosed-dev/slogan
```

## Usage

The code below shows how to use the handler in `appinsights` package.

```go
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
```

## Known Issue

This module depends on the module [github.com/microsoft/ApplicationInsights-Go](https://pkg.go.dev/github.com/microsoft/ApplicationInsights-Go) developed by Microsoft, that is not actively maintained for several years.
While there seems to be no functional problem in the module, we need to constantly check the status of the project.
