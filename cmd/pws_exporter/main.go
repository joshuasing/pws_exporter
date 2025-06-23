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
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"

	"github.com/joshuasing/pws_exporter/internal/collector"
	"github.com/joshuasing/pws_exporter/internal/collector/wu"
	"github.com/joshuasing/pws_exporter/internal/exporter"
	"github.com/joshuasing/pws_exporter/internal/exporter/homeassistant"
	"github.com/joshuasing/pws_exporter/internal/exporter/prom"
)

const defaultPromListenAddress = ":9452"

const cmdHelp = `pws_exporter exports sensor data from Personal Weather Stations.

Flags:
`

type options struct {
	logLevel string

	// Prometheus exporter
	promListenAddress string

	// Home Assistant exporter
	homeAssistantMQTTURL      string
	homeAssistantMQTTUsername string
	homeAssistantMQTTPassword string

	// HTTP collector
	httpListenAddress string
	dnsListenAddress  string
	enableWU          bool

	// DNS
	exporterIPAddress       string
	upstreamResolverAddress string
}

func parseFlags(opts *options) error {
	pf := pflag.NewFlagSet("pws_exporter", pflag.ExitOnError)

	pf.StringVar(&opts.logLevel, "log", "info", "Log level (options: debug, info, warn, error)")

	// Exporters
	pf.StringVar(&opts.promListenAddress, "prom-listen", defaultPromListenAddress, "Prometheus HTTP listen address")
	pf.StringVar(&opts.homeAssistantMQTTURL, "ha-mqtt-url", "", "Home Assistant MQTT broker URL")
	pf.StringVar(&opts.homeAssistantMQTTUsername, "ha-mqtt-user", "", "Home Assistant MQTT broker username")
	pf.StringVar(&opts.homeAssistantMQTTPassword, "ha-mqtt-pass", "", "Home Assistant MQTT broker password")

	// Collectors
	pf.StringVar(&opts.httpListenAddress, "collector-listen", ":8080", "HTTP collector listen address")
	pf.StringVar(&opts.dnsListenAddress, "dns-listen", "", "DNS server listen address")
	pf.BoolVar(&opts.enableWU, "wu", true, "Whether to support Weather Underground API submission")

	pf.StringVar(&opts.exporterIPAddress, "exporter-ip", "", "Exporter IP address")
	pf.StringVar(&opts.upstreamResolverAddress, "resolver", "8.8.8.8:53", "Upstream DNS resolver address (host:port)")

	var help bool
	pf.BoolVarP(&help, "help", "h", false, "Displays help message")
	pf.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "%s", cmdHelp)
		pf.PrintDefaults()
	}

	// Parse flags
	if err := pf.Parse(os.Args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	if help {
		pf.Usage()
		return pflag.ErrHelp
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		slog.Error("An error occurred", slog.Any("err", err))
		os.Exit(1)
	}
}

func run() error {
	var opts options
	if err := parseFlags(&opts); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			return nil
		}
		return fmt.Errorf("parse flags: %w", err)
	}

	lvl, err := parseLogLevel(opts.logLevel)
	if err != nil {
		return fmt.Errorf("parse log level: %w", err)
	}
	slog.SetLogLoggerLevel(lvl)

	slog.Info("Starting PWS exporter")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Setup exporters
	exporters, handler, err := setupExporters(ctx, &opts)
	if err != nil {
		return fmt.Errorf("setup exporters: %w", err)
	}

	// Setup collectors
	c, err := collector.NewCollectors(collector.Config{
		ListenAddress:    opts.httpListenAddress,
		DNSListenAddress: opts.dnsListenAddress,
		CollectorIP:      opts.exporterIPAddress,
		UpstreamResolver: opts.upstreamResolverAddress,
	})
	if err != nil {
		return fmt.Errorf("new collectors: %w", err)
	}

	// Setup WU collector
	if opts.enableWU {
		slog.Info("Using WU API collector")
		c.RegisterHTTPCollectors(wu.NewCollector(handler))
	}

	// Start collectors
	var errg errgroup.Group
	errg.Go(c.Run)

	// Start exporters
	for _, ex := range exporters {
		errg.Go(ex.Run)
	}

	errc := make(chan error)
	go func() {
		errc <- errg.Wait()
		close(errc)
	}()

	var exitErr error
	select {
	case <-ctx.Done():
	case exitErr = <-errc:
		slog.Error("An error occurred, shutting down...",
			slog.Any("err", exitErr))
	}

	// Close everything.
	if err = c.Close(); err != nil {
		slog.Error("Failed to close collectors", slog.Any("err", err))
	}
	for _, ex := range exporters {
		if err = ex.Close(); err != nil {
			slog.Error("Failed to close exporter",
				slog.String("exporter_id", ex.ExporterID()),
				slog.Any("err", err))
		}
	}

	if exitErr != nil {
		return exitErr
	}
	return <-errc
}

func setupExporters(_ context.Context, opts *options) ([]exporter.Exporter, collector.HandlerFunc, error) {
	var exporters []exporter.Exporter

	// Setup Prometheus exporter
	if opts.promListenAddress != "" {
		slog.Info("Using Prometheus exporter")
		promEx, err := prom.NewExporter(prom.Config{
			ListenAddress: opts.promListenAddress,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("new prometheus exporter: %w", err)
		}

		exporters = append(exporters, promEx)
	}

	// Setup Home Assistant exporter
	if opts.homeAssistantMQTTURL != "" {
		slog.Info("Using Home Assistant MQTT exporter")
		haEx, err := homeassistant.NewExporter(homeassistant.Config{
			MQTTBrokerURL: opts.homeAssistantMQTTURL,
			MQTTUsername:  opts.homeAssistantMQTTUsername,
			MQTTPassword:  opts.homeAssistantMQTTPassword,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("new home assistant exporter: %w", err)
		}

		exporters = append(exporters, haEx)
	}

	return exporters, newCollectorHandler(exporters), nil
}

func newCollectorHandler(exporters []exporter.Exporter) collector.HandlerFunc {
	return func(deviceID string, dm *exporter.DeviceMeasurement) {
		for _, ex := range exporters {
			if err := ex.HandleDeviceMeasurement(deviceID, dm); err != nil {
				slog.Error("Failed to handle device measurement",
					slog.String("exporter_id", ex.ExporterID()),
					slog.String("device_id", deviceID))
			}
		}
	}
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
