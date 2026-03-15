package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/edereagzi/portui/internal/process"
	"github.com/edereagzi/portui/internal/scanner"
	"github.com/edereagzi/portui/internal/tui"
	"github.com/edereagzi/portui/internal/update"
)

var version = "dev"

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: portui [flags] [command]\n\n")
	fmt.Fprintf(flag.CommandLine.Output(), "Commands:\n")
	fmt.Fprintf(flag.CommandLine.Output(), "  update    Update portui to the latest release\n\n")
	fmt.Fprintf(flag.CommandLine.Output(), "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.BoolVar(showVersion, "v", false, "print version and exit (shorthand)")
	flag.Parse()

	if *showVersion {
		fmt.Printf("portui %s\n", version)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) > 0 && args[0] == "update" {
		latest, updated, err := update.Run(version)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: update failed: %v\n", err)
			os.Exit(1)
		}
		if updated {
			if runtime.GOOS == "windows" {
				fmt.Printf("update to %s has been scheduled; restart portui to finish replacement\n", latest)
			} else {
				fmt.Printf("updated to %s\n", latest)
			}
		} else {
			fmt.Printf("already up to date (%s)\n", latest)
		}
		os.Exit(0)
	}

	s := scanner.NewScanner()
	ps := process.NewProcessService()
	if err := tui.Run(s, ps); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
