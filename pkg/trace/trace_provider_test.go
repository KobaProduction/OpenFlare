// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	collecttrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/proto"
)

func TestOTLPHTTPTraceExportBearerAndResources(t *testing.T) {
	type received struct {
		header string
		data   []byte
		path   string
		method string
	}
	requests := make(chan received, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		requests <- received{header: r.Header.Get("Authorization"), data: body, path: r.URL.Path, method: r.Method}
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(protoBytes(t, &collecttrace.ExportTraceServiceResponse{}))
	}))
	defer server.Close()
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", server.URL)
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "Authorization=Bearer%20test-secret")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.namespace=intens,deployment.environment.name=production")

	provider, err := newTracerProvider(Config{AppName: "openflare-control-plane", SamplingRate: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = provider.Shutdown(context.Background()) }()
	_, span := provider.Tracer("test").Start(context.Background(), "canary")
	span.End()
	if err := provider.ForceFlush(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case req := <-requests:
		if req.header != "Bearer test-secret" {
			t.Errorf("invalid bearer header value")
		}
		if req.method != "POST" || req.path != "/v1/traces" {
			t.Errorf("unexpected OTLP route %s %s", req.method, req.path)
		}
		var payload collecttrace.ExportTraceServiceRequest
		if err := proto.Unmarshal(req.data, &payload); err != nil {
			t.Fatal(err)
		}
		if len(payload.ResourceSpans) == 0 {
			t.Fatal("missing resource spans")
		}
		attrs := make(map[string]string)
		for _, a := range payload.ResourceSpans[0].Resource.Attributes {
			attrs[a.Key] = a.Value.GetStringValue()
		}
		for k, v := range map[string]string{"service.name": "openflare-control-plane", "service.namespace": "intens", "deployment.environment.name": "production"} {
			if attrs[k] != v {
				t.Errorf("resource attribute %s mismatch: got %q", k, attrs[k])
			}
		}
	default:
		t.Fatal("OTLP HTTP request was not received")
	}
}

func protoBytes(t *testing.T, v proto.Message) []byte {
	t.Helper()
	data, err := proto.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestUnknownOTLPProtocolFails(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/json")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "")
	_, err := newTraceExporter(context.Background())
	if err == nil || !strings.Contains(err.Error(), "unsupported OTLP") {
		t.Fatalf("expected protocol error, got %v", err)
	}
}
