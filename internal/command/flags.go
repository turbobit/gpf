package command

import "strings"

// Action represents a gpf command mode.
type Action int

const (
	InteractiveConfig Action = iota
	ShowPorts
	CreateTunnel
	ShowTunnels
	StopTunnel
	StopAllTunnels
)

// Which parses CLI args and returns the action and its value.
func Which(args []string) (Action, string) {
	if len(args) == 0 {
		return InteractiveConfig, ""
	}

	// "gpf -" or "gpf - keyword"
	if args[0] == "-" {
		keyword := strings.Join(args[1:], " ")
		return InteractiveConfig, keyword
	}

	// Subcommands
	switch args[0] {
	case "ports":
		return ShowPorts, strings.Join(args[1:], " ")
	case "forward":
		return CreateTunnel, strings.Join(args[1:], " ")
	case "tunnels":
		return ShowTunnels, ""
	case "stop":
		if len(args) < 2 {
			return InteractiveConfig, ""
		}
		return StopTunnel, args[1]
	case "stop-all":
		return StopAllTunnels, ""
	default:
		return InteractiveConfig, ""
	}
}
