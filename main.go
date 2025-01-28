package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/miekg/dns"
)

const (
	listenAddr  = ":53533"
	networkType = "udp"
	fallbackDNS = "8.8.8.8"

	maxRetries           = 3
	queryTimeout         = 5 * time.Second
	cacheCleanupInterval = 1 * time.Hour

	networkInterface = "en0"
)

func main() {
	rand.NewSource(time.Now().UnixNano())

	dnsServer := newDNSServer()
	_ = dnsServer.updateServers()
	cache := newDNSCache()
	handler := newHandler(cache, dnsServer)

	dns.HandleFunc(".", handler.handleDNS)

	server := &dns.Server{
		Addr: listenAddr,
		Net:  networkType,
	}

	log.Printf("starting DNS forwarder on %s using interface %s", listenAddr, networkInterface)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
