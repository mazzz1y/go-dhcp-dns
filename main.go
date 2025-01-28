package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/miekg/dns"
)

var (
	listenAddr       string
	fallbackDNS      string
	networkInterface string
	maxRetries       int
	queryTimeout     time.Duration
)

func init() {
	flag.StringVar(&listenAddr, "listen", "127.0.0.1:53533", "address to listen on")
	flag.StringVar(&fallbackDNS, "fallback", "8.8.8.8", "fallback DNS server")
	flag.StringVar(&networkInterface, "interface", "en0", "network interface for outgoing requests")
	flag.IntVar(&maxRetries, "retries", 3, "maximum number of query attempts")
	flag.DurationVar(&queryTimeout, "timeout", 5*time.Second, "query timeout duration")
}

func main() {
	flag.Parse()

	rand.NewSource(time.Now().UnixNano())

	dnsServer := newDNSServer()
	_ = dnsServer.updateServers()
	cache := newDNSCache()
	handler := newHandler(cache, dnsServer)

	server := &dns.Server{Addr: listenAddr, Net: "udp", Handler: dns.HandlerFunc(handler.handleDNS)}
	log.Printf("starting DNS forwarder on %s using interface %s", listenAddr, networkInterface)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
