// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Rain-kl/Wavelet/pkg/trace"
)

// initAgentTelemetry enables tracing only when an OTLP endpoint is supplied.
// Existing installations without collector configuration continue unmodified.
func initAgentTelemetry() func() {
	if strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")) == "" {
		return func() {}
	}
	serviceName := strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME"))
	if serviceName == "" {
		serviceName = "OpenFlare Node"
	}
	samplingRate := 0.1
	if raw := strings.TrimSpace(os.Getenv("OTEL_SAMPLING_RATE")); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil && v >= 0 && v <= 1 {
			samplingRate = v
		}
	}
	trace.Init(trace.Config{AppName: serviceName, SamplingRate: samplingRate, TracerName: "agent"})
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		trace.Shutdown(ctx)
	}
}
