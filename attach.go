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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func cmdAttach(workspaceConfig string, args []string) error {
	// Parse --podman flag
	podmanContainer := ""
	for i := 0; i < len(args); i++ {
		if args[i] == "--podman" {
			if i+1 >= len(args) {
				return fmt.Errorf("--podman requires a value")
			}
			podmanContainer = args[i+1]
			i++
		} else {
			return fmt.Errorf("unexpected argument: %q", args[i])
		}
	}
	if podmanContainer == "" {
		return fmt.Errorf("usage: agnt attach --podman <container-name>")
	}

	// Must be inside a tmux session
	if os.Getenv("TMUX") == "" {
		fmt.Fprintf(os.Stderr, "agnt attach must be run inside a tmux session\n")
		os.Exit(1)
	}

	// Discover server via status file
	statusPath, err := serverStatusPath(workspaceConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "agnt attach requires a running server — run \"agnt server start\" first\n")
		os.Exit(1)
	}
	s, err := readServerStatus(statusPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "agnt attach requires a running server — run \"agnt server start\" first\n")
		os.Exit(1)
	}

	// Get current pane index
	pane, err := currentTmuxPane()
	if err != nil {
		return fmt.Errorf("failed to determine current tmux pane: %w", err)
	}

	// Look up which agent is registered to this pane
	configPath, err := resolveWorkspacePath(workspaceConfig, false)
	if err != nil {
		return fmt.Errorf("failed to find workspace: %w", err)
	}
	if configPath == "" {
		return fmt.Errorf("no workspace found (no .agnt.yaml in current directory or any parent up to $HOME)")
	}
	cfg, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	agentName := ""
	for name, agent := range cfg.Agents {
		if agent.Pane == pane {
			agentName = name
			break
		}
	}
	if agentName == "" {
		fmt.Fprintf(os.Stderr, "no agent registered to the current tmux pane\n")
		os.Exit(1)
	}

	// Call POST /agents/{name}/attach
	url := fmt.Sprintf("http://localhost:%d/agents/%s/attach", s.Port, agentName)
	body, _ := json.Marshal(map[string]string{"podman": podmanContainer})
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to contact server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("Attached %s (podman: %s)\n", agentName, podmanContainer)
		return nil
	}

	var errResp struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	fmt.Fprintf(os.Stderr, "%s\n", errResp.Error)
	os.Exit(1)
	return nil
}
