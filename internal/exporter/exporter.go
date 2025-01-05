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

package exporter

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"

	"github.com/joshuasing/pws_exporter/internal/dns"
	"github.com/joshuasing/pws_exporter/internal/exporter/wu"
)

var (
	// wuDomains are domains used to submit data to the Weather Underground (WU)
	// API. The DNS resolver will answer A queries for these domains with the
	// exporter IP address.
	//
	// The self-signed TLS certificate used by the WU API server will be issued
	// to all of these domains.
	wuDomains = []string{
		"weatherstation.wunderground.com", // Standard submission API
		"rtupdate.wunderground.com",       // RapidFire (real-time) submission API
	}

	// forwardDomains are domains that should be forwarded to the upstream DNS
	// resolver. They are necessary for the function of the Weather Station.
	//
	// Queried domains that are not in this list are black holed (NXDOMAIN) to
	// prevent unwanted connections.
	forwardDomains = []string{
		"time.nist.gov.",
		"time-nw.nist.gov.",
		"time-a.nist.gov.",
		"time-b.nist.gov.",
		"time.windows.com.",
	}
)

type Exporter struct {
	exporterIP         string
	upstreamResolver   string
	dnsListenAddress   string
	wuListenAddress    string
	wuTLSListenAddress string

	running atomic.Bool

	registry *prometheus.Registry
	metrics  *Metrics

	dnsServer  *dns.Server
	httpServer *http.Server
}

type Config struct {
	ExporterIP         string
	UpstreamResolver   string
	DNSListenAddress   string
	WUListenAddress    string
	WUTLSListenAddress string
}

// NewExporter returns a new exporter.
func NewExporter(c Config) (*Exporter, error) {
	if c.ExporterIP == "" {
		ip, err := outboundIP()
		if err != nil {
			return nil, fmt.Errorf("could not determine exporter IP address: %w", err)
		}
		c.ExporterIP = ip.String()
	}
	if c.WUListenAddress == "" {
		c.WUListenAddress = ":80"
	}
	if c.WUTLSListenAddress == "" {
		c.WUTLSListenAddress = ":443"
	}

	reg := prometheus.NewRegistry()
	return &Exporter{
		exporterIP:         c.ExporterIP,
		upstreamResolver:   c.UpstreamResolver,
		dnsListenAddress:   c.DNSListenAddress,
		wuListenAddress:    c.WUListenAddress,
		wuTLSListenAddress: c.WUTLSListenAddress,
		registry:           reg,
		metrics:            newMetrics("weather", reg),
	}, nil
}

func (e *Exporter) Registry() *prometheus.Registry {
	return e.registry
}

// ListenAndServe starts the exporter and the DNS and HTTP servers.
func (e *Exporter) ListenAndServe() error {
	if !e.running.CompareAndSwap(false, true) {
		return errors.New("already running")
	}
	defer e.running.CompareAndSwap(true, false)

	// TLS configuration.
	var tlsConfig *tls.Config
	if e.wuTLSListenAddress != "" {
		// Generate temporary TLS certificate
		slog.Debug("Generating temporary self-signed TLS certificate")
		cert, err := genTLSCertificate()
		if err != nil {
			return fmt.Errorf("generate self signed certificate: %w", err)
		}
		slog.Debug("Generated self-signed TLS certificate")

		tlsConfig = &tls.Config{ //nolint:gosec
			// TLS v1.0 is used for compatibility reasons, as many weather
			// stations do not support modern TLS connections, or TLS at all.
			//
			// Whilst not ideal, traffic to the exporter should be entirely
			// local, limiting the risk of using an older and deprecated TLS
			// version.
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS10,
		}
	}

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.Handle(wu.SubmissionPath, wu.NewSubmissionAPI(e.handleWUSubmission))
	e.httpServer = &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		TLSConfig:         tlsConfig,
	}

	// Setup DNS server
	localDomains := make(map[string]string, len(wuDomains))
	for _, domain := range wuDomains {
		localDomains[domain+"."] = e.exporterIP
	}
	e.dnsServer = dns.NewServer(dns.Config{
		UpstreamResolver: e.upstreamResolver,
		Records:          localDomains,
		ForwardDomains:   forwardDomains,
	})

	var errg errgroup.Group

	// Start DNS server
	if e.dnsListenAddress != "" {
		errg.Go(func() error {
			slog.Info("DNS server listening",
				slog.String("address", e.dnsListenAddress))
			return e.dnsServer.ListenAndServe(e.dnsListenAddress)
		})
	}

	// Start HTTPS server
	errg.Go(func() error {
		ln, err := net.Listen("tcp", e.wuListenAddress)
		if err != nil {
			return err
		}
		slog.Info("WU API server listening",
			slog.String("address", e.wuListenAddress))
		return e.httpServer.Serve(ln)
	})
	if e.wuListenAddress != "" {
		errg.Go(func() error {
			ln, err := net.Listen("tcp", e.wuTLSListenAddress)
			if err != nil {
				return err
			}
			slog.Info("WU API TLS server listening",
				slog.String("address", ln.Addr().String()))
			return e.httpServer.ServeTLS(ln, "", "")
		})
	}

	return errg.Wait()
}

// Close shuts down the exporter.
func (e *Exporter) Close() error {
	if !e.running.Load() {
		// Nothing to do.
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var errg errgroup.Group
	errg.Go(func() error {
		return e.dnsServer.Shutdown(ctx)
	})
	errg.Go(func() error {
		return e.httpServer.Shutdown(ctx)
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
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP, nil
}

// genTLSCertificate generates a temporary self-signed in-memory TLS
// certificate with an RSA 2048-bit private key.
//
// This is not designed to be, nor needs to be secure, as it is only used for
// TLS connections between the Weather Station and the exporter's WU API server.
//
// This only works if the Weather Station accepts any TLS certificate, which
// appears to be the case most of the time.
func genTLSCertificate() (tls.Certificate, error) {
	var outCert tls.Certificate

	// Generate private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return outCert, fmt.Errorf("generate RSA 2048 private key: %s", err)
	}

	// Generate certificate serial number
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return outCert, fmt.Errorf("generate serial number: %s", err)
	}

	// Create certificate
	t := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"WU Weather Station Exporter"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              wuDomains,
	}
	cert, err := x509.CreateCertificate(rand.Reader, &t, &t, priv.Public(), priv)
	if err != nil {
		return outCert, fmt.Errorf("create certificate: %w", err)
	}

	outCert.Certificate = append(outCert.Certificate, cert)
	outCert.PrivateKey = priv
	return outCert, nil
}
