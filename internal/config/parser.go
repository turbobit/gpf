package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// SSHConfig represents a parsed SSH host entry.
type SSHConfig struct {
	Name         string // Host alias
	Host         string // HostName directive
	Port         string // Port directive (default "22")
	User         string // User directive
	IdentityFile string // IdentityFile directive
	Comment      string // Comment or description
}

// GetConfig reads and parses ~/.ssh/config (or the given path).
func GetConfig() ([]SSHConfig, error) {
	return ParseConfig(DefaultConfigPath())
}

// GetConfigWithSearch parses config and filters by keyword.
func GetConfigWithSearch(keyword string) ([]SSHConfig, error) {
	servers, err := GetConfig()
	if err != nil {
		return nil, err
	}
	if keyword == "" {
		return servers, nil
	}
	keyword = strings.ToLower(keyword)
	var result []SSHConfig
	for _, s := range servers {
		if strings.Contains(strings.ToLower(s.Name), keyword) ||
			strings.Contains(strings.ToLower(s.Host), keyword) ||
			strings.Contains(strings.ToLower(s.User), keyword) {
			result = append(result, s)
		}
	}
	return result, nil
}

// ParseConfig reads and parses an SSH config file.
func ParseConfig(path string) ([]SSHConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	return parse(string(data))
}

// parse parses SSH config text into SSHConfig entries.
func parse(text string) ([]SSHConfig, error) {
	var configs []SSHConfig
	var current *SSHConfig
	scanner := bufio.NewScanner(strings.NewReader(text))

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check if this is a Host line (not indented)
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			lower := strings.ToLower(trimmed)
			if strings.HasPrefix(lower, "host ") || trimmed == "host" {
				// Save previous entry
				if current != nil {
					finalize(current)
					configs = append(configs, *current)
				}
				// Start new entry
				parts := strings.Fields(trimmed)
				if len(parts) < 2 {
					current = &SSHConfig{}
				} else {
					// Support "Host name1 name2 name3" — create one entry per name
					for _, name := range parts[1:] {
						current = &SSHConfig{Name: name}
						configs = append(configs, *current)
					}
					current = &SSHConfig{Name: parts[1]}
				}
				continue
			}
			// Unknown top-level directive, skip
			continue
		}

		// Indented line — key-value pair
		if current == nil {
			continue
		}
		parts := strings.Fields(trimmed)
		if len(parts) < 2 {
			continue
		}
		key := strings.ToLower(parts[0])
		value := strings.Join(parts[1:], " ")

		// Strip quotes
		value = unquote(value)

		switch key {
		case "hostname":
			current.Host = value
		case "port":
			current.Port = value
		case "user":
			current.User = value
		case "identityfile":
			current.IdentityFile = value
		case "comment":
			current.Comment = value
		}
	}

	// Finalize last entry
	if current != nil {
		finalize(current)
		configs = append(configs, *current)
	}

	return configs, nil
}

// finalize applies defaults to a parsed entry.
func finalize(s *SSHConfig) {
	if s.Host == "" {
		s.Host = s.Name
	}
	if s.Port == "" {
		s.Port = "22"
	}
	if s.User == "" {
		s.User = os.Getenv("USER")
		if s.User == "" {
			s.User = os.Getenv("USERNAME")
		}
	}
}

// unquote strips surrounding quotes from a string.
func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') ||
			(s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
