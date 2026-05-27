package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/turbobit/gpf/internal/command"
	"github.com/turbobit/gpf/internal/config"
	"github.com/turbobit/gpf/internal/interactive"
	"github.com/turbobit/gpf/internal/ssh"
	"github.com/turbobit/gpf/internal/tunnel"
)

func Main() {
	command.CheckSSH()
	args := os.Args[1:]

	action, value := command.Which(args)
	switch action {
	case command.InteractiveConfig:
		interactive.Config(value)
	case command.ShowPorts:
		ssh.ShowPorts(value)
	case command.CreateTunnel:
		createTunnel(value)
	case command.ShowTunnels:
		interactive.ShowTunnels()
	case command.StopTunnel:
		tunnel.Stop(value)
	case command.StopAllTunnels:
		tunnel.StopAll()
	default:
		interactive.Config("")
	}
}

func createTunnel(raw string) {
	parts := strings.SplitN(raw, " ", 3)
	if len(parts) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: gpf forward <alias> <remote-port> [local-port]\n")
		os.Exit(1)
	}
	alias := parts[0]
	remotePort := parts[1]

	servers, err := config.GetConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading SSH config: %v\n", err)
		os.Exit(1)
	}

	var server config.SSHConfig
	for _, s := range servers {
		if s.Name == alias {
			server = s
			break
		}
	}
	if server.Name == "" {
		fmt.Fprintf(os.Stderr, "Server not found: %s\n", alias)
		os.Exit(1)
	}

	tunnel.Create(server, remotePort, parts[2])
}
