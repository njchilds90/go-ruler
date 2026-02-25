package ruler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/njchilds90/go-ruler"
)

// --- Condition tests ---

func TestEquals(t *testing.T) {
	tests := []struct {
		name    string
		facts   ruler.FactMap
		want    bool
	}{
		{"match string", ruler.FactMap{"status": "active"}, true},
		{"no match string", ruler.FactMap{"status": "inactive"}, false},
		{"missing field", ruler.FactMap{}, false},
		{"match int", ruler.FactMap{"code": 42}, true},
	}
	conditions := []ruler.Condition{
		ruler.Equals("status", "active"),
		ruler.Equals("status", "active"),
		ruler.Equals("status", "active"),
		ruler.Equals("code", 42),
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ruler.NewEngine()
			e.MustAddRule(ruler.Rule{
				Name:       "r",
				Op:         ruler.OpAnd,
				Conditions: []ruler.Condition{conditions[i]},
			})
			res, err := e.EvaluateMatching(context.Background(), tt.facts)
			if err != nil {
				t.Fatal(err)
			}
			got := len(res) > 0
			if got != tt.want {
				t.Errorf("want %v got %v", tt.want, got)
			}
		})
	}
}

func TestNumericOperators(t *testing.T) {
	tests := []struct {
		name  string
		cond  ruler.Condition
		facts ruler.FactMap
		want  bool
	}{
		{"gt pass", ruler.GreaterThan("age", 18), ruler.FactMap{"age": 21}, true},
		{"gt fail", ruler.GreaterThan("age", 18), ruler.FactMap{"age": 18}, false},
		{"gte pass equal", ruler.GreaterThanEquals("age", 18), ruler.FactMap{"age": 18}, true},
		{"lt pass", ruler.LessThan("score", 100), ruler.FactMap{"score": 50}, true},
		{"lt fail", ruler.LessThan("score", 100), ruler.FactMap{"score": 100}, false},
		{"lte pass equal", ruler.LessThanEquals("score", 100), ruler.FactMap{"score": 100}, true},
		{"float64 gt", ruler.GreaterThan("price", 9.99), ruler.FactMap{"price": 10.5}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ruler.NewEngine()
			e.MustAddRule(ruler.Rule{Name: "r", Op: ruler.OpAnd, Conditions: []ruler.Condition{tt.cond}})
			res, err := e.EvaluateMatching(context.Background(), tt.facts)
			if err != nil {
				t.Fatal(err)
			}
			got := len(res) > 0
			if got != tt.want {
				t.Errorf("want %v got %v", tt.want, got)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name  string
		cond  ruler.Condition
		facts ruler.FactMap
		want  bool
	}{
		{"string contains", ruler.Contains("bio", "gopher"), ruler.FactMap{"bio": "I am a gopher"}, true},
		{"string not contains", ruler.Contains("bio", "rust"), ruler.FactMap{"bio": "I am a gopher"}, false},
		{"not_contains pass", ruler.NotContains("bio", "rust"), ruler.FactMap{"bio": "I am a gopher"}, true},
		{"slice contains", ruler.Contains("roles", "admin"), ruler.FactMap{"roles": []any{"user", "admin"}}, true},
		{"slice not contains", ruler.Contains("roles", "god"), ruler.FactMap{"roles": []any{"user", "admin"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ruler.NewEngine()
			e.MustAddRule(ruler.Rule{Name: "r", Op: ruler.OpAnd, Conditions: []ruler.Condition{tt.cond}})
			res, err := e.EvaluateMatching(context.Background(), tt.facts)
			if err != nil {
				t.Fatal(err)
			}
			got := len(res) > 0
			if got != tt.want {
				t.Errorf("want %v got %v", tt.want, got)
			}
		})
	}
}

func TestIn(t *testing.T) {
	tests := []struct {
		name  string
		cond  ruler.Condition
		facts ruler.FactMap
		want  bool
	}{
		{"in pass", ruler.In("role", []any{"admin", "editor"}), ruler.FactMap{"role": "admin"}, true},
		{"in fail", ruler.In("role", []any{"admin", "editor"}), ruler.FactMap{"role": "viewer"}, false},
		{"not_in pass", ruler.NotIn("role", []any{"banned", "suspended"}), ruler.FactMap{"role": "user"}, true},
		{"not_in fail", ruler.NotIn("role", []any{"banned", "suspended"}), ruler.FactMap{"role": "banned"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ruler.NewEngine()
			e.MustAddRule(ruler.Rule{Name: "r", Op: ruler.OpAnd, Conditions: []ruler.Condition{tt.cond}})
			res, err := e.EvaluateMatching(context.Background(), tt.facts)
			if err != nil {
				t.Fatal(err)
			}
			got := len(res) > 0
			if got != tt.want {
				t.Errorf("want %v got %v", tt.want, got)
			}
		})
	}
}

func TestMatches(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		fact    string
		want    bool
	}{
		{"valid email", `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`, "test@example.com", true},
		{"invalid email", `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`, "not-an-email", false},
		{"uuid pattern", `^[0-9a-f]{8}-[0-9a-f]{4}`, "550e8400-e29b-41d4", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ruler.NewEngine()
			e.MustAddRule(ruler.Rule{
				Name: "r", Op: ruler.OpAnd,
				Conditions: []ruler.Condition{ruler.Matches("val", tt.pattern)},
			})
			res, err := e.EvaluateMatching(context.Background(), ruler.FactMap{"val": tt.fact})
			if err != nil {
				t.Fatal(err)
			}
			got := len(res) > 0
			if got != tt.want {
				t.Errorf("want %v got %v", tt.want, got)
			}
		})
	}
}

func TestExistsNotExists(t *testing.T) {
	e := ruler.NewEngine()
	e.MustAddRule(ruler.Rule{Name: "has-email", Op: ruler.OpAnd, Conditions: []ruler.Condition{ruler.Exists("email")}})
	e.MustAddRule(ruler.Rule{Name: "no-ban", Op: ruler.OpAnd, Conditions: []ruler.Condition{ruler.NotExists("ban_reason")}})

	facts := ruler.FactMap{"email": "a@b.com"}
	results, err := e.EvaluateMatching(context.Background(), facts)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 matched rules, got %d", len(results))
	}
}

// --- Op tests ---

func TestOpAnd(t *testing.T) {
	e := ruler.NewEngine()
	e.MustAddRule(ruler.Rule{
		Name: "both",
		Op:   ruler.OpAnd,
		Conditions: []ruler.Condition{
			ruler.Equals("a", 1),
			ruler.Equals("b", 2),
		},
	})

	t.Run("both pass", func(t *testing.T) {
		res, _ := e.EvaluateMatching(context.Background(), ruler.FactMap{"a": 1, "b": 2})
		if len(res) != 1 {
			t.Error("expected match")
		}
	})
	t.Run("one fail", func(t *testing.T) {
		res, _ := e.EvaluateMatching(context.Background(), ruler.FactMap{"a": 1, "b": 9})
		if len(res) != 0 {
			t.Error("expected no match")
		}
	})
}

func TestOpOr(t *testing.T) {
	e := ruler.NewEngine()
	e.MustAddRule(ruler.Rule{
		Name: "either",
		Op:   ruler.OpOr,
		Conditions: []ruler.Condition{
			ruler.Equals("a", 1),
			ruler.Equals("b", 2),
		},
	})

	t.Run("first matches", func(t *testing.T) {
		res, _ := e.EvaluateMatching(context.Background(), ruler.FactMap{"a": 1})
		if len(res) != 1 {
			t.Error("expected match")
		}
	})
	t.Run("neither matches", func(t *testing.T) {
		res, _ := e.EvaluateMatching(context.Background(), ruler.FactMap{"a": 9, "b": 9})
		if len(res) != 0 {
			t.Error("expected no match")
		}
	})
}

// --- Priority tests ---

func TestPriority(t *testing.T) {
	e := ruler.NewEngine()
	e.MustAddRule(ruler.Rule{
		Name: "low", Priority: 1,
		Op:         ruler.OpAnd,
		Conditions: []ruler.Condition{ruler.Equals("x", 1)},
	})
	e.MustAddRule(ruler.Rule{
		Name: "high", Priority: 10,
		Op:         ruler.OpAnd,
		Conditions: []ruler.Condition{ruler.Equals("x", 1)},
	})

	res, err := e.EvaluateMatching(context.Background(), ruler.FactMap{"x": 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 results got %d", len(res))
	}
	if res[0].RuleName != "high" {
		t.Errorf("expected high-priority first, got %q", res[0].RuleName)
	}
}

// --- EvaluateFirst tests ---

func TestEvaluateFirst(t *testing.T) {
	e := ruler.NewEngine()
	e.MustAddRule(ruler.Rule{
		Name: "vip", Priority: 100,
		Op:         ruler.OpAnd,
		Conditions: []ruler.Condition{ruler.Equals("tier", "vip")},
	})
	e.MustAddRule(ruler.Rule{
		Name: "standard", Priority: 1,
		Op:         ruler.OpAnd,
		Conditions: []ruler.Condition{ruler.Equals("tier", "standard")},
	})

	t.Run("first match is vip", func(t *testing.T) {
		res, ok, err := e.EvaluateFirst(context.Background(), ruler.FactMap{"tier": "vip"})
		if err != nil || !ok {
			t.Fatalf("expected match; err=%v ok=%v", err, ok)
		}
		if res.RuleName != "vip" {
			t.Errorf("expected vip, got %q", res.RuleName)
		}
	})
	t.Run("no match", func(t *testing.T) {
		_, ok, err := e.EvaluateFirst(context.Background(), ruler.FactMap{"tier": "guest"})
		if err != nil || ok {
			t.Errorf("expected no match; err=%v ok=%v", err, ok)
		}
	})
}

// --- TotalScore tests ---

func TestTotalScore(t *testing.T) {
	e := ruler.NewEngine()
	e.MustAddRule(ruler.Rule{
		Name: "risk-new-account", Score: 30,
		Op:         ruler.OpAnd,
		Conditions: []ruler.Condition{ruler.Equals("new_account", true)},
	})
	e.MustAddRule(ruler.Rule{
		Name: "risk-foreign-ip", Score: 50,
		Op:         ruler.OpAnd,
		Conditions: []ruler.Condition{ruler.Equals("foreign_ip", true)},
	})

	score, err := e.TotalScore(context.Background(), ruler.FactMap{
		"new_account": true,
		"foreign_ip":  true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if score != 80 {
		t.Errorf("expected score 80, got %v", score)
	}
}

// --- Validation / error tests ---

func TestDuplicateRule(t *testing.T) {
	e := ruler.NewEngine()
	r := ruler.Rule{Name: "dup", Op: ruler.OpAnd, Conditions: []ruler.Condition{ruler.Exists("x")}}
	if err := e.AddRule(r); err != nil {
		t.Fatal(err)
	}
	err := e.AddRule(r)
	if !errors.Is(err, ruler.ErrDuplicateRule) {
		t.Errorf("expected ErrDuplicateRule, got %v", err)
	}
}

func TestInvalidRule_NoName(t *testing.T) {
	e := ruler.NewEngine()
	err := e.AddRule(ruler.Rule{Op: ruler.OpAnd, Conditions: []ruler.Condition{ruler.Exists("x")}})
	if !errors.Is(err, ruler.ErrInvalidRule) {
		t.Errorf("expected ErrInvalidRule, got %v", err)
	}
}

func TestInvalidRule_NoConditions(t *testing.T) {
	e := ruler.NewEngine()
	err := e.AddRule(ruler.Rule{Name: "empty", Op: ruler.OpAnd})
	if !errors.Is(err, ruler.ErrInvalidRule) {
		t.Errorf("expected ErrInvalidRule, got %v", err)
	}
}

func TestTypeMismatch(t *testing.T) {
	e := ruler.NewEngine()
	e.MustAddRule(ruler.Rule{
		Name: "r", Op: ruler.OpAnd,
		Conditions: []ruler.Condition{ruler.GreaterThan("name", 10)},
	})
	_, err := e.EvaluateAll(context.Background(), ruler.FactMap{"name": "alice"})
	if !errors.Is(err, ruler.ErrTypeMismatch) {
		t.Errorf("expected ErrTypeMismatch, got %v", err)
	}
}

func TestContextCanceled(t *testing.T) {
	e := ruler.NewEngine()
	for i := 0; i < 5; i++ {
		e.MustAddRule(ruler.Rule{
			Name: fmt.Sprintf("r%d", i), Op: ruler.OpAnd,
			Conditions: []ruler.Condition{ruler.Exists("x")},
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := e.EvaluateAll(ctx, ruler.FactMap{"x": 1})
	if !errors.Is(err, ruler.ErrContextCanceled) {
		t.Errorf("expected ErrContextCanceled, got %v", err)
	}
}

func TestEvaluateAll_ReturnsAll(t *testing.T) {
	e := ruler.NewEngine()
	e.MustAddRule(ruler.Rule{Name: "match", Op: ruler.OpAnd, Conditions: []ruler.Condition{ruler.Equals("x", 1)}})
	e.MustAddRule(ruler.Rule{Name: "no-match", Op: ruler.OpAnd, Conditions: []ruler.Condition{ruler.Equals("x", 2)}})

	res, err := e.EvaluateAll(context.Background(), ruler.FactMap{"x": 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 {
		t.Errorf("expected 2 results, got %d", len(res))
	}
}

func TestRuleNames(t *testing.T) {
	e := ruler.NewEngine()
	e.MustAddRule(ruler.Rule{Name: "b", Priority: 1, Op: ruler.OpAnd, Conditions: []ruler.Condition{ruler.Exists("x")}})
	e.MustAddRule(ruler.Rule{Name: "a", Priority: 10, Op: ruler.OpAnd, Conditions: []ruler.Condition{ruler.Exists("x")}})
	names := e.RuleNames()
	if names[0] != "a" || names[1] != "b" {
		t.Errorf("expected [a b], got %v", names)
	}
}

func fmt_Sprintf(format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}

// needed for inline fmt usage in test
var _ = fmt_Sprintf
