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

// Package collector collects data from Personal Weather Stations (PWS).
package collector

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/joshuasing/pws_exporter/internal/dns"
	"github.com/joshuasing/pws_exporter/internal/exporter"
)

// forwardDomains are domains that should be forwarded to the upstream DNS
// resolver. They are necessary for the function of the Weather Station.
//
// Queried domains that are not in this list are black holed (NXDOMAIN) to
// prevent unwanted connections.
var forwardDomains = []string{
	"time.nist.gov.",
	"time-nw.nist.gov.",
	"time-a.nist.gov.",
	"time-b.nist.gov.",
	"time.windows.com.",
}

// HandlerFunc handles a new measurement from a device.
type HandlerFunc func(deviceID string, dm *exporter.DeviceMeasurement)

// HTTPCollector is a collector that receives data over HTTP.
type HTTPCollector interface {
	Domains() []string
	RegisterRoutes(mux *http.ServeMux) error
}

// Config is the Collectors configuration.
type Config struct {
	ListenAddress    string
	DNSListenAddress string
	CollectorIP      string
	UpstreamResolver string
}

// Collectors manages collectors.
type Collectors struct {
	listenAddress    string
	dnsListenAddress string

	collectorIP      string
	upstreamResolver string

	httpCollectors []HTTPCollector

	running    atomic.Bool
	httpServer *http.Server
	dnsServer  *dns.Server
}

// NewCollectors returns a new collectors.
func NewCollectors(c Config) (*Collectors, error) {
	if c.CollectorIP == "" {
		ip, err := outboundIP()
		if err != nil {
			return nil, fmt.Errorf("could not determine collector IP address: %w", err)
		}
		c.CollectorIP = ip.String()
	}
	if c.ListenAddress == "" {
		c.ListenAddress = ":80"
	}

	return &Collectors{
		listenAddress:    c.ListenAddress,
		dnsListenAddress: c.DNSListenAddress,
		collectorIP:      c.CollectorIP,
		upstreamResolver: c.UpstreamResolver,
	}, nil
}

// RegisterHTTPCollectors registers HTTP collectors.
func (c *Collectors) RegisterHTTPCollectors(hc ...HTTPCollector) {
	c.httpCollectors = append(c.httpCollectors, hc...)
}

// Run starts the collectors.
func (c *Collectors) Run() error {
	if !c.running.CompareAndSwap(false, true) {
		return errors.New("already running")
	}
	defer c.running.CompareAndSwap(true, false)

	// Setup DNS server
	localDomains := make(map[string]string)
	for _, hc := range c.httpCollectors {
		for _, d := range hc.Domains() {
			localDomains[d+"."] = c.collectorIP
		}
	}
	c.dnsServer = dns.NewServer(dns.Config{
		UpstreamResolver: c.upstreamResolver,
		Records:          localDomains,
		ForwardDomains:   forwardDomains,
	})

	var errg errgroup.Group

	// Start DNS server
	if c.dnsListenAddress != "" {
		errg.Go(func() error {
			slog.Info("DNS server listening",
				slog.String("address", c.dnsListenAddress))
			return c.dnsServer.ListenAndServe(c.dnsListenAddress)
		})
	}

	// Start HTTP server
	if c.listenAddress != "" {
		mux := http.NewServeMux()
		for _, hc := range c.httpCollectors {
			if err := hc.RegisterRoutes(mux); err != nil {
				return fmt.Errorf("register HTTP collector routes: %w", err)
			}
		}

		// TODO: Currently we only setup one HTTP server on the main address,
		//  which should be okay for most collectors. If someone wants to run
		//  multiple collectors, which require different ports
		//  (e.g. 80, 443, 3000), we may need more than one server or they
		//  would need to run multiple pws_exporter instances.
		c.httpServer = &http.Server{
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		}
		errg.Go(func() error {
			ln, err := net.Listen("tcp", c.listenAddress)
			if err != nil {
				return err
			}
			slog.Info("Collector HTTP server listening",
				slog.String("address", c.listenAddress))
			return c.httpServer.Serve(ln)
		})
	}

	return errg.Wait()
}

// Close closes the collectors.
func (c *Collectors) Close() error {
	if !c.running.Load() {
		return errors.New("not running")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var errg errgroup.Group
	errg.Go(func() error {
		if err := c.httpServer.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})
	errg.Go(func() error {
		return c.dnsServer.Shutdown(ctx)
	})

	return errg.Wait()
}

// outboundIP returns the local outbound address of the machine.
// This is used for attempting to guess the exporter IP address when it is not
// explicitly configured.
func outboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	//nolint:errcheck // Error is intentionally ignored.
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP, nil
}
