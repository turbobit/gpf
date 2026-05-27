package interactive

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/user/port-forwarding/internal/config"
)

// Config launches the server list TUI.
func Config(keyword string) {
	m := initialModel(keyword, ModeConfig)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		// TUI ended, nothing to do
	}
}

// ShowTunnels launches the tunnel manager TUI.
func ShowTunnels() {
	m := initialModel("", ModeTunnels)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		// TUI ended with error
	}
}

// Mode represents the current TUI mode.
type Mode int

const (
	ModeConfig Mode = iota // Server list
	ModePorts              // Port list for selected server
	ModeTunnelCreate       // Tunnel creation in progress
	ModeSSH                // SSH connection
	ModeTunnels            // Active tunnels list
	ModeActionSelect       // Action selection (Forward vs SSH)
)

// ServerChoice is returned when a server is selected.
type ServerChoice struct {
	Server config.SSHConfig
	Action ActionChoice // ActionForward or ActionSSH
}

// ActionChoice represents the user's action selection.
type ActionChoice int

const (
	ActionForward ActionChoice = iota
	ActionSSH
)
