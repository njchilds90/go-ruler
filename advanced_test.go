package ruler_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	ruler "github.com/njchilds90/go-ruler"
)

func TestAdvancedOperatorsAndWhatIf(t *testing.T) {
	e := ruler.NewEngine(ruler.WithCache())
	e.MustAddRule(ruler.NewRule("allow:high").Priority(10).Score(50).
		Condition(ruler.StartsWith("env", "prod")).Condition(ruler.Between("risk", 0, 50)).Build())
	e.MustAddRule(ruler.NewRule("allow:suffix").Priority(5).Score(10).
		Condition(ruler.EndsWith("email", "@example.com")).Build())
	e.Freeze()

	whatIf, err := e.EvaluateWhatIf(context.Background(),
		ruler.FactMap{"env": "staging", "risk": 30, "email": "a@x.com"},
		ruler.FactMap{"env": "prod-eu", "risk": 30, "email": "a@example.com"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !whatIf.Candidate.Allowed || whatIf.Candidate.Score != 60 {
		t.Fatalf("unexpected candidate decision: %#v", whatIf.Candidate)
	}
	if whatIf.Delta <= 0 {
		t.Fatalf("expected positive delta, got %f", whatIf.Delta)
	}
}

func TestRuleFileRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "rules.json")
	rules := []ruler.Rule{ruler.NewRule("allow:admin").Condition(ruler.Equals("role", "admin")).Build()}
	if err := ruler.SaveRulesFile(path, rules); err != nil {
		t.Fatal(err)
	}
	got, err := ruler.LoadRulesFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "allow:admin" {
		t.Fatalf("unexpected loaded rules: %#v", got)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkEvaluateDecision(b *testing.B) {
	e := ruler.NewEngine(ruler.WithCache())
	for i := range 100 {
		e.MustAddRule(ruler.NewRule(fmt.Sprintf("r-%d", i)).Priority(ruler.Priority(i)).Score(1).Condition(ruler.GreaterThan("x", i)).Build())
	}
	facts := ruler.FactMap{"x": 1000}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := e.EvaluateDecision(context.Background(), facts)
		if err != nil {
			b.Fatal(err)
		}
	}
}
