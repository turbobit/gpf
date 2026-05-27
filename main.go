package main

import (
	"fmt"
	"os"

	"github.com/turbobit/gpf/cmd"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "version" || os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("gpf version %s\n  built: %s\n  commit: %s\n", version, date, commit)
		os.Exit(0)
	}

	cmd.Main()
}
