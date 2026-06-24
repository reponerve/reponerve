package discipline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const policyFileName = "discipline-policy.json"

// PolicyPath returns the discipline policy path inside a workspace directory.
func PolicyPath(workspaceDir string) string {
	return filepath.Join(workspaceDir, policyFileName)
}

// WritePolicy persists policy to the workspace.
func WritePolicy(workspaceDir string, policy *Policy) error {
	if policy == nil {
		return fmt.Errorf("policy cannot be nil")
	}
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		return fmt.Errorf("create workspace: %w", err)
	}
	data, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal policy: %w", err)
	}
	path := PolicyPath(workspaceDir)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write policy temp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename policy: %w", err)
	}
	return nil
}

// LoadPolicy reads policy from the workspace. Missing file returns nil, nil.
func LoadPolicy(workspaceDir string) (*Policy, error) {
	path := PolicyPath(workspaceDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read policy: %w", err)
	}
	var policy Policy
	if err := json.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("parse policy: %w", err)
	}
	return &policy, nil
}
