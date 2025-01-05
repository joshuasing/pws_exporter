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

package dns

import (
	"context"
	"log/slog"
	"net"

	"github.com/miekg/dns"
)

// Server implements a simple proxying DNS server.
type Server struct {
	mux       *dns.ServeMux
	dnsServer *dns.Server

	records        map[string]string
	forwardDomains map[string]struct{}

	upstreamResolver string
	dnsClient        *dns.Client
}

// Config is the DNS server configuration.
type Config struct {
	// UpstreamResolver is the upstream DNS resolver to forward queries for
	// domains in the ForwardDomains list.
	UpstreamResolver string

	// Records is a list of A records to answer locally. Queries for names that
	// are not in this list or ForwardDomains will receive an answer of
	// NXDOMAIN.
	Records map[string]string

	// ForwardDomains is a list of domains for which to forward queries to the
	// UpstreamResolver. Domains that are not in this list or Records will
	// receive an answer of NXDOMAIN.
	ForwardDomains []string
}

// NewServer returns a new DNS server.
func NewServer(c Config) *Server {
	s := &Server{
		mux:              dns.NewServeMux(),
		records:          c.Records,
		forwardDomains:   make(map[string]struct{}),
		upstreamResolver: c.UpstreamResolver,
		dnsClient:        &dns.Client{},
	}
	for _, domain := range c.ForwardDomains {
		s.forwardDomains[domain] = struct{}{}
	}
	s.mux.Handle(".", s)
	return s
}

func (s *Server) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) != 1 {
		return
	}
	q := r.Question[0]
	domain := q.Name

	l := slog.With(slog.String("name", domain),
		slog.String("type", dns.TypeToString[q.Qtype]))
	l.Debug("Handling DNS query")

	// TODO: Probably not needed, but may need to eventually support AAAA?
	if q.Qtype == dns.TypeA {
		if ip, ok := s.records[domain]; ok {
			m := new(dns.Msg)
			m.SetReply(r)
			m.Answer = append(m.Answer, &dns.A{
				Hdr: dns.RR_Header{
					Name:   domain,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    3600,
				},
				A: net.ParseIP(ip),
			})
			l.Debug("Answering with local record",
				slog.String("a", ip))
			_ = w.WriteMsg(m)
			return
		}
	}

	// Forward queries for allowed/forwarded domains to the upstream resolver.
	if _, ok := s.forwardDomains[domain]; ok {
		res, _, err := s.dnsClient.Exchange(r, s.upstreamResolver)
		if err != nil {
			l.Error("Error forwarding DNS query",
				slog.Any("err", err))
			return
		}
		l.Debug("Resolved forwarded query",
			slog.Any("answers", res.Answer))
		_ = w.WriteMsg(res)
		return
	}

	// Blackhole other queries (NXDOMAIN)
	m := new(dns.Msg)
	m.SetRcode(r, dns.RcodeNameError)
	l.Debug("Answering with NXDOMAIN")
	_ = w.WriteMsg(m)
}

// ListenAndServe starts the DNS server on the given address.
func (s *Server) ListenAndServe(addr string) error {
	s.dnsServer = &dns.Server{Addr: addr, Net: "udp", Handler: s}
	return s.dnsServer.ListenAndServe()
}

// Shutdown shuts down the DNS server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.dnsServer.ShutdownContext(ctx)
}
