package ruler

import "errors"

// Result is the structured output of evaluating a single Rule against a FactMap.
// It is designed to be machine-readable and agent-friendly.
type Result struct {
    // RuleName is the name of the evaluated rule.
    RuleName string `json:"rule_name"
    // RuleDescription is the human-readable description of the rule.
    RuleDescription string `json:"rule_description,omitempty"
    // Matched indicates whether the rule as a whole passed.
    Matched bool `json:"matched"
    // Score is the rule's weight, non-zero only when Matched is true.
    Score float64 `json:"score,omitempty"
    // MatchedConditions lists human-readable descriptions of conditions that passed.
    MatchedConditions []string `json:"matched_conditions,omitempty"
    // Tags are the tags copied from the matched Rule.
    Tags []string `json:"tags,omitempty"
    // Metadata is arbitrary data copied from the matched Rule.
    Metadata map[string]any `json:"metadata,omitempty"
}

// Sentinel errors for structured error handling.
var (
    // ErrInvalidRule is returned when a Rule fails structural validation.
    ErrInvalidRule = errors.New("ruler: invalid rule"
    // ErrInvalidCondition is returned when a Condition uses an unsupported operator or bad value.
    ErrInvalidCondition = errors.New("ruler: invalid condition"
    // ErrDuplicateRule is returned when two rules share the same name.
    ErrDuplicateRule = errors.New("ruler: duplicate rule"
    // ErrTypeMismatch is returned when a condition's fact value and expected value are incompatible types.
    ErrTypeMismatch = errors.New("ruler: type mismatch"
    // ErrContextCanceled is returned when the context is canceled during evaluation.
    ErrContextCanceled = errors.New("ruler: context canceled"
)
