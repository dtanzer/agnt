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
	"os/exec"
	"sort"
	"strings"
)

func cmdValidate(explicit string) error {
	// Step 1: find workspace
	configPath, err := resolveWorkspacePath(explicit, false)
	if err != nil {
		return err
	}
	if configPath == "" {
		return fmt.Errorf("no workspace found (no .agnt.yaml in current directory or any parent up to $HOME)")
	}

	fmt.Printf("Workspace:  %s\n", configPath)

	// Step 2: syntax check
	cfg, err := loadConfig(configPath)
	if err != nil {
		fmt.Printf("Syntax:     FAILED (%s)\n", err)
		fmt.Printf("\nSummary: 0 checks passed, 1 failed\n")
		return fmt.Errorf("validation failed")
	}
	fmt.Printf("Syntax:     OK\n")

	// Step 3: per-agent pane checks
	names := make([]string, 0, len(cfg.Agents))
	for name := range cfg.Agents {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Printf("Agents (%d):\n", len(names))

	failed := 0
	for _, name := range names {
		agent := cfg.Agents[name]
		exists := paneExists(agent.Pane)
		if exists {
			cmd := paneCommand(agent.Pane)
			if cmd != "" {
				fmt.Printf("  %-12s pane %-4s OK  [%s]\n", name, agent.Pane, cmd)
			} else {
				fmt.Printf("  %-12s pane %-4s OK\n", name, agent.Pane)
			}
		} else {
			fmt.Printf("  %-12s pane %-4s MISSING\n", name, agent.Pane)
			failed++
		}
	}

	passed := len(names) - failed
	total := passed + failed
	fmt.Printf("\nSummary: %d/%d checks passed", passed, total)
	if failed > 0 {
		fmt.Printf(", %d failed", failed)
	}
	fmt.Println()

	if failed > 0 {
		return fmt.Errorf("validation failed")
	}
	return nil
}

// paneExists returns true if the given window.pane index is currently active.
func paneExists(paneIndex string) bool {
	out, err := exec.Command("tmux", "list-panes", "-a", "-F", "#{window_index}.#{pane_index}").Output()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if strings.TrimSpace(line) == paneIndex {
			return true
		}
	}
	return false
}

// paneCommand returns the command currently running in the given window.pane index, or empty string on failure.
func paneCommand(paneIndex string) string {
	out, err := exec.Command("tmux", "display-message", "-p", "-t", paneIndex, "#{pane_current_command}").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

