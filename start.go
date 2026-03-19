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
)

func cmdStart(explicit string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: agnt start <name>")
	}
	name := args[0]

	configPath, err := resolveWorkspacePath(explicit, false)
	if err != nil {
		return err
	}
	if configPath == "" {
		return fmt.Errorf("no workspace found (no .agnt.yaml in current directory or any parent up to $HOME)")
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	agent, ok := cfg.Agents[name]
	if !ok {
		return fmt.Errorf("agent %q not found in %s", name, configPath)
	}

	agentType, ok := cfg.Types[agent.Type]
	if !ok {
		return fmt.Errorf("type %q for agent %q is not defined in %s", agent.Type, name, configPath)
	}

	cmd, err := resolvePlaceholders(agentType.Run, name, agent.Variant)
	if err != nil {
		return fmt.Errorf("agent %q: %w", name, err)
	}

	if !paneExists(agent.Pane) {
		return fmt.Errorf("pane %s for agent %q does not exist in the current tmux session", agent.Pane, name)
	}

	if err := exec.Command("tmux", "send-keys", "-t", agent.Pane, cmd, "Enter").Run(); err != nil {
		return fmt.Errorf("failed to send keys to pane %s: %w", agent.Pane, err)
	}

	fmt.Printf("Starting %s in pane %s: %s\n", name, agent.Pane, cmd)
	return nil
}
