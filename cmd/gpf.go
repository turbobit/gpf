package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/turbobit/gpf/internal/command"
	"github.com/turbobit/gpf/internal/config"
	"github.com/turbobit/gpf/internal/interactive"
	"github.com/turbobit/gpf/internal/ssh"
	"github.com/turbobit/gpf/internal/tunnel"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func Main() {
	command.CheckSSH()
	args := os.Args[1:]

	opts := command.Which(args)
	switch opts.Action {
	case command.InteractiveConfig:
		runInteractive(opts.Value, opts.Locale)
	case command.ShowPorts:
		ssh.ShowPorts(opts.Value)
	case command.CreateTunnel:
		createTunnel(opts.Value)
	case command.ShowTunnels:
		interactive.ShowTunnels(opts.Locale)
	case command.StopTunnel:
		tunnel.Stop(opts.Value)
	case command.StopAllTunnels:
		tunnel.StopAll()
	case command.ShowVersion:
		showVersion()
	default:
		runInteractive("", opts.Locale)
	}
}

func runInteractive(keyword, locale string) {
	fmt.Println("\n gpf — Greenfield Port Forwarding")
	choice := interactive.Config(keyword, locale)
	if choice == nil {
		return
	}
	switch choice.Action {
	case interactive.ActionSSH:
		fmt.Printf("\nConnecting to %s...\n", choice.Server.Host)
		runSSH(choice.Server)
	case interactive.ActionForward:
		runForward(choice.Server, choice.LocalPort, choice.RemotePort)
	}
}

func runSSH(server config.SSHConfig) {
	args := []string{
		"-p", server.Port,
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=10",
		server.User + "@" + server.Host,
	}
	if server.IdentityFile != "" {
		args = append([]string{"-i", server.IdentityFile}, args...)
	}

	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "SSH connection failed: %v\n", err)
	}
}

func runForward(server config.SSHConfig, localPort, remotePort int) {
	// Auto-assign local port if in use
	actualLocalPort := localPort
	if !tunnel.IsPortAvailable(localPort) {
		actualLocalPort = tunnel.FindNextPort()
		fmt.Printf("Local port %d is in use, using %d instead\n", localPort, actualLocalPort)
	}

	args := []string{
		"-N",
		"-L", fmt.Sprintf(":%d:localhost:%d", actualLocalPort, remotePort),
		"-o", "ExitOnForwardFailure=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=10",
		"-o", "ServerAliveInterval=60",
		"-o", "ServerAliveCountMax=9999",
		"-o", "TCPKeepAlive=yes",
		"-p", server.Port,
		server.User + "@" + server.Host,
	}
	if server.IdentityFile != "" {
		args = append([]string{"-i", server.IdentityFile}, args...)
	}

	fmt.Printf("Creating tunnel: localhost:%d -> %s:%d (server: %s)\n", actualLocalPort, server.Name, remotePort, server.Name)

	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start tunnel: %v\n", err)
		os.Exit(1)
	}

	// Wait briefly to check if it started successfully
	go func() {
		_ = cmd.Wait()
	}()

	// Record the tunnel
	tunnels, _ := tunnel.LoadState()
	tunnels = append(tunnels, tunnel.Tunnel{
		ID:         fmt.Sprintf("%d", cmd.Process.Pid),
		ServerName: server.Name,
		LocalPort:  actualLocalPort,
		RemotePort: remotePort,
		PID:        cmd.Process.Pid,
	})
	tunnel.SaveState(tunnels)

	fmt.Printf("Tunnel created: localhost:%d -> %s:%d (PID: %d)\n", actualLocalPort, server.Name, remotePort, cmd.Process.Pid)
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

func showVersion() {
	fmt.Printf("gpf version %s\n  built: %s\n  commit: %s\n", version, date, commit)
}
