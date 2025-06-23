// Copyright (c) 2025 Joshua Sing <joshua@Joshuasing.dev>
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

// Package prom implements an exporter that exposes sensor data as Prometheus
// metrics and runs an HTTP server to serve the metrics.
package prom

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/joshuasing/pws_exporter/internal/exporter"
)

const exporterID = "prometheus"

// Config is the Exporter configuration.
type Config struct {
	ListenAddress string
}

// Exporter is an exporter that exposes sensor data as Prometheus metrics.
type Exporter struct {
	listenAddress string

	metrics    *metrics
	registry   *prometheus.Registry
	httpServer *http.Server

	running atomic.Bool
}

// NewExporter returns a new Prometheus exporter.
func NewExporter(conf Config) (*Exporter, error) {
	reg := prometheus.NewRegistry()
	return &Exporter{
		listenAddress: conf.ListenAddress,
		metrics:       newMetrics(reg),
		registry:      reg,
	}, nil
}

// Run starts the Prometheus HTTP server.
func (e *Exporter) Run() error {
	if !e.running.CompareAndSwap(false, true) {
		return errors.New("already running")
	}
	defer e.running.CompareAndSwap(true, false)

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(e.registry, promhttp.HandlerOpts{}))
	e.httpServer = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ln, err := net.Listen("tcp", e.listenAddress)
	if err != nil {
		return err
	}
	slog.Info("Prometheus metrics server listening",
		slog.String("address", e.listenAddress))
	return e.httpServer.Serve(ln)
}

// Close closes the collectors.
func (e *Exporter) Close() error {
	if !e.running.Load() {
		return errors.New("not running")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := e.httpServer.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// ExporterID returns the exporter ID.
func (e *Exporter) ExporterID() string {
	return exporterID
}

// HandleDeviceMeasurement handles a new measurement from a device.
func (e *Exporter) HandleDeviceMeasurement(deviceID string, dm *exporter.DeviceMeasurement) error {
	m := e.metrics
	l := prometheus.Labels{"station_id": deviceID}

	m.BarometricPressure.With(l).Set(float64(dm.BarometricPressure))
	m.DewPoint.With(l).Set(float64(dm.DewPoint))
	m.Humidity.With(l).Set(float64(dm.Humidity / 100))
	m.IndoorHumidity.With(l).Set(float64(dm.IndoorHumidity / 100))
	m.IndoorTemperature.With(l).Set(float64(dm.IndoorTemp))
	m.RainPastHour.With(l).Set(float64(dm.RainPastHour))
	m.Rain.Delete(l) // Counter state is stored on the station, not in the exporter.
	m.Rain.With(l).Add(float64(dm.RainToday))
	m.Temperature.With(l).Set(float64(dm.Temperature))
	m.WindDirection.With(l).Set(float64(dm.WindDirection))
	m.WindGustSpeed.With(l).Set(float64(dm.WindGustSpeed))
	m.WindSpeed.With(l).Set(float64(dm.WindSpeed))

	return nil
}
