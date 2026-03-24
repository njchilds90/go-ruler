package ruler

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// Operator is the comparison operator used within a Condition.
type Operator string

const (
	OpEquals            Operator = "eq"
	OpNotEquals         Operator = "neq"
	OpGreaterThan       Operator = "gt"
	OpGreaterThanEquals Operator = "gte"
	OpLessThan          Operator = "lt"
	OpLessThanEquals    Operator = "lte"
	OpContains          Operator = "contains"
	OpNotContains       Operator = "not_contains"
	OpIn                Operator = "in"
	OpNotIn             Operator = "not_in"
	OpMatches           Operator = "matches"
	OpExists            Operator = "exists"
	OpNotExists         Operator = "not_exists"
	OpStartsWith        Operator = "starts_with"
	OpEndsWith          Operator = "ends_with"
	OpBetween           Operator = "between"
)

// FactMap is the set of named facts evaluated against rules.
type FactMap map[string]any

// Condition represents a single comparison between a fact value and an expected value.
type Condition struct {
	Field string   `json:"field"`
	Op    Operator `json:"op"`
	Value any      `json:"value,omitempty"`
}

func (c Condition) label() string { return fmt.Sprintf("%s %s %v", c.Field, c.Op, c.Value) }

func (c Condition) evaluate(facts FactMap) (bool, error) {
	switch c.Op {
	case OpExists:
		_, ok := facts[c.Field]
		return ok, nil
	case OpNotExists:
		_, ok := facts[c.Field]
		return !ok, nil
	}

	factVal, ok := facts[c.Field]
	if !ok {
		return false, nil
	}

	switch c.Op {
	case OpEquals:
		return reflect.DeepEqual(factVal, c.Value), nil
	case OpNotEquals:
		return !reflect.DeepEqual(factVal, c.Value), nil
	case OpGreaterThan:
		return compareNumeric(factVal, c.Value, func(a, b float64) bool { return a > b })
	case OpGreaterThanEquals:
		return compareNumeric(factVal, c.Value, func(a, b float64) bool { return a >= b })
	case OpLessThan:
		return compareNumeric(factVal, c.Value, func(a, b float64) bool { return a < b })
	case OpLessThanEquals:
		return compareNumeric(factVal, c.Value, func(a, b float64) bool { return a <= b })
	case OpContains:
		return evalContains(factVal, c.Value, true)
	case OpNotContains:
		return evalContains(factVal, c.Value, false)
	case OpIn:
		return evalIn(factVal, c.Value, true)
	case OpNotIn:
		return evalIn(factVal, c.Value, false)
	case OpMatches:
		return evalMatches(factVal, c.Value)
	case OpStartsWith:
		return evalPrefixSuffix(factVal, c.Value, true)
	case OpEndsWith:
		return evalPrefixSuffix(factVal, c.Value, false)
	case OpBetween:
		return evalBetween(factVal, c.Value)
	default:
		return false, fmt.Errorf("%w: unknown operator %q", ErrInvalidCondition, c.Op)
	}
}

// Equals creates a Condition that passes when fact[field] == value.
func Equals(field string, value any) Condition {
	return Condition{Field: field, Op: OpEquals, Value: value}
}

// NotEquals creates a Condition that passes when fact[field] != value.
func NotEquals(field string, value any) Condition {
	return Condition{Field: field, Op: OpNotEquals, Value: value}
}

// GreaterThan creates a Condition that passes when fact[field] > value (numeric).
func GreaterThan(field string, value any) Condition {
	return Condition{Field: field, Op: OpGreaterThan, Value: value}
}

// GreaterThanEquals creates a Condition that passes when fact[field] >= value (numeric).
func GreaterThanEquals(field string, value any) Condition {
	return Condition{Field: field, Op: OpGreaterThanEquals, Value: value}
}

// LessThan creates a Condition that passes when fact[field] < value (numeric).
func LessThan(field string, value any) Condition {
	return Condition{Field: field, Op: OpLessThan, Value: value}
}

// LessThanEquals creates a Condition that passes when fact[field] <= value (numeric).
func LessThanEquals(field string, value any) Condition {
	return Condition{Field: field, Op: OpLessThanEquals, Value: value}
}

// Contains creates a Condition that passes when fact[field] contains value.
func Contains(field string, value any) Condition {
	return Condition{Field: field, Op: OpContains, Value: value}
}

// NotContains creates a Condition that passes when fact[field] does NOT contain value.
func NotContains(field string, value any) Condition {
	return Condition{Field: field, Op: OpNotContains, Value: value}
}

// In creates a Condition that passes when fact[field] is present in value ([]any).
func In(field string, value []any) Condition { return Condition{Field: field, Op: OpIn, Value: value} }

// NotIn creates a Condition that passes when fact[field] is NOT in value ([]any).
func NotIn(field string, value []any) Condition {
	return Condition{Field: field, Op: OpNotIn, Value: value}
}

// Matches creates a Condition that passes when fact[field] (string) matches the regex pattern.
func Matches(field string, pattern string) Condition {
	return Condition{Field: field, Op: OpMatches, Value: pattern}
}

// Exists creates a Condition that passes when fact[field] is present in the FactMap.
func Exists(field string) Condition { return Condition{Field: field, Op: OpExists} }

// NotExists creates a Condition that passes when fact[field] is absent from the FactMap.
func NotExists(field string) Condition { return Condition{Field: field, Op: OpNotExists} }

// StartsWith creates a Condition that passes when fact[field] starts with prefix.
func StartsWith(field, prefix string) Condition {
	return Condition{Field: field, Op: OpStartsWith, Value: prefix}
}

// EndsWith creates a Condition that passes when fact[field] ends with suffix.
func EndsWith(field, suffix string) Condition {
	return Condition{Field: field, Op: OpEndsWith, Value: suffix}
}

// Between creates a Condition that passes when lower <= fact[field] <= upper.
func Between(field string, lower, upper any) Condition {
	return Condition{Field: field, Op: OpBetween, Value: []any{lower, upper}}
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	}
	return 0, false
}

func compareNumeric(a, b any, fn func(float64, float64) bool) (bool, error) {
	fa, ok := toFloat64(a)
	if !ok {
		return false, fmt.Errorf("%w: fact value %v (%T) is not numeric", ErrTypeMismatch, a, a)
	}
	fb, ok := toFloat64(b)
	if !ok {
		return false, fmt.Errorf("%w: condition value %v (%T) is not numeric", ErrTypeMismatch, b, b)
	}
	return fn(fa, fb), nil
}

func evalContains(factVal, condVal any, wantContains bool) (bool, error) {
	switch fv := factVal.(type) {
	case string:
		sv, ok := condVal.(string)
		if !ok {
			return false, fmt.Errorf("%w: contains on string field requires string value, got %T", ErrTypeMismatch, condVal)
		}
		result := strings.Contains(fv, sv)
		if wantContains {
			return result, nil
		}
		return !result, nil
	case []any:
		for _, item := range fv {
			if reflect.DeepEqual(item, condVal) {
				return wantContains, nil
			}
		}
		return !wantContains, nil
	default:
		return false, fmt.Errorf("%w: contains requires string or []any fact, got %T", ErrTypeMismatch, factVal)
	}
}

func evalIn(factVal, condVal any, wantIn bool) (bool, error) {
	list, ok := condVal.([]any)
	if !ok {
		return false, fmt.Errorf("%w: 'in' operator requires []any value, got %T", ErrTypeMismatch, condVal)
	}
	for _, item := range list {
		if reflect.DeepEqual(item, factVal) {
			return wantIn, nil
		}
	}
	return !wantIn, nil
}

func evalMatches(factVal, condVal any) (bool, error) {
	s, ok := factVal.(string)
	if !ok {
		return false, fmt.Errorf("%w: matches requires string fact, got %T", ErrTypeMismatch, factVal)
	}
	pattern, ok := condVal.(string)
	if !ok {
		return false, fmt.Errorf("%w: matches requires string pattern, got %T", ErrTypeMismatch, condVal)
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, fmt.Errorf("%w: invalid regex %q: %v", ErrInvalidCondition, pattern, err)
	}
	return re.MatchString(s), nil
}

func evalPrefixSuffix(factVal, condVal any, starts bool) (bool, error) {
	s, ok := factVal.(string)
	if !ok {
		return false, fmt.Errorf("%w: starts_with/ends_with requires string fact, got %T", ErrTypeMismatch, factVal)
	}
	prefix, ok := condVal.(string)
	if !ok {
		return false, fmt.Errorf("%w: starts_with/ends_with requires string value, got %T", ErrTypeMismatch, condVal)
	}
	if starts {
		return strings.HasPrefix(s, prefix), nil
	}
	return strings.HasSuffix(s, prefix), nil
}

func evalBetween(factVal, condVal any) (bool, error) {
	bounds, ok := condVal.([]any)
	if !ok || len(bounds) != 2 {
		return false, fmt.Errorf("%w: between requires []any with [lower, upper]", ErrInvalidCondition)
	}
	lower, ok := toFloat64(bounds[0])
	if !ok {
		return false, fmt.Errorf("%w: between lower bound must be numeric", ErrTypeMismatch)
	}
	upper, ok := toFloat64(bounds[1])
	if !ok {
		return false, fmt.Errorf("%w: between upper bound must be numeric", ErrTypeMismatch)
	}
	value, ok := toFloat64(factVal)
	if !ok {
		return false, fmt.Errorf("%w: between fact value must be numeric", ErrTypeMismatch)
	}
	return value >= lower && value <= upper, nil
}
