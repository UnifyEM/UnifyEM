/******************************************************************************
 * Copyright (c) 2024-2026 Tenebris Technologies Inc.                         *
 * Please see the LICENSE file for details                                    *
 ******************************************************************************/

package status

import (
	"net"
	"os"
	"strings"
)

//
// Cross-platform data retrieval functions
//

// hostname returns the hostname of the system
func (h *Handler) hostname() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return strings.ToLower(name)
}

// ip returns a comma-separated list of IP addresses for the system
// Loopback, local link, ULA IPv6 addresses, and down interfaces are excluded
func (h *Handler) ip() string {
	var ips []string
	interfaces, err := net.Interfaces()
	if err != nil {
		return "unknown"
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			// Exclude link-local IPv6 addresses
			if ip.IsLinkLocalUnicast() {
				continue
			}

			// Exclude unique local IPv6 addresses (ULA)
			if ip.IsPrivate() && ip.To16() != nil && ip.To4() == nil && ip[0] == 0xfd {
				continue
			}

			ips = append(ips, ip.String())
		}
	}

	if len(ips) == 0 {
		return "unknown"
	}

	return strings.Join(ips, ",")
}
