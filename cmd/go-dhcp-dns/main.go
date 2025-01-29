package main

import (
	"flag"
	"go-dhcp-dns/internal/cache"
	"go-dhcp-dns/internal/handler"
	"log"
	"math/rand"
	"time"

	"github.com/miekg/dns"
	fw "go-dhcp-dns/internal/forwarder"
)

var (
	listenAddr   string
	netIf        string
	maxRetries   int
	queryTimeout time.Duration
)

func init() {
	flag.StringVar(&listenAddr, "listen", "127.0.0.1:53533", "address to listen on")
	flag.StringVar(&netIf, "interface", "en0", "network interface for outgoing requests")
	flag.IntVar(&maxRetries, "retries", 3, "maximum number of query attempts")
	flag.DurationVar(&queryTimeout, "timeout", 5*time.Second, "query timeout duration")
}

func main() {
	flag.Parse()

	rand.NewSource(time.Now().UnixNano())

	dnsServer := fw.New(maxRetries, queryTimeout, netIf)
	_ = dnsServer.UpdateServers()
	c := cache.NewDNSCache()
	h := handler.New(c, dnsServer)

	server := &dns.Server{
		Addr:    listenAddr,
		Net:     "udp",
		Handler: dns.HandlerFunc(h.HandleDNS),
	}

	log.Printf("starting DNS forwarder on %s using interface %s", listenAddr, netIf)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
