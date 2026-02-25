package ruler

import (
	"context"
	"fmt"
)

// Engine holds a registered set of Rules and evaluates them against FactMaps.
// An Engine is safe for concurrent read use after construction. It must not be
// mutated after the first call to Evaluate or EvaluateAll.
//
// Example:
//
//	e := ruler.NewEngine()
//	e.AddRule(ruler.Rule{
//	    Name: "adult",
//	    Op:   ruler.OpAnd,
//	    Conditions: []ruler.Condition{
//	        ruler.GreaterThanEquals("age", 18),
//	    },
//	})
//	results, err := e.EvaluateAll(context.Background(), ruler.FactMap{"age": 21})
type Engine struct {
	rules []Rule
}

// NewEngine creates an empty Engine ready for rule registration.
//
// Example:
//
//	e := ruler.NewEngine()
func NewEngine() *Engine {
	return &Engine{}
}

// AddRule registers a Rule with the Engine. Returns an error if the rule is invalid
// or if a rule with the same Name has already been added.
//
// Example:
//
//	err := e.AddRule(ruler.Rule{
//	    Name: "premium",
//	    Op:   ruler.OpAnd,
//	    Conditions: []ruler.Condition{
//	        ruler.Equals("plan", "premium"),
//	    },
//	})
func (e *Engine) AddRule(r Rule) error {
	if err := r.validate(); err != nil {
		return err
	}
	for _, existing := range e.rules {
		if existing.Name == r.Name {
			return fmt.Errorf("%w: rule %q already registered", ErrDuplicateRule, r.Name)
		}
	}
	e.rules = append(e.rules, r)
	return nil
}

// MustAddRule registers a Rule and panics if the rule is invalid or duplicated.
// Intended for use in package-level var blocks or test setup.
//
// Example:
//
//	e.MustAddRule(ruler.Rule{Name: "vip", Op: ruler.OpAnd, Conditions: []ruler.Condition{ruler.Equals("tier", "vip")}})
func (e *Engine) MustAddRule(r Rule) {
	if err := e.AddRule(r); err != nil {
		panic(err)
	}
}

// RuleCount returns the number of rules registered in the Engine.
func (e *Engine) RuleCount() int {
	return len(e.rules)
}

// RuleNames returns the names of all registered rules in priority order (highest first).
func (e *Engine) RuleNames() []string {
	sorted := sortRules(e.rules)
	names := make([]string, len(sorted))
	for i, r := range sorted {
		names[i] = r.Name
	}
	return names
}

// EvaluateAll evaluates all registered rules against facts and returns a Result
// for every rule, regardless of whether it matched.
// Rules are evaluated in descending Priority order.
//
// Example:
//
//	results, err := e.EvaluateAll(ctx, ruler.FactMap{"age": 25, "plan": "premium"})
func (e *Engine) EvaluateAll(ctx context.Context, facts FactMap) ([]Result, error) {
	sorted := sortRules(e.rules)
	results := make([]Result, 0, len(sorted))
	for _, r := range sorted {
		res, err := evalRule(ctx, r, facts)
		if err != nil {
			return results, err
		}
		results = append(results, res)
	}
	return results, nil
}

// EvaluateMatching evaluates all rules and returns only those that matched.
// Rules are returned in descending Priority order.
//
// Example:
//
//	matched, err := e.EvaluateMatching(ctx, ruler.FactMap{"age": 25})
func (e *Engine) EvaluateMatching(ctx context.Context, facts FactMap) ([]Result, error) {
	all, err := e.EvaluateAll(ctx, facts)
	if err != nil {
		return nil, err
	}
	var matched []Result
	for _, r := range all {
		if r.Matched {
			matched = append(matched, r)
		}
	}
	return matched, nil
}

// EvaluateFirst evaluates rules in priority order and returns the first match.
// Returns (Result{}, false, nil) if no rule matched.
//
// Example:
//
//	result, matched, err := e.EvaluateFirst(ctx, ruler.FactMap{"plan": "free"})
func (e *Engine) EvaluateFirst(ctx context.Context, facts FactMap) (Result, bool, error) {
	sorted := sortRules(e.rules)
	for _, r := range sorted {
		res, err := evalRule(ctx, r, facts)
		if err != nil {
			return Result{}, false, err
		}
		if res.Matched {
			return res, true, nil
		}
	}
	return Result{}, false, nil
}

// TotalScore sums the Score of all matched rules for a given FactMap.
//
// Example:
//
//	score, err := e.TotalScore(ctx, ruler.FactMap{"risk_flag": true, "new_account": true})
func (e *Engine) TotalScore(ctx context.Context, facts FactMap) (float64, error) {
	matched, err := e.EvaluateMatching(ctx, facts)
	if err != nil {
		return 0, err
	}
	var total float64
	for _, r := range matched {
		total += r.Score
	}
	return total, nil
}

// evalRule is the internal per-rule evaluator.
func evalRule(ctx context.Context, r Rule, facts FactMap) (Result, error) {
	matchedConds, ok, err := r.evaluate(ctx, facts)
	if err != nil {
		return Result{}, err
	}
	res := Result{
		RuleName:          r.Name,
		RuleDescription:   r.Description,
		Matched:           ok,
		Score:             0,
		MatchedConditions: matchedConds,
		Tags:              r.Tags,
		Metadata:          r.Metadata,
	}
	if ok {
		res.Score = r.Score
	}
	return res, nil
}
