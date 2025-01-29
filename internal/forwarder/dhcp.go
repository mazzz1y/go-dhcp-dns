package forwarder

import (
	"net"
)

func isValidIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	return !parsedIP.IsUnspecified() &&
		!parsedIP.IsLoopback() &&
		!parsedIP.IsMulticast() &&
		!parsedIP.IsLinkLocalUnicast()
}
