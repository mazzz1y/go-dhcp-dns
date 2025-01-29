package forwarder

import (
	"fmt"
	"github.com/miekg/dns"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

type Forwarder struct {
	servers      []string
	mu           sync.RWMutex
	maxRetries   int
	queryTimeout time.Duration
	ifaceName    string
}

func New(maxRetries int, queryTimeout time.Duration, ifaceName string) *Forwarder {
	return &Forwarder{maxRetries: maxRetries, queryTimeout: queryTimeout, ifaceName: ifaceName}
}

func (d *Forwarder) Query(r *dns.Msg) (*dns.Msg, error) {
	for i := 0; i < d.maxRetries; i++ {
		if err := d.UpdateServers(); err != nil {
			return nil, fmt.Errorf("error getting dns servers: %v", err)
		}

		upstream := d.getRandomServer()
		localAddr, err := d.getInterfaceIP()
		if err != nil {
			return nil, fmt.Errorf("failing to get interface ip: %v", err)
		}

		c := &dns.Client{
			Timeout: d.queryTimeout,
			Dialer: &net.Dialer{
				LocalAddr: &net.UDPAddr{IP: localAddr},
				Timeout:   d.queryTimeout,
			},
		}

		resp, _, err := c.Exchange(r, fmt.Sprintf("%s:53", upstream))
		if err == nil && resp != nil {
			return resp, nil
		}

		log.Printf("attempt %d failed: %v", i+1, err)
		time.Sleep(500 * time.Millisecond)
	}

	return nil, fmt.Errorf("dns query failed after %d attempts", d.maxRetries)
}

func (d *Forwarder) UpdateServers() error {
	servers, err := d.getDNSServers()
	if err != nil {
		return fmt.Errorf("failed to update dns servers: %v", err)
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	if len(servers) > 0 {
		d.servers = servers
	}

	return nil
}

func (d *Forwarder) getRandomServer() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.servers[rand.Intn(len(d.servers))]
}

func (d *Forwarder) getInterfaceIP() (net.IP, error) {
	iface, err := net.InterfaceByName(d.ifaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get interface %s: %v", d.ifaceName, err)
	}

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

	return nil, fmt.Errorf("no IPv4 address found for interface %s", d.ifaceName)
}
