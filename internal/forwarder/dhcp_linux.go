package forwarder

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

const resolvconf = "/run/NetworkManager/no-stub-resolv.conf"

func (d *Forwarder) getDNSServers() ([]string, error) {
	content, err := os.ReadFile(resolvconf)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %v", resolvconf, err)
	}

	return parse(content)
}

func parse(content []byte) ([]string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	validIPs := make([]string, 0)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// parse nameserver lines
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "nameserver" {
			ip := fields[1]
			if isValidIP(ip) {
				validIPs = append(validIPs, ip)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning resolv.conf: %v", err)
	}

	if len(validIPs) == 0 {
		return nil, fmt.Errorf("no valid dns servers found in resolv.conf")
	}

	return validIPs, nil
}
