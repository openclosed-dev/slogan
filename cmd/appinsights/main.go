package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/openclosed-dev/slogan/appinsights"
)

func main() {

	connectionString := os.Getenv("APPLICATIONINSIGHTS_CONNECTION_STRING")
	connectionString = strings.TrimSpace(connectionString)
	if connectionString == "" {
		fmt.Fprintln(os.Stderr,
			"Error: Environment variable APPLICATIONINSIGHTS_CONNECTION_STRING is not defined.")
		os.Exit(1)
	}

	handler, err := appinsights.NewHandler(connectionString, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"Error: Failed to create log handler: %v.\n", err)
		os.Exit(1)
	}
	defer handler.Close()

	slog.SetDefault(slog.New(handler))

	fmt.Println("This program sends the lines you type to Azure Application Insights.")
	fmt.Println("If you enter a blank line, the program will terminate normally.")

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		slog.Info(line)
	}
}
