package interactive

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/turbobit/gpf/internal/config"
	"github.com/turbobit/gpf/internal/i18n"
)

// Choice represents the user's selection from the TUI.
type Choice struct {
	Server     config.SSHConfig
	Action     ActionChoice
	LocalPort  int
	RemotePort int
}

// Config launches the server list TUI and returns the user's choice.
// If locale is non-empty, it overrides the auto-detected language.
func Config(keyword string, locale string) *Choice {
	if locale != "" {
		i18n.Force(locale)
	}

	m := initialModel(keyword, ModeConfig)
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\nRetrying without alternate screen...\n", err)
		p2 := tea.NewProgram(m)
		result2, err2 := p2.Run()
		if err2 != nil {
			fmt.Fprintf(os.Stderr, "TUI error: %v\n", err2)
			return nil
		}
		return extractChoice(result2)
	}
	// Handle both value and pointer types returned by Bubble Tea
	var r *model
	switch v := result.(type) {
	case *model:
		r = v
	case model:
		r = &v
	default:
	}
	if r != nil {
		if r.sshTarget.Name != "" {
			return &Choice{
				Server: r.sshTarget,
				Action: ActionSSH,
			}
		}
		if r.forwardTarget != nil {
			return &Choice{
				Server:     r.forwardTarget.Server,
				Action:     ActionForward,
				LocalPort:  r.forwardTarget.LocalPort,
				RemotePort: r.forwardTarget.RemotePort,
			}
		}
	}

	return nil
}

func extractChoice(result tea.Model) *Choice {
	var m *model
	switch v := result.(type) {
	case *model:
		m = v
	case model:
		m = &v
	default:
		return nil
	}
	if m.sshTarget.Name != "" {
		return &Choice{Server: m.sshTarget, Action: ActionSSH}
	}
	if m.forwardTarget != nil {
		return &Choice{
			Server:     m.forwardTarget.Server,
			Action:     ActionForward,
			LocalPort:  m.forwardTarget.LocalPort,
			RemotePort: m.forwardTarget.RemotePort,
		}
	}
	return nil
}

// ShowTunnels launches the tunnel manager TUI.
// If locale is non-empty, it overrides the auto-detected language.
func ShowTunnels(locale string) {
	if locale != "" {
		i18n.Force(locale)
	}
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
	ModeTunnels            // Active tunnels list
	ModeActionSelect       // Action selection (Forward vs SSH)
)

// ActionChoice represents the user's action selection.
type ActionChoice int

const (
	ActionForward ActionChoice = iota
	ActionSSH
)
