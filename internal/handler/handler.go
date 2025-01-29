package handler

import (
	"github.com/miekg/dns"
	"go-dhcp-dns/internal/cache"
	fw "go-dhcp-dns/internal/forwarder"
	"log"
)

type Handler struct {
	cache  *cache.DnsCache
	fw     *fw.Forwarder
	ifName string
}

func New(cache *cache.DnsCache, fw *fw.Forwarder) *Handler {
	return &Handler{
		cache: cache,
		fw:    fw,
	}
}

func (h *Handler) HandleDNS(w dns.ResponseWriter, r *dns.Msg) {
	if len(r.Question) == 0 {
		sendError(w, r, dns.RcodeFormatError)
		return
	}

	cacheKey := r.Question[0].String()
	if cached, exists := h.cache.Get(cacheKey); exists {
		cached.Id = r.Id
		_ = w.WriteMsg(cached)
		return
	}

	resp, err := h.fw.Query(r)
	if err != nil {
		log.Printf("forward dns error: %v", err)
		sendError(w, r, dns.RcodeServerFailure)
		return
	}

	h.cache.Set(cacheKey, resp)
	_ = w.WriteMsg(resp)
}

func sendError(w dns.ResponseWriter, r *dns.Msg, code int) {
	m := new(dns.Msg)
	m.SetRcode(r, code)
	_ = w.WriteMsg(m)
}
