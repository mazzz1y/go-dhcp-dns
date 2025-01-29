package cache

import (
	"github.com/miekg/dns"
	"sync"
	"time"
)

type DnsCache struct {
	mu    sync.RWMutex
	items map[string]*cacheEntry
}

type cacheEntry struct {
	msg       *dns.Msg
	expiresAt time.Time
}

func NewDNSCache() *DnsCache {
	cache := &DnsCache{
		items: make(map[string]*cacheEntry),
	}

	go cache.startCleanup(time.Duration(1) * time.Hour)
	return cache
}

func (c *DnsCache) Get(key string) (*dns.Msg, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.items[key]
	if !exists {
		return nil, false
	}

	now := time.Now()
	if now.After(entry.expiresAt) {
		go c.delete(key) // Cleanup expired entry
		return nil, false
	}

	// Calculate remaining TTL in seconds
	remainingTTL := uint32(entry.expiresAt.Sub(now).Seconds())
	if remainingTTL < 1 {
		go c.delete(key)
		return nil, false
	}

	resp := entry.msg.Copy()

	for _, rr := range resp.Answer {
		rr.Header().Ttl = remainingTTL
	}
	for _, rr := range resp.Ns {
		rr.Header().Ttl = remainingTTL
	}
	for _, rr := range resp.Extra {
		// Skip OPT records as they don't have a meaningful TTL
		if _, isOPT := rr.(*dns.OPT); !isOPT {
			rr.Header().Ttl = remainingTTL
		}
	}

	return resp, true
}

func (c *DnsCache) Set(key string, msg *dns.Msg) {
	if msg == nil {
		return
	}

	ttl := calculateTTL(msg)
	if ttl == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &cacheEntry{
		msg:       msg.Copy(),
		expiresAt: time.Now().Add(ttl),
	}
}

func (c *DnsCache) delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *DnsCache) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		c.mu.Lock()
		defer c.mu.Unlock()

		now := time.Now()
		for key, entry := range c.items {
			if now.After(entry.expiresAt) {
				delete(c.items, key)
			}
		}
	}
}

func calculateTTL(msg *dns.Msg) time.Duration {
	if msg.Rcode == dns.RcodeNameError {
		for _, rr := range msg.Ns {
			if soa, ok := rr.(*dns.SOA); ok {
				return time.Duration(soa.Minttl) * time.Second
			}
		}
	}

	sections := [][]dns.RR{msg.Answer, msg.Ns, msg.Extra}
	for _, section := range sections {
		for _, rr := range section {
			if ttl := rr.Header().Ttl; ttl > 0 {
				return time.Duration(ttl) * time.Second
			}
		}
	}

	return 0
}
