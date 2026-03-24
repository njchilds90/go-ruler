// Package ruler provides a declarative, zero-dependency rule engine for Go.
package ruler

import (
	"context"
	"fmt"
	"sort"
)

type ConditionOp string

const (
	OpAnd ConditionOp = "AND"
	OpOr  ConditionOp = "OR"
)

type Priority int

// Rule defines a named, weighted policy that is evaluated against a FactMap.
type Rule struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Priority    Priority       `json:"priority,omitempty"`
	Score       float64        `json:"score,omitempty"`
	Op          ConditionOp    `json:"op,omitempty"`
	Conditions  []Condition    `json:"conditions"`
	Tags        []string       `json:"tags,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

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

type byPriority []Rule

func (b byPriority) Len() int      { return len(b) }
func (b byPriority) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b byPriority) Less(i, j int) bool {
	if b[i].Priority == b[j].Priority {
		return b[i].Name < b[j].Name
	}
	return b[i].Priority > b[j].Priority
}

func sortRules(rules []Rule) []Rule {
	cp := make([]Rule, len(rules))
	copy(cp, rules)
	sort.Sort(byPriority(cp))
	return cp
}
