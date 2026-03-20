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
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func cmdStart(explicit string, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: agnt start <name>")
	}
	name := args[0]

	statusPath, err := serverStatusPath(explicit)
	if err != nil {
		return fmt.Errorf("agnt start requires a running server — run \"agnt server start\" first")
	}

	s, err := readServerStatus(statusPath)
	if err != nil {
		return fmt.Errorf("agnt start requires a running server — run \"agnt server start\" first")
	}

	url := fmt.Sprintf("http://localhost:%d/agents/%s/start", s.Port, name)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to contact server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		type startResponse struct {
			Name    string `json:"name"`
			Pane    string `json:"pane"`
			Command string `json:"command"`
		}
		var result startResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to decode server response: %w", err)
		}
		fmt.Printf("Starting %s in pane %s: %s\n", result.Name, result.Pane, result.Command)
		return nil
	}

	var errResp struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}
	return fmt.Errorf("%s", errResp.Error)
}
