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

import "testing"

func TestResolvePlaceholders(t *testing.T) {
	tests := []struct {
		run     string
		name    string
		variant string
		want    string
		wantErr bool
	}{
		{"echo {{name}}", "Bob", "", "echo Bob", false},
		{"echo {{name}}", "Alice", "go", "echo Alice", false},
		{"echo {{variant}}", "Bob", "", "echo ", false},
		{"echo {{variant}}", "Bob", "java", "echo java", false},
		{"echo {{name}} {{variant}}", "Alice", "go", "echo Alice go", false},
		{"docker run -e NAME={{name}} -e V={{variant}} img", "Alice", "go", "docker run -e NAME=Alice -e V=go img", false},
		{"echo hello", "Bob", "", "echo hello", false},
		{"echo {{server_url}}", "Bob", "", "", true},
		{"echo {{name}} {{unknown}}", "Bob", "", "", true},
	}

	for _, tt := range tests {
		got, err := resolvePlaceholders(tt.run, tt.name, tt.variant)
		if tt.wantErr {
			if err == nil {
				t.Errorf("resolvePlaceholders(%q, %q, %q): expected error, got %q", tt.run, tt.name, tt.variant, got)
			}
		} else {
			if err != nil {
				t.Errorf("resolvePlaceholders(%q, %q, %q): unexpected error: %v", tt.run, tt.name, tt.variant, err)
			} else if got != tt.want {
				t.Errorf("resolvePlaceholders(%q, %q, %q): got %q, want %q", tt.run, tt.name, tt.variant, got, tt.want)
			}
		}
	}
}
