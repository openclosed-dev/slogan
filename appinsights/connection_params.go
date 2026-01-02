package appinsights

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type connectionParams struct {
	instrumentationKey string
	ingestionEndpoint  *url.URL
}

func parseConnectionString(connectionString string) (*connectionParams, error) {

	connectionString = strings.TrimSpace(connectionString)
	if connectionString == "" {
		return nil, errors.New("connection string is empty")
	}

	var instrumentationKey string
	var ingestionEndpoint string

	for _, v := range strings.Split(connectionString, ";") {
		pair := strings.SplitN(v, "=", 2)
		if len(pair) == 2 {
			switch pair[0] {
			case "InstrumentationKey":
				instrumentationKey = pair[1]
			case "IngestionEndpoint":
				ingestionEndpoint = pair[1]
			}
		}
	}

	if instrumentationKey == "" {
		return nil, errors.New("instrumentation key is missing")
	}

	if ingestionEndpoint == "" {
		return nil, errors.New("ingestion endpoint is missing")
	}

	ingestionUrl, err := url.Parse(ingestionEndpoint)
	if err != nil {
		return nil, fmt.Errorf("ingestion endpoint is not a valid URL: %w", err)
	}

	return &connectionParams{
		instrumentationKey: instrumentationKey,
		ingestionEndpoint:  ingestionUrl,
	}, nil
}
