package command

import (
	"fmt"
	"os"
	"os/exec"
)

// CheckSSH verifies that ssh is available on PATH.
func CheckSSH() {
	if _, err := exec.LookPath("ssh"); err != nil {
		fmt.Fprintf(os.Stderr, "error: ssh not found in PATH\n")
		fmt.Fprintf(os.Stderr, "please install OpenSSH or add it to your PATH\n")
		os.Exit(1)
	}
}
