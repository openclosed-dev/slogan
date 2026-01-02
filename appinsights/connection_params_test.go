package appinsights

import "testing"

func TestParseValidConnectionString(t *testing.T) {

	connectionString := "InstrumentationKey=f81d4fae-7dec-11d0-a765-00a0c91e6bf6;IngestionEndpoint=https://southcentralus.in.applicationinsights.azure.com/"

	spec, err := parseConnectionString(connectionString)
	if err != nil {
		t.Errorf("failed to parse valid connection string: %v", err)
	}

	if spec.instrumentationKey != "f81d4fae-7dec-11d0-a765-00a0c91e6bf6" {
		t.Errorf("instrument key is wrong: %s", spec.instrumentationKey)
	}

	if spec.ingestionEndpoint == nil {
		t.Errorf("ingestion endpoint is nil")
	}

	if spec.ingestionEndpoint.String() != "https://southcentralus.in.applicationinsights.azure.com/" {
		t.Errorf("ingestion endpoint is wrong: %v", spec.ingestionEndpoint)
	}
}
