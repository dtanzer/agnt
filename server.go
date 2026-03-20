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
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultPort = 7717
const serverStatusFile = ".agnt-server.yaml"

type serverStatus struct {
	PID     int    `yaml:"pid"`
	Port    int    `yaml:"port"`
	Started string `yaml:"started"`
}

func serverStatusPath(workspaceConfig string) (string, error) {
	configPath, err := resolveWorkspacePath(workspaceConfig, false)
	if err != nil {
		return "", err
	}
	if configPath == "" {
		return "", fmt.Errorf("no workspace found (no .agnt.yaml in current directory or any parent up to $HOME)")
	}
	return filepath.Join(filepath.Dir(configPath), serverStatusFile), nil
}

func readServerStatus(statusPath string) (*serverStatus, error) {
	data, err := os.ReadFile(statusPath)
	if err != nil {
		return nil, err
	}
	var s serverStatus
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func writeServerStatus(statusPath string, s serverStatus) error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(statusPath, data, 0644)
}

func pidRunning(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

func serverLog(format string, args ...any) {
	fmt.Printf("[%s] %s\n", time.Now().UTC().Format(time.RFC3339), fmt.Sprintf(format, args...))
}

func handleAgentStart(configPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract agent name from URL path: /agents/{name}/start
		path := strings.TrimPrefix(r.URL.Path, "/agents/")
		path = strings.TrimSuffix(path, "/start")
		name := path

		writeJSONError := func(status int, msg string) {
			serverLog("start %q: %s", name, msg)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			json.NewEncoder(w).Encode(map[string]string{"error": msg})
		}

		cfg, err := loadConfig(configPath)
		if err != nil {
			writeJSONError(http.StatusInternalServerError, fmt.Sprintf("failed to load config: %s", err))
			return
		}

		agent, ok := cfg.Agents[name]
		if !ok {
			writeJSONError(http.StatusNotFound, fmt.Sprintf("agent %q not found", name))
			return
		}

		agentType, ok := cfg.Types[agent.Type]
		if !ok {
			writeJSONError(http.StatusInternalServerError, fmt.Sprintf("type %q for agent %q is not defined", agent.Type, name))
			return
		}

		cmd, err := resolvePlaceholders(agentType.Run, name, agent.Variant)
		if err != nil {
			writeJSONError(http.StatusInternalServerError, fmt.Sprintf("agent %q: %s", name, err))
			return
		}

		if !paneExists(agent.Pane) {
			writeJSONError(http.StatusUnprocessableEntity, fmt.Sprintf("pane %s for agent %q does not exist in the current tmux session", agent.Pane, name))
			return
		}

		if err := exec.Command("tmux", "send-keys", "-t", agent.Pane, cmd, "Enter").Run(); err != nil {
			writeJSONError(http.StatusInternalServerError, fmt.Sprintf("failed to send keys to pane %s: %s", agent.Pane, err))
			return
		}

		serverLog("started %q in pane %s: %s", name, agent.Pane, cmd)

		type startResponse struct {
			Name    string `json:"name"`
			Pane    string `json:"pane"`
			Command string `json:"command"`
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(startResponse{Name: name, Pane: agent.Pane, Command: cmd})
	}
}

func cmdServerStart(workspaceConfig string, args []string) error {
	if os.Getenv("TMUX") == "" {
		return fmt.Errorf("agnt server must be run inside a tmux session")
	}

	port := defaultPort
	for i := 0; i < len(args); i++ {
		if args[i] == "--port" {
			if i+1 >= len(args) {
				return fmt.Errorf("--port requires a value")
			}
			p, err := strconv.Atoi(args[i+1])
			if err != nil {
				return fmt.Errorf("invalid port: %s", args[i+1])
			}
			port = p
			i++
		}
	}

	statusPath, err := serverStatusPath(workspaceConfig)
	if err != nil {
		return err
	}

	configPath, err := resolveWorkspacePath(workspaceConfig, false)
	if err != nil {
		return err
	}
	if configPath == "" {
		return fmt.Errorf("no workspace found (no .agnt.yaml in current directory or any parent up to $HOME)")
	}

	// Check if already running
	if s, err := readServerStatus(statusPath); err == nil {
		if pidRunning(s.PID) {
			return fmt.Errorf("server is already running (pid %d on port %d)", s.PID, s.Port)
		}
	}

	addr := fmt.Sprintf("localhost:%d", port)
	startedTime := time.Now().UTC()

	type healthResponse struct {
		PID     int    `json:"pid"`
		Port    int    `json:"port"`
		Started string `json:"started"`
		Uptime  string `json:"uptime"`
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		resp := healthResponse{
			PID:     os.Getpid(),
			Port:    port,
			Started: startedTime.Format(time.RFC3339),
			Uptime:  time.Since(startedTime).Round(time.Second).String(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/agents/", handleAgentStart(configPath))

	// Enforce localhost-only connections
	srv := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil || (host != "127.0.0.1" && host != "::1") {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			mux.ServeHTTP(w, r)
		}),
	}

	if err := writeServerStatus(statusPath, serverStatus{
		PID:     os.Getpid(),
		Port:    port,
		Started: time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		return fmt.Errorf("failed to write status file: %w", err)
	}

	// Remove status file on shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		os.Remove(statusPath)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	fmt.Printf("agnt server listening on %s\n", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		os.Remove(statusPath)
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func cmdServerStatus(workspaceConfig string) error {
	statusPath, err := serverStatusPath(workspaceConfig)
	if err != nil {
		return err
	}

	s, err := readServerStatus(statusPath)
	if err != nil {
		fmt.Println("Server:   not running")
		return nil
	}

	healthURL := fmt.Sprintf("http://localhost:%d/health", s.Port)
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(healthURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Println("Server:   not running")
		fmt.Printf("Checked:  %s\n", healthURL)
		return nil
	}
	defer resp.Body.Close()

	type healthResponse struct {
		PID     int    `json:"pid"`
		Port    int    `json:"port"`
		Started string `json:"started"`
		Uptime  string `json:"uptime"`
	}
	var health healthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		fmt.Println("Server:   not running")
		fmt.Printf("Checked:  %s\n", healthURL)
		return nil
	}

	fmt.Printf("Server:   running\n")
	fmt.Printf("PID:      %d\n", health.PID)
	fmt.Printf("Address:  localhost:%d\n", health.Port)
	fmt.Printf("Uptime:   %s\n", health.Uptime)
	return nil
}

func cmdServer(workspaceConfig string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: agnt server <start|status>")
	}
	switch args[0] {
	case "start":
		return cmdServerStart(workspaceConfig, args[1:])
	case "status":
		return cmdServerStatus(workspaceConfig)
	default:
		return fmt.Errorf("unknown server subcommand %q", args[0])
	}
}
