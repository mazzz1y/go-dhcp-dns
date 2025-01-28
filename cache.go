package main

import (
	"github.com/miekg/dns"
	"sync"
	"time"
)

type dnsCache struct {
	cache map[string]*dns.Msg
	times map[string]time.Time
	mu    sync.RWMutex
}

func newDNSCache() *dnsCache {
	cache := &dnsCache{
		cache: make(map[string]*dns.Msg),
		times: make(map[string]time.Time),
	}
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		for range ticker.C {
			cache.cleanup()
		}
	}()

	return cache
}

func (c *dnsCache) get(key string) (*dns.Msg, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if msg, exists := c.cache[key]; exists {
		if time.Since(c.times[key]) < cacheTime {
			return msg.Copy(), true
		}
		delete(c.cache, key)
		delete(c.times, key)
	}
	return nil, false
}

func (c *dnsCache) set(key string, msg *dns.Msg) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = msg.Copy()
	c.times[key] = time.Now()
}

func (c *dnsCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, timestamp := range c.times {
		if now.Sub(timestamp) >= cacheTime {
			delete(c.cache, key)
			delete(c.times, key)
		}
	}
}
