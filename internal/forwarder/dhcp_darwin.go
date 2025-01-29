package forwarder

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func (d *Forwarder) getDNSServers() ([]string, error) {
	output, err := exec.Command("ipconfig", "getpacket", d.ifaceName).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute ipconfig: %v", err)
	}

	return parse(output)
}

func parse(output []byte) ([]string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "domain_name_server") {
			continue
		}

		start := strings.Index(line, "{")
		end := strings.Index(line, "}")
		if start == -1 || end == -1 {
			return nil, fmt.Errorf("invalid dns server format")
		}

		validIPs := make([]string, 0)
		for _, ip := range strings.Split(line[start+1:end], ",") {
			ip = strings.TrimSpace(ip)
			if isValidIP(ip) {
				validIPs = append(validIPs, ip)
			}
		}

		if len(validIPs) > 0 {
			return validIPs, nil
		}
		return nil, fmt.Errorf("no valid dns servers found")
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning output: %v", err)
	}

	return nil, fmt.Errorf("dns server section not found")
}
