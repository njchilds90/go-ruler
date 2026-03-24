package ruler

import (
	"encoding/json"
	"fmt"
	"os"
)

// RuleFile describes a serialized collection of rules.
type RuleFile struct {
	Version string `json:"version"`
	Rules   []Rule `json:"rules"`
}

// LoadRulesFile reads rules from a JSON file path.
func LoadRulesFile(path string) ([]Rule, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read rule file: %w", err)
	}
	var rf RuleFile
	if err := json.Unmarshal(b, &rf); err != nil {
		return nil, fmt.Errorf("decode rule file: %w", err)
	}
	return rf.Rules, nil
}

// SaveRulesFile writes rules to a JSON file.
func SaveRulesFile(path string, rules []Rule) error {
	b, err := json.MarshalIndent(RuleFile{Version: "v1.1.0", Rules: rules}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode rule file: %w", err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("write rule file: %w", err)
	}
	return nil
}
