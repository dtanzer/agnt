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
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const configFileName = ".agnt.yaml"

type Agent struct {
	Type    string `yaml:"type"`
	Variant string `yaml:"variant,omitempty"`
	Pane    string `yaml:"pane"`
}

type Config struct {
	Agents map[string]Agent `yaml:"agents"`
}

// findWorkspace searches for .agnt.yaml walking up from cwd toward $HOME.
// Returns the path if found, empty string if not.
func findWorkspace() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot determine current directory: %w", err)
	}

	for {
		candidate := filepath.Join(dir, configFileName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		if dir == home {
			break
		}
		next := filepath.Dir(dir)
		if next == dir {
			break
		}
		dir = next
	}
	return "", nil
}

func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	if cfg.Agents == nil {
		cfg.Agents = map[string]Agent{}
	}
	return cfg, nil
}

func saveConfig(path string, cfg Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

func cmdNewWorkspace() error {
	if _, err := os.Stat(configFileName); err == nil {
		return fmt.Errorf("%s already exists in the current directory", configFileName)
	}

	cfg := Config{Agents: map[string]Agent{}}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	if err := os.WriteFile(configFileName, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", configFileName, err)
	}

	fmt.Printf("Created workspace at %s/%s\n", mustGetwd(), configFileName)
	return nil
}

func currentTmuxPane() (string, error) {
	if os.Getenv("TMUX") == "" {
		return "", fmt.Errorf("not inside a tmux session")
	}
	out, err := exec.Command("tmux", "display-message", "-p", "#{pane_id}").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get tmux pane ID: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func cmdRegister(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: agnt register <name> <type> [--variant <variant>]")
	}

	name := args[0]
	agentType := args[1]
	variant := ""

	for i := 2; i < len(args); i++ {
		if args[i] == "--variant" {
			if i+1 >= len(args) {
				return fmt.Errorf("--variant requires a value")
			}
			variant = args[i+1]
			i++
		} else {
			return fmt.Errorf("unexpected argument: %q", args[i])
		}
	}

	pane, err := currentTmuxPane()
	if err != nil {
		return err
	}

	configPath, err := findWorkspace()
	if err != nil {
		return err
	}

	if configPath == "" {
		// No workspace found — create one in the current directory
		configPath = filepath.Join(mustGetwd(), configFileName)
		cfg := Config{Agents: map[string]Agent{}}
		if err := saveConfig(configPath, cfg); err != nil {
			return fmt.Errorf("failed to create workspace: %w", err)
		}
		fmt.Printf("Created workspace at %s\n", configPath)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	if _, exists := cfg.Agents[name]; exists {
		return fmt.Errorf("agent %q already exists in %s", name, configPath)
	}

	cfg.Agents[name] = Agent{Type: agentType, Variant: variant, Pane: pane}

	if err := saveConfig(configPath, cfg); err != nil {
		return err
	}

	fmt.Printf("Registered agent %q (type: %s", name, agentType)
	if variant != "" {
		fmt.Printf(", variant: %s", variant)
	}
	fmt.Printf(") in %s\n", configPath)
	return nil
}

func mustGetwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}
