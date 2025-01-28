package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

type dnsServer struct {
	servers []string
	mu      sync.RWMutex
}

func newDNSServer() *dnsServer {
	return &dnsServer{
		servers: []string{fallbackDNS},
	}
}

func (s *dnsServer) updateServers() error {
	servers, err := getDNSServers()
	if err != nil {
		return fmt.Errorf("failed to update DNS servers: %v, using fallback: %s", err, fallbackDNS)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if len(servers) > 0 {
		s.servers = servers
	}

	return nil
}

func (s *dnsServer) getRandomServer() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.servers) == 0 {
		return fallbackDNS
	}
	return s.servers[rand.Intn(len(s.servers))]
}

func isValidIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && !parsedIP.IsUnspecified() &&
		!parsedIP.IsLoopback() && !parsedIP.IsMulticast()
}

func getDNSServers() ([]string, error) {
	if runtime.GOOS != "darwin" {
		return []string{}, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	cmd := exec.Command("ipconfig", "getpacket", networkInterface)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute ipconfig: %v", err)
	}

	return parseDNSServers(output)
}

func parseDNSServers(output []byte) ([]string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "domain_name_server") {
			start := strings.Index(line, "{")
			end := strings.Index(line, "}")
			if start == -1 || end == -1 {
				return nil, fmt.Errorf("invalid DNS server format")
			}

			ips := strings.Split(line[start+1:end], ", ")
			validIPs := make([]string, 0)

			for _, ip := range ips {
				ip = strings.TrimSpace(ip)
				if isValidIP(ip) {
					validIPs = append(validIPs, ip)
				} else {
					log.Printf("invalid IP address found: %s", ip)
				}
			}

			if len(validIPs) > 0 {
				return validIPs, nil
			}
			return nil, fmt.Errorf("no valid DNS servers found")
		}
	}

	return nil, fmt.Errorf("DNS server section not found")
}
