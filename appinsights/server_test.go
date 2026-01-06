package appinsights_test

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

// fake instrumentation key
const instrumentationKey = "f81d4fae-7dec-11d0-a765-00a0c91e6bf6"

// Trace telemetry item collected by Application Insights
type telemetry struct {
	IKey string `json:"iKey"`
	Data struct {
		BaseType string `json:"baseType"`
		BaseData struct {
			Ver           int               `json:"ver"`
			Message       string            `json:"message"`
			SeverityLevel int               `json:"severityLevel"`
			Properties    map[string]string `json:"properties"`
		} `json:"baseData"`
	} `json:"data"`
}

func (t *telemetry) properties() map[string]string {
	return t.Data.BaseData.Properties
}

type stubServer struct {
	*httptest.Server
	items chan *telemetry
}

func newStubServer(capacity int) *stubServer {
	mux := http.NewServeMux()

	s := &stubServer{
		httptest.NewServer(mux),
		make(chan *telemetry, capacity),
	}

	mux.Handle("/v2/track", s)

	return s
}

// connectionString returns the connection string for the server.
func (s *stubServer) connectionString() string {
	return fmt.Sprintf(
		"InstrumentationKey=%s;IngestionEndpoint=%s;",
		instrumentationKey, s.URL,
	)
}

func (s *stubServer) getTelemetry() *telemetry {
	return <-s.items
}

func (s *stubServer) telemetryItems() []*telemetry {

	count := len(s.items)
	items := make([]*telemetry, 0, count)

	for count > 0 {
		items = append(items, <-s.items)
		count--
	}

	return items
}

func (s *stubServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	items, err := decodeRequestBody(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, item := range items {
		s.items <- item
	}

	w.WriteHeader(http.StatusOK)
}

func decodeRequestBody(req *http.Request) ([]*telemetry, error) {

	if req.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		return readTelemetryItems(gzipReader)
	} else {
		return readTelemetryItems(req.Body)
	}
}

func readTelemetryItems(reader io.Reader) ([]*telemetry, error) {

	decoder := json.NewDecoder(reader)

	items := make([]*telemetry, 0, 1)
	for {
		var item telemetry
		err := decoder.Decode(&item)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to decode JSON: %w", err)
		}

		items = append(items, &item)
	}
	return items, nil
}
