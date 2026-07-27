package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SecurityLevel defines the operating mode of rex-node
type SecurityLevel string

const (
	LevelAllowlist    SecurityLevel = "allowlist"     // Strict: only explicitly allowed commands
	LevelReview       SecurityLevel = "review"        // Review: allow read-only + safe, require confirmation for dangerous
	LevelAutonomous   SecurityLevel = "autonomous"    // Full: execute everything except explicit destructive bans
)

type Allowlist struct {
	Mode            SecurityLevel `yaml:"mode"` // "allowlist" | "review" | "autonomous"
	AllowedCommands []string      `yaml:"allowed_commands"`
	AllowedPaths    []string      `yaml:"allowed_paths"`
	DeniedCommands  []string      `yaml:"denied_commands"`
}

func LoadAllowlist(path string) (*Allowlist, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read allowlist %q: %w", path, err)
	}

	var al Allowlist
	if err := yaml.Unmarshal(data, &al); err != nil {
		return nil, fmt.Errorf("cannot parse allowlist: %w", err)
	}

	// Default to allowlist mode if unassigned
	if al.Mode == "" {
		al.Mode = LevelAllowlist
	}

	log.Printf("[rex-security] Mode: %s | Allowed: %d | Denied: %d",
		strings.ToUpper(string(al.Mode)), len(al.AllowedCommands), len(al.DeniedCommands))
	return &al, nil
}

// IsCommandAllowed evaluates permissions based on mode
func (al *Allowlist) IsCommandAllowed(command string) (bool, string) {
	cmd := strings.TrimSpace(command)

	// Denied commands are ALWAYS blocked regardless of mode
	for _, denied := range al.DeniedCommands {
		if matchPattern(denied, cmd) {
			log.Printf("[rex-security] DENIED: %q (matched deny rule: %q)", cmd, denied)
			return false, fmt.Sprintf("command matches absolute deny rule: %s", denied)
		}
	}

	switch al.Mode {

	case LevelAutonomous:
		// Autonomous mode: Execute EVERYTHING unless explicitly in denied_commands
		log.Printf("[rex-security] AUTONOMOUS ALLOW: %q", cmd)
		return true, ""

	case LevelReview:
		// Review mode: Allow standard/read-only commands; flag unknown/dangerous for review
		if al.isReadOnlyOrSafe(cmd) {
			return true, ""
		}
		// If command is explicitly allowed, run it
		for _, allowed := range al.AllowedCommands {
			if matchPattern(allowed, cmd) {
				return true, ""
			}
		}
		log.Printf("[rex-security] REVIEW REQUIRED: %q", cmd)
		return false, "command requires human review/approval under 'review' mode"

	case LevelAllowlist:
		fallthrough
	default:
		// Strict allowlist mode
		for _, allowed := range al.AllowedCommands {
			if matchPattern(allowed, cmd) {
				return true, ""
			}
		}
		log.Printf("[rex-security] ALLOWLIST DENIED: %q", cmd)
		return false, "command not in explicit allowlist"
	}
}

func (al *Allowlist) IsPathAllowed(path string) bool {
	if al.Mode == LevelAutonomous {
		return true
	}

	for _, allowed := range al.AllowedPaths {
		allowedDir := filepath.Clean(allowed)
		cleanPath := filepath.Clean(path)
		if strings.HasPrefix(cleanPath, allowedDir) {
			return true
		}
	}
	return false
}

// isReadOnlyOrSafe identifies common safe inspection commands for 'review' mode
func (al *Allowlist) isReadOnlyOrSafe(cmd string) bool {
	safePrefixes := []string{
		"ls", "cat", "df", "free", "uptime", "uname", "hostname", "ip",
		"systemctl status", "docker ps", "journalctl", "top", "htop", "ss", "ps",
	}
	for _, p := range safePrefixes {
		if strings.HasPrefix(cmd, p) {
			return true
		}
	}
	return false
}

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
