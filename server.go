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
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
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

	// Check if already running
	if s, err := readServerStatus(statusPath); err == nil {
		if pidRunning(s.PID) {
			return fmt.Errorf("server is already running (pid %d on port %d)", s.PID, s.Port)
		}
	}

	addr := fmt.Sprintf("localhost:%d", port)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

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

	if !pidRunning(s.PID) {
		fmt.Println("Server:   not running (stale status file)")
		return nil
	}

	// Ping /health
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/health", s.Port))
	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Println("Server:   not healthy (process exists but not responding)")
		return nil
	}

	started, _ := time.Parse(time.RFC3339, s.Started)
	uptime := time.Since(started).Round(time.Second)

	fmt.Printf("Server:   running\n")
	fmt.Printf("PID:      %d\n", s.PID)
	fmt.Printf("Address:  localhost:%d\n", s.Port)
	fmt.Printf("Uptime:   %s\n", uptime)
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
