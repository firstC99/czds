// Package main provides the czds command-line tool for interacting with ICANN's
// Centralized Zone Data Service (CZDS). The tool allows users to download zone files,
// request access to zones, and check the status of their requests.
package main

import (
	"fmt"
	"os"
)

// version contains the version string for the czds binary, set at build time.
var version = "dev"

// globalConfig holds the loaded configuration file settings
var globalConfig *Config

// main is the entry point for the czds CLI tool. It parses command-line arguments
// and dispatches to the appropriate subcommand handler.
func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]

	if subcommand == "help" || subcommand == "-h" || subcommand == "--help" {
		printUsage()
		os.Exit(0)
	}

	if subcommand == "version" || subcommand == "-version" || subcommand == "--version" {
		fmt.Printf("czds version %s\n", version)
		os.Exit(0)
	}

	if subcommand != "download" && subcommand != "dl" &&
		subcommand != "request" && subcommand != "req" &&
		subcommand != "status" && subcommand != "st" {
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}

	cfg, err := loadConfig("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
	}
	globalConfig = cfg

	ctx := getContext()

	switch subcommand {
	case "download", "dl":
		if err := downloadCmd().Run(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "request", "req":
		if err := requestCmd().Run(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "status", "st":
		if err := statusCmd().Run(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

// printUsage displays the help message for the czds command-line tool,
// including available subcommands, global options, and usage examples.
func printUsage() {
	fmt.Fprintf(os.Stderr, `czds - CZDS (Centralized Zone Data Service) client

Usage:
  czds <command> [options]

Available Commands:
  download, dl    Download zone files from CZDS
  request, req    Request access to zones, extensions, cancellations
  status, st      Check status of zone requests and generate reports
  version         Print version information
  help            Show this help message

Use "czds <command> -h" for more information about a command.

Global Options:
  -username string    Username to authenticate with (or set CZDS_USERNAME env var)
  -password string    Password to authenticate with (or set CZDS_PASSWORD env var)
  -verbose            Enable verbose logging

Examples:
  czds download -parallel 10 com org
  czds request -request-all -reason "Research project"
  czds status -zone com

`)
}
