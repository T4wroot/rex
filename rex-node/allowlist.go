package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Allowlist controls which commands and paths are permitted
type Allowlist struct {
	AllowedCommands []string `yaml:"allowed_commands"`
	AllowedPaths    []string `yaml:"allowed_paths"`
	DeniedCommands  []string `yaml:"denied_commands"`
}

// LoadAllowlist reads the allowlist YAML file
func LoadAllowlist(path string) (*Allowlist, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read allowlist %q: %w", path, err)
	}

	var al Allowlist
	if err := yaml.Unmarshal(data, &al); err != nil {
		return nil, fmt.Errorf("cannot parse allowlist: %w", err)
	}

	log.Printf("[rex] allowlist loaded — %d allowed, %d denied",
		len(al.AllowedCommands), len(al.DeniedCommands))
	return &al, nil
}

// IsCommandAllowed checks if a command is permitted.
// Denied list takes priority over allowed list.
func (al *Allowlist) IsCommandAllowed(command string) bool {
	cmd := strings.TrimSpace(command)

	// Denied list takes priority
	for _, denied := range al.DeniedCommands {
		if matchPattern(denied, cmd) {
			log.Printf("[rex] DENIED: %q (rule: %q)", cmd, denied)
			return false
		}
	}

	// Check allowed list
	for _, allowed := range al.AllowedCommands {
		if matchPattern(allowed, cmd) {
			return true
		}
	}

	log.Printf("[rex] DENIED: %q (not in allowlist)", cmd)
	return false
}

// IsPathAllowed checks if a file path is permitted for streaming
func (al *Allowlist) IsPathAllowed(path string) bool {
	for _, allowed := range al.AllowedPaths {
		allowedDir := filepath.Clean(allowed)
		cleanPath := filepath.Clean(path)
		if strings.HasPrefix(cleanPath, allowedDir) {
			return true
		}
	}
	log.Printf("[rex] DENIED path: %q", path)
	return false
}

// matchPattern supports trailing wildcard: "docker restart *"
func matchPattern(pattern, command string) bool {
	if pattern == command {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(command, strings.TrimSpace(prefix))
	}
	if strings.HasPrefix(command, pattern+" ") {
		return true
	}
	return false
}
