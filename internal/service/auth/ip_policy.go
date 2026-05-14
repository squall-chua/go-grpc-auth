package auth

import (
	"net"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ValidateIP(clientIP string, allowlist, denylist []string) error {
	if clientIP == "" {
		return nil
	}

	ip := net.ParseIP(clientIP)
	if ip == nil {
		return nil
	}

	// Check denylist first
	for _, entry := range denylist {
		if matchesIP(ip, entry) {
			return status.Error(codes.PermissionDenied, "IP address is blocked")
		}
	}

	// If allowlist is not empty, IP must be in it
	if len(allowlist) > 0 {
		allowed := false
		for _, entry := range allowlist {
			if matchesIP(ip, entry) {
				allowed = true
				break
			}
		}
		if !allowed {
			return status.Error(codes.PermissionDenied, "IP address not allowed")
		}
	}

	return nil
}

func matchesIP(ip net.IP, entry string) bool {
	// Check if entry is a CIDR
	_, ipnet, err := net.ParseCIDR(entry)
	if err == nil {
		return ipnet.Contains(ip)
	}

	// Otherwise exact match
	entryIP := net.ParseIP(entry)
	if entryIP != nil {
		return ip.Equal(entryIP)
	}

	return false
}
