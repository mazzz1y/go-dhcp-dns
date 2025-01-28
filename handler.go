package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/miekg/dns"
)

type Handler struct {
	cache     *dnsCache
	dnsServer *dnsServer
}

func newHandler(cache *dnsCache, dnsServer *dnsServer) *Handler {
	return &Handler{
		cache:     cache,
		dnsServer: dnsServer,
	}
}

func (h *Handler) handleDNS(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) == 0 {
		m := new(dns.Msg)
		m.SetRcode(r, dns.RcodeFormatError)
		w.WriteMsg(m)
		return
	}

	cacheKey := r.Question[0].String()
	if cached, exists := h.cache.get(cacheKey); exists {
		cached.Id = r.Id
		w.WriteMsg(cached)
		return
	}

	iface, err := net.InterfaceByName(networkInterface)
	if err != nil {
		log.Printf("failed to get interface %s: %v", networkInterface, err)
		m := new(dns.Msg)
		m.SetRcode(r, dns.RcodeServerFailure)
		w.WriteMsg(m)
		return
	}

	localAddr, err := h.getInterfaceIP(iface)
	if err != nil {
		log.Printf("failed to get IP address: %v", err)
		m := new(dns.Msg)
		m.SetRcode(r, dns.RcodeServerFailure)
		w.WriteMsg(m)
		return
	}

	h.forwardDNSQuery(w, r, localAddr)
}

func (h *Handler) getInterfaceIP(iface *net.Interface) (net.IP, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.To4() != nil {
				return ipnet.IP, nil
			}
		}
	}

	return nil, fmt.Errorf("no IPv4 address found for interface %s", networkInterface)
}

func (h *Handler) forwardDNSQuery(w dns.ResponseWriter, r *dns.Msg, localAddr net.IP) {
	for i := 0; i < maxRetries; i++ {
		err := h.dnsServer.updateServers()
		if err != nil {
			log.Printf("error updating servers: %v", err)
		}
		upstream := h.dnsServer.getRandomServer()

		c := &dns.Client{
			Timeout: queryTimeout,
			Dialer: &net.Dialer{
				LocalAddr: &net.UDPAddr{
					IP: localAddr,
				},
				Timeout: queryTimeout,
			},
		}

		resp, _, err := c.Exchange(r, upstream+":53")
		if err != nil {
			log.Printf("failed to exchange with %s: %v", upstream, err)
			continue
		}

		if resp != nil {
			h.cache.set(r.Question[0].String(), resp)
			w.WriteMsg(resp)
			return
		}
		time.Sleep(time.Duration(500) * time.Millisecond)
	}

	m := new(dns.Msg)
	m.SetRcode(r, dns.RcodeServerFailure)
	w.WriteMsg(m)
}
