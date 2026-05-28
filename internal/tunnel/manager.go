package tunnel

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/turbobit/gpf/internal/config"
	"github.com/turbobit/gpf/internal/ssh"
)

// Create creates a new port forwarding tunnel.
func Create(server config.SSHConfig, remotePort, localPort string) {
	// Auto-assign local port if not specified
	if localPort == "" {
		localPort = strconv.Itoa(FindNextPort())
	}

	// Check if local port is already in use
	localPortNum, _ := strconv.Atoi(localPort)
	remotePortNum, _ := strconv.Atoi(remotePort)
	if !IsPortAvailable(localPortNum) {
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
		LocalPort:  localPortNum,
		RemotePort: remotePortNum,
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

	if err := KillProcess(pid); err != nil {
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
		if err := KillProcess(t.PID); err == nil {
			count++
		}
	}
	SaveState(nil)
	fmt.Printf("Stopped %d tunnel(s).\n", count)
}

// KillProcess sends SIGTERM then SIGKILL (Unix) or taskkill (Windows) to a process.
func KillProcess(pid int) error {
	if runtime.GOOS == "windows" {
		out, err := exec.Command("taskkill", "/F", "/PID", strconv.Itoa(pid)).CombinedOutput()
		fmt.Fprintf(os.Stderr, "[KILL] PID=%d exit=%v output=%s\n", pid, err, string(out))
		if err != nil {
			return fmt.Errorf("taskkill failed: %v, output: %s", err, string(out))
		}
		return nil
	}
	cmd := exec.Command("kill", "-TERM", strconv.Itoa(pid))
	if err := cmd.Run(); err != nil {
		// Try SIGKILL
		cmd = exec.Command("kill", "-KILL", strconv.Itoa(pid))
		return cmd.Run()
	}
	return nil
}

// KillProcessByPort finds and kills SSH processes listening on a given local port.
func KillProcessByPort(port int) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("only supported on Windows")
	}
	// Find all PIDs listening on this port
	out, err := exec.Command("netstat", "-ano").CombinedOutput()
	if err != nil {
		return fmt.Errorf("netstat failed: %v", err)
	}

	killCount := 0
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(line, fmt.Sprintf(":%d ", port)) && !strings.Contains(line, fmt.Sprintf(":%d\t", port)) {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		pidStr := fields[len(fields)-1]
		pid, err := strconv.Atoi(pidStr)
		if err != nil || pid == 0 {
			continue
		}
		// Only kill ssh.exe processes
		cmdOut, _ := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/FO", "CSV").CombinedOutput()
		if !strings.Contains(string(cmdOut), "ssh") && !strings.Contains(string(cmdOut), "SSH") {
			continue
		}
		kout, kerr := exec.Command("taskkill", "/F", "/PID", pidStr).CombinedOutput()
		fmt.Fprintf(os.Stderr, "[KILL-PORT] PID=%d port=%d exit=%v output=%s\n", pid, port, kerr, string(kout))
		if kerr == nil {
			killCount++
		}
	}
	if killCount == 0 {
		return fmt.Errorf("no SSH process found on port %d", port)
	}
	return nil
}

// FindNextPort finds the next available port starting from 13000.
func FindNextPort() int {
	for port := 13000; port < 65535; port++ {
		if IsPortAvailable(port) {
			return port
		}
	}
	fmt.Fprintf(os.Stderr, "No available ports starting from 13000\n")
	os.Exit(1)
	return 0
}

// IsPortAvailable checks if a TCP port is available.
func IsPortAvailable(port int) bool {
	// Check if gpf already has a tunnel on this port
	tunnels, _ := LoadState()
	for _, t := range tunnels {
		if t.LocalPort == port {
			return false
		}
	}

	// Check if OS is using this port
	conn, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func now() time.Time {
	return time.Now()
}
