package ssh

import (
	"regexp"
	"strconv"
	"strings"
)

// PortInfo represents a listening port on a remote server.
type PortInfo struct {
	Port      int    // Port number
	Protocol  string // "tcp" or "udp"
	LocalAddr string // Binding address (e.g., "127.0.0.1", "0.0.0.0")
	Process   string // Process name (e.g., "python3", "nginx")
}

// ssLineRe matches "ss -tlnp" output lines.
// Example: LISTEN  0  128  127.0.0.1:13000  0.0.0.0:*  users:(("python3",pid=12345,fd=3))
var ssLineRe = regexp.MustCompile(`\S+\s+\d+\s+\d+\s+(\S+?):(\d+)\s+\S+.*users:\(\("([^"]+)`)

// netstatLineRe matches "netstat -tlnp" output lines.
// Example: tcp  0  0  127.0.0.1:13000  0.0.0.0:*  LISTEN  12345/python3
var netstatLineRe = regexp.MustCompile(`(tcp|udp)\s+\d+\s+\d+\s+(\S+?):(\d+)\s+\S+\s+\w+\s+(\d+)/(\S+)`)

// lsofLineRe matches "lsof -iTCP -sTCP:LISTEN -P -n" output lines.
// Only matches (LISTEN) connections, not (ESTABLISHED), etc.
// Example: node    12345 user   12u  IPv4 0x...      0t0  TCP 127.0.0.1:13000 (LISTEN)
var lsofLineRe = regexp.MustCompile(`(\S+)\s+(\d+)\s+\S+\s+\S+\s+\S+\s+\S+\s+\S+\s+TCP\s+(\S+?):(\d+)\s+\(LISTEN\)`)

// ParseSSOutput parses the output of "ss -tlnp", "netstat -tlnp", or "lsof -i -P -n".
func ParseSSOutput(output string) []PortInfo {
	var ports []PortInfo
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try ss format first
		if matches := ssLineRe.FindStringSubmatch(line); matches != nil {
			port, _ := strconv.Atoi(matches[2])
			ports = append(ports, PortInfo{
				Port:      port,
				Protocol:  "tcp",
				LocalAddr: matches[1],
				Process:   matches[3],
			})
			continue
		}

		// Try netstat format
		if matches := netstatLineRe.FindStringSubmatch(line); matches != nil {
			port, _ := strconv.Atoi(matches[3])
			ports = append(ports, PortInfo{
				Port:      port,
				Protocol:  strings.ToLower(matches[1]),
				LocalAddr: matches[2],
				Process:   matches[5],
			})
			continue
		}

		// Try lsof format (macOS)
		// Example: node    12345 user   12u  IPv4 0x...      0t0  TCP 127.0.0.1:13000 (LISTEN)
		if matches := lsofLineRe.FindStringSubmatch(line); matches != nil {
			port, _ := strconv.Atoi(matches[4])
			ports = append(ports, PortInfo{
				Port:      port,
				Protocol:  "tcp",
				LocalAddr: matches[3],
				Process:   matches[1],
			})
		}
	}

	return ports
}
