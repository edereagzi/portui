package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/edereagzi/portui/internal/process"
	"github.com/edereagzi/portui/internal/scanner"
	"github.com/edereagzi/portui/internal/tui"
)

var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.BoolVar(showVersion, "v", false, "print version and exit (shorthand)")
	flag.Parse()

	if *showVersion {
		fmt.Printf("portui %s\n", version)
		os.Exit(0)
	}

	s := scanner.NewScanner()
	ps := process.NewProcessService()
	if err := tui.Run(s, ps); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
