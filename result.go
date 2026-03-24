package ruler

import (
	"errors"
	"time"
)

// Result is the structured output of evaluating a single Rule against a FactMap.
type Result struct {
	RuleName          string         `json:"rule_name"`
	RuleDescription   string         `json:"rule_description,omitempty"`
	Matched           bool           `json:"matched"`
	Score             float64        `json:"score,omitempty"`
	MatchedConditions []string       `json:"matched_conditions,omitempty"`
	Tags              []string       `json:"tags,omitempty"`
	Metadata          map[string]any `json:"metadata,omitempty"`
}

// AgentDecision is the top-level structured output intended for autonomous agents.
type AgentDecision struct {
	Action       string         `json:"action"`
	Allowed      bool           `json:"allowed"`
	Score        float64        `json:"score"`
	TraceID      string         `json:"trace_id,omitempty"`
	Reasons      []string       `json:"reasons,omitempty"`
	MatchedRules []Result       `json:"matched_rules,omitempty"`
	AllRules     []Result       `json:"all_rules,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	EvaluatedAt  time.Time      `json:"evaluated_at"`
}

// WhatIfResult contains current and candidate decisions for side-by-side analysis.
type WhatIfResult struct {
	Current   AgentDecision `json:"current"`
	Candidate AgentDecision `json:"candidate"`
	Delta     float64       `json:"delta"`
}

var (
	ErrInvalidRule      = errors.New("ruler: invalid rule")
	ErrInvalidCondition = errors.New("ruler: invalid condition")
	ErrDuplicateRule    = errors.New("ruler: duplicate rule")
	ErrTypeMismatch     = errors.New("ruler: type mismatch")
	ErrContextCanceled  = errors.New("ruler: context canceled")
	ErrEngineFrozen     = errors.New("ruler: engine is frozen")
)
