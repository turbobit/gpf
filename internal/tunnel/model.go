package tunnel

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/turbobit/gpf/internal/config"
)

// Tunnel represents an active port forwarding tunnel.
type Tunnel struct {
	ID         string    `json:"id"`
	ServerName string    `json:"server_name"`
	LocalPort  int       `json:"local_port"`
	RemotePort int       `json:"remote_port"`
	PID        int       `json:"pid"`
	CreatedAt  time.Time `json:"created_at"`
}

// StateFile returns the path to the tunnels state file.
func StateFile() string {
	dir := filepath.Join(os.Getenv("HOME"), ".gpf")
	os.MkdirAll(dir, 0700)
	return filepath.Join(dir, "tunnels.json")
}

// LoadState reads the tunnels state file.
func LoadState() ([]Tunnel, error) {
	data, err := os.ReadFile(StateFile())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var state struct {
		Tunnels []Tunnel `json:"tunnels"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return state.Tunnels, nil
}

// SaveState writes the tunnels state file.
func SaveState(tunnels []Tunnel) error {
	dir := filepath.Join(os.Getenv("HOME"), ".gpf")
	os.MkdirAll(dir, 0700)
	data, err := json.MarshalIndent(map[string][]Tunnel{"tunnels": tunnels}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(StateFile(), data, 0600)
}

// ToServer converts a Tunnel back to its SSH config.
func (t *Tunnel) ToServer() config.SSHConfig {
	servers, _ := config.GetConfig()
	for _, s := range servers {
		if s.Name == t.ServerName {
			return s
		}
	}
	return config.SSHConfig{}
}
