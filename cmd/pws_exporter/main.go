// Copyright (c) 2025 Joshua Sing <joshua@joshuasing.dev>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/joshuasing/pws_exporter/internal/exporter"
)

const defaultListenAddress = ":9452"

var (
	logLevel           = flag.String("log", "info", "Log level")
	listenAddress      = flag.String("listen", defaultListenAddress, "Listen address")
	exporterAddress    = flag.String("exporter", "", "Exporter IP address")
	upstreamResolver   = flag.String("resolver", "8.8.8.8:53", "Upstream DNS resolver")
	dnsListenAddress   = flag.String("dns-listen", "", "DNS server listen address")
	wuListenAddress    = flag.String("wu-listen", ":80", "WU HTTP server listen address")
	wuTLSListenAddress = flag.String("wu-tls-listen", ":443", "WU HTTPS server listen address")
)

func main() {
	flag.Parse()
	os.Exit(run())
}

func run() int {
	lvl, err := parseLogLevel(*logLevel)
	if err != nil {
		slog.Error("Failed to parse log level", slog.Any("err", err))
		return 1
	}
	slog.SetLogLoggerLevel(lvl)

	slog.Info("Starting WU Weather Station exporter")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	ex, err := exporter.NewExporter(exporter.Config{
		ExporterIP:         *exporterAddress,
		UpstreamResolver:   *upstreamResolver,
		DNSListenAddress:   *dnsListenAddress,
		WUListenAddress:    *wuListenAddress,
		WUTLSListenAddress: *wuTLSListenAddress,
	})
	if err != nil {
		slog.Error("Failed to create exporter", slog.Any("err", err))
		return 1
	}

	exErr := make(chan error)
	go func() {
		exErr <- ex.ListenAndServe()
	}()

	// Metrics handler
	http.Handle("/metrics", promhttp.HandlerFor(ex.Registry(), promhttp.HandlerOpts{}))

	// Run HTTP server in a goroutine
	httpErr := make(chan error)
	go func() {
		srv := http.Server{
			Addr:              *listenAddress,
			ReadHeaderTimeout: 5 * time.Second,
		}
		slog.Info("Metrics HTTP server listening", slog.String("address", srv.Addr))
		httpErr <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
	case err = <-exErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to start exporter", slog.Any("err", err))
		}
		return 1
	case err = <-httpErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Failed to start HTTP server", slog.Any("err", err))
		}
		if eerr := ex.Close(); eerr != nil {
			slog.Error("Failed to close exporter", slog.Any("err", eerr))
			return 1
		}
		if err != nil {
			return 1
		}
	}

	return 0
}

func parseLogLevel(level string) (slog.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelError, fmt.Errorf("invalid log level: %s", level)
	}
}
