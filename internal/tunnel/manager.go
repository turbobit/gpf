package tunnel

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/user/port-forwarding/internal/config"
	"github.com/user/port-forwarding/internal/ssh"
)

// Create creates a new port forwarding tunnel.
func Create(server config.SSHConfig, remotePort, localPort string) {
	// Auto-assign local port if not specified
	if localPort == "" {
		localPort = findNextPort()
	}

	// Check if local port is already in use
	if !isPortAvailable(localPort) {
		fmt.Fprintf(os.Stderr, "Error: local port %s is already in use\n", localPort)
		fmt.Fprintf(os.Stderr, "Try a different local port or stop the existing tunnel\n")
		os.Exit(1)
	}

	args := ssh.GenerateTunnelArgs(server, localPort, remotePort)
	fmt.Printf("Creating tunnel: %s -> localhost:%s (server: %s)\n", localPort, remotePort, server.Name)

	// Start the SSH tunnel in detached mode
	cmd := exec.Command("ssh", args...)
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
	pids := cmd.Process.Pid
	tunnels, _ := LoadState()
	tunnels = append(tunnels, Tunnel{
		ID:         fmt.Sprintf("%d", pids),
		ServerName: server.Name,
		LocalPort:  parseInt(localPort),
		RemotePort: parseInt(remotePort),
		PID:        pids,
		CreatedAt:  now(),
	})
	SaveState(tunnels)

	fmt.Printf("Tunnel created: localhost:%s -> %s:%s (PID: %d)\n", localPort, server.Name, remotePort, pids)
}

// Stop stops a specific tunnel by PID.
func Stop(id string) {
	pid, err := strconv.Atoi(id)
	if err != nil {
		// Try to find tunnel by ID string
		tunnels, _ := LoadState()
		found := false
		for _, t := range tunnels {
			if t.ID == id {
				pid = t.PID
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintf(os.Stderr, "Tunnel not found: %s\n", id)
			os.Exit(1)
		}
	}

	if err := killProcess(pid); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stop tunnel (PID %d): %v\n", pid, err)
		os.Exit(1)
	}

	// Remove from state
	tunnels, _ := LoadState()
	var remaining []Tunnel
	for _, t := range tunnels {
		if t.PID != pid {
			remaining = append(remaining, t)
		}
	}
	SaveState(remaining)

	fmt.Printf("Tunnel stopped (PID: %d)\n", pid)
}

// StopAll stops all active tunnels.
func StopAll() {
	tunnels, _ := LoadState()
	if len(tunnels) == 0 {
		fmt.Println("No active tunnels.")
		return
	}

	count := 0
	for _, t := range tunnels {
		if err := killProcess(t.PID); err == nil {
			count++
		}
	}
	SaveState(nil)
	fmt.Printf("Stopped %d tunnel(s).\n", count)
}

// KillProcess sends SIGTERM then SIGKILL to a process.
func killProcess(pid int) error {
	cmd := exec.Command("kill", "-TERM", strconv.Itoa(pid))
	if err := cmd.Run(); err != nil {
		// Try SIGKILL
		cmd = exec.Command("kill", "-KILL", strconv.Itoa(pid))
		return cmd.Run()
	}
	return nil
}

// findNextPort finds the next available port starting from 13000.
func findNextPort() string {
	for port := 13000; port < 65535; port++ {
		if isPortAvailable(strconv.Itoa(port)) {
			return strconv.Itoa(port)
		}
	}
	fmt.Fprintf(os.Stderr, "No available ports starting from 13000\n")
	os.Exit(1)
	return ""
}

// isPortAvailable checks if a TCP port is available.
func isPortAvailable(portStr string) bool {
	port := parseInt(portStr)

	// Check if gpf already has a tunnel on this port
	tunnels, _ := LoadState()
	for _, t := range tunnels {
		if t.LocalPort == port {
			return false
		}
	}

	// Check if OS is using this port
	conn, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%s", portStr))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func parseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func now() time.Time {
	return time.Now()
}
