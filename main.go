// Copyright 2025 David Tanzer (business@davidtanzer.net)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"runtime"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	args := os.Args[1:]

	// Parse global flags before the subcommand
	workspaceConfig := ""
	for len(args) > 0 && args[0] == "--workspace-config" {
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "error: --workspace-config requires a file path\n")
			os.Exit(1)
		}
		workspaceConfig = args[1]
		args = args[2:]
	}

	if len(args) == 0 {
		printHelp()
		os.Exit(0)
	}

	cmd := args[0]
	cmdArgs := args[1:]

	var err error
	switch cmd {
	case "help", "--help", "-h":
		printHelp()
	case "info":
		printInfo()
	case "new-workspace":
		err = cmdNewWorkspace(workspaceConfig)
	case "register":
		err = cmdRegister(workspaceConfig, cmdArgs)
	case "validate":
		err = cmdValidate(workspaceConfig)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		printHelp()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(`Usage: agnt [--workspace-config <file>] <command>

Commands:
  new-workspace    Create a new .agnt.yaml workspace in the current directory
  register         Register the current tmux pane as a named agent
  validate         Check the workspace config and verify panes exist
  info             Print version and build information
  help             Show this help message

`)
}

func printInfo() {
	exe, err := os.Executable()
	if err != nil {
		exe = "(unknown)"
	}
	fmt.Printf("Version:    %s\n", version)
	fmt.Printf("Commit:     %s\n", commit)
	fmt.Printf("Built:      %s\n", date)
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("Binary:     %s\n", exe)
}
