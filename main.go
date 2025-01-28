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

	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		dnsHandler(w, r, cache, dnsServer)
	})

	server := &dns.Server{
		Addr: listenAddr,
		Net:  networkType,
	}

	log.Printf("starting DNS forwarder on %s", listenAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func dnsHandler(w dns.ResponseWriter, r *dns.Msg, cache *dnsCache, dnsServer *dnsServer) {
	if len(r.Question) == 0 {
		m := new(dns.Msg)
		m.SetRcode(r, dns.RcodeFormatError)
		w.WriteMsg(m)
		return
	}

	cacheKey := r.Question[0].String()
	if cached, exists := cache.get(cacheKey); exists {
		cached.Id = r.Id
		w.WriteMsg(cached)
		return
	}

	for i := 0; i < maxRetries; i++ {
		err := dnsServer.updateServers()
		if err != nil {
			log.Printf("error updating servers: %v", err)

		}
		upstream := dnsServer.getRandomServer()

		c := new(dns.Client)
		c.Timeout = queryTimeout
		resp, _, err := c.Exchange(r, upstream+":53")

		if err != nil {
			log.Printf("failed to exchange with %s: %v", upstream, err)
			continue
		}

		if resp != nil {
			cache.set(cacheKey, resp)
			w.WriteMsg(resp)
			return
		}
		time.Sleep(time.Duration(500) * time.Millisecond)
	}

	m := new(dns.Msg)
	m.SetRcode(r, dns.RcodeServerFailure)
	w.WriteMsg(m)
}
