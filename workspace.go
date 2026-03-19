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

func mustGetwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}
