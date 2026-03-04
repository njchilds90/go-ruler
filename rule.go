// Package ruler provides a declarative, zero-dependency rule engine for Go.
// It enables you to define named rules composed of typed conditions, evaluate
// them against arbitrary fact maps, and receive structured, explainable results.
//
// go-ruler is designed for deterministic behavior, machine-readable outputs,
// and AI-agent-friendly evaluation pipelines.
package ruler

import (
	"context"
	"errors"
	"fmt"
	"sort"
)

// ConditionOp defines the logical operator used to combine multiple conditions
// within a single Rule.
type ConditionOp string

const (
	// OpAnd requires all conditions to pass for the rule to match.
	OpAnd ConditionOp = "AND"
	// OpOr requires at least one condition to pass for the rule to match.
	OpOr ConditionOp = "OR"
)

// Priority determines the order in which rules are evaluated and returned.
// Higher values are evaluated first.
type Priority int

// Rule defines a named, weighted policy that is evaluated against a FactMap.
// A Rule consists of one or more Conditions combined with a logical operator.
//
// Example:
//
//	rule := ruler.Rule{
//		Name:       "high-value-customer",
//		Priority:   10,
//		Op:         ruler.OpAnd,
//		Conditions: []ruler.Condition{
//			ruler.GreaterThan("total_spend", 1000.0),
//			ruler.Equals("status", "active"),
//		},
//	}
type Rule struct {
	// Name is the unique identifier for this rule.
	Name string
	// Description is a human-readable explanation of what this rule represents.
	Description string
	// Priority controls evaluation order. Higher = evaluated first.
	Priority Priority
	// Score is an optional numeric weight assigned when this rule matches.
	Score float64
	// Op is the logical operator (AND/OR) used to combine Conditions.
	Op ConditionOp
	// Conditions is the list of conditions that must be satisfied.
	Conditions []Condition
	// Tags are optional labels for grouping or filtering rules.
	Tags []string
	// Metadata holds arbitrary key-value data attached to a rule.
	Metadata map[string]any
}

var ErrInvalidRule = errors.New("invalid rule")
var ErrContextCanceled = errors.New("context canceled")

// validate checks that a Rule is well-formed.
func (r Rule) validate() error {
	if r.Name == "" {
		return fmt.Errorf("%w: rule name must not be empty", ErrInvalidRule)
	}
	if len(r.Conditions) == 0 {
		return fmt.Errorf("%w: rule %q has no conditions", ErrInvalidRule, r.Name)
	}
	op := r.Op
	if op == "" {
		op = OpAnd
	}
	if op != OpAnd && op != OpOr {
		return fmt.Errorf("%w: rule %q has invalid Op %q", ErrInvalidRule, r.Name, r.Op)
	}
	return nil
}

// evaluate runs all conditions against facts using the rule's Op.
// Returns matched conditions and whether the rule as a whole matched.
func (r Rule) evaluate(ctx context.Context, facts FactMap) (matched []string, ok bool, err error) {
	op := r.Op
	if op == "" {
		op = OpAnd
	}

	passed := 0
	for _, c := range r.Conditions {
		select {
		case <-ctx.Done():
			return nil, false, fmt.Errorf("%w: %w", ErrContextCanceled, ctx.Err())
		default:
		}

		hit, evalErr := c.evaluate(facts)
		if evalErr != nil {
			return nil, false, fmt.Errorf("rule %q condition %q: %w", r.Name, c.Field, evalErr)
		}
		if hit {
			passed++
			matched = append(matched, c.label())
		}
	}

	total := len(r.Conditions)
	switch op {
	case OpAnd:
		ok = passed == total
	case OpOr:
		ok = passed > 0
	}
	return matched, ok, nil
}

// byPriority sorts rules in descending priority order.
type byPriority []Rule

func (b byPriority) Len() int           { return len(b) }
func (b byPriority) Less(i, j int) bool { return b[i].Priority > b[j].Priority }
func (b byPriority) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

// sortRules returns a copy of rules sorted by descending priority.
func sortRules(rules []Rule) []Rule {
	cp := make([]Rule, len(rules))
	copy(cp, rules)
	if len(cp) > 0 {
		sort.Sort(byPriority(cp))
	}
	return cp
}
