package config

import (
	"os"
	"path/filepath"
)

// DefaultConfigPath returns the default SSH config file path.
func DefaultConfigPath() string {
	return filepath.Join(os.Getenv("HOME"), ".ssh", "config")
}

// HomeDir returns the user's home directory.
func HomeDir() string {
	return os.Getenv("HOME")
}
