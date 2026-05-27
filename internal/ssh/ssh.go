package ssh

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/user/port-forwarding/internal/config"
)

// GenerateTunnelArgs builds the ssh command args for port forwarding.
func GenerateTunnelArgs(server config.SSHConfig, localPort, remotePort string) []string {
	args := []string{
		"-N",
		"-f",
		"-L", fmt.Sprintf("%s:localhost:%s", localPort, remotePort),
		"-o", "ExitOnForwardFailure=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=10",
		"-p", server.Port,
		server.User + "@" + server.Host,
	}
	if server.IdentityFile != "" {
		args = append(args, "-i", server.IdentityFile)
	}
	return args
}

// Run executes an ssh command with the given args.
func Run(args []string) error {
	cmd := exec.Command("ssh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// RunDetached executes an ssh command in the background.
func RunDetached(args []string) error {
	cmd := exec.Command("ssh", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Start()
}

// ShowPorts connects to the server and displays listening ports.
func ShowPorts(alias string) {
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

	output, err := runRemoteCommand(server, "ss -tlnp 2>/dev/null || netstat -tlnp 2>/dev/null")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning ports: %v\n", err)
		os.Exit(1)
	}

	ports := ParseSSOutput(output)
	if len(ports) == 0 {
		fmt.Println("No listening ports found.")
		return
	}

	fmt.Printf("%-8s %-8s %-12s %s\n", "PORT", "PROTO", "LOCAL_ADDR", "PROCESS")
	for _, p := range ports {
		fmt.Printf("%-8d %-8s %-12s %s\n", p.Port, p.Protocol, p.LocalAddr, p.Process)
	}
}

// runRemoteCommand executes a command on a remote server via SSH and returns stdout.
func runRemoteCommand(server config.SSHConfig, command string) (string, error) {
	args := []string{
		"-p", server.Port,
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "ConnectTimeout=10",
		"-o", "BatchMode=yes",
		server.User + "@" + server.Host,
		"-" + command,
	}
	if server.IdentityFile != "" {
		args = append([]string{"-i", server.IdentityFile}, args...)
	}

	cmd := exec.Command("ssh", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ssh command failed: %w: %s", err, string(output))
	}
	return string(output), nil
}
