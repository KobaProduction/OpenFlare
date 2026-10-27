// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
)

func TestAgentTelemetryExportsAuthenticatedTrace(t *testing.T) {
	type request struct{ path, authorization string }
	received := make(chan request, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		received <- request{path: r.URL.Path, authorization: r.Header.Get("Authorization")}
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", "")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", server.URL)
	t.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "Authorization=Bearer%20test-agent-token")
	t.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.namespace=OpenFlare,deployment.environment.name=production")
	t.Setenv("OTEL_SERVICE_NAME", "OpenFlare Node")
	t.Setenv("OTEL_SAMPLING_RATE", "1")

	shutdown := initAgentTelemetry()
	_, span := otel.Tracer("agent").Start(context.Background(), "agent.test")
	span.End()
	shutdown()
	select {
	case actual := <-received:
		if actual.path != "/v1/traces" || actual.authorization != "Bearer test-agent-token" {
			t.Fatalf("unexpected OTLP request: path=%q bearer_present=%t", actual.path, actual.authorization == "Bearer test-agent-token")
		}
	default:
		t.Fatal("agent produced no OTLP/HTTP trace")
	}
}
