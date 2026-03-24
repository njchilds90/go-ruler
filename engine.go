package ruler

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Evaluator is the interface implemented by Engine.
type Evaluator interface {
	EvaluateAll(ctx context.Context, facts FactMap) ([]Result, error)
	EvaluateDecision(ctx context.Context, facts FactMap) (AgentDecision, error)
}

// Option applies optional behavior to Engine.
type Option func(*Engine)

// WithCache enables deterministic in-memory caching of decision responses.
func WithCache() Option {
	return func(e *Engine) { e.cacheEnabled = true }
}

// Engine holds a registered set of Rules and evaluates them against FactMaps.
type Engine struct {
	mu           sync.RWMutex
	rules        []Rule
	frozen       bool
	cacheEnabled bool
	cache        map[[32]byte]AgentDecision
}

// NewEngine creates an empty Engine ready for rule registration.
func NewEngine(opts ...Option) *Engine {
	e := &Engine{cache: make(map[[32]byte]AgentDecision)}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// NewFromRules constructs an Engine from a rule list.
func NewFromRules(rules []Rule, opts ...Option) (*Engine, error) {
	e := NewEngine(opts...)
	for _, r := range rules {
		if err := e.AddRule(r); err != nil {
			return nil, err
		}
	}
	return e, nil
}

// Freeze prevents further rule registration and makes mutation-safe concurrent use explicit.
func (e *Engine) Freeze() { e.mu.Lock(); e.frozen = true; e.mu.Unlock() }

// AddRule registers a Rule with the Engine.
func (e *Engine) AddRule(r Rule) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.frozen {
		return ErrEngineFrozen
	}
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

// MustAddRule registers a Rule and panics on error.
func (e *Engine) MustAddRule(r Rule) {
	if err := e.AddRule(r); err != nil {
		panic(err)
	}
}

// RuleCount returns the number of rules registered in the Engine.
func (e *Engine) RuleCount() int { e.mu.RLock(); defer e.mu.RUnlock(); return len(e.rules) }

// RuleNames returns sorted names in evaluation order.
func (e *Engine) RuleNames() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	sorted := sortRules(e.rules)
	names := make([]string, len(sorted))
	for i, r := range sorted {
		names[i] = r.Name
	}
	return names
}

// EvaluateAll evaluates all registered rules against facts.
func (e *Engine) EvaluateAll(ctx context.Context, facts FactMap) ([]Result, error) {
	e.mu.RLock()
	rules := append([]Rule(nil), e.rules...)
	e.mu.RUnlock()
	sorted := sortRules(rules)
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
func (e *Engine) EvaluateFirst(ctx context.Context, facts FactMap) (Result, bool, error) {
	e.mu.RLock()
	rules := append([]Rule(nil), e.rules...)
	e.mu.RUnlock()
	for _, r := range sortRules(rules) {
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

// EvaluateDecision returns a deterministic, agent-friendly decision document.
func (e *Engine) EvaluateDecision(ctx context.Context, facts FactMap) (AgentDecision, error) {
	if e.cacheEnabled {
		if cached, ok := e.getCached(facts); ok {
			return cached, nil
		}
	}
	all, err := e.EvaluateAll(ctx, facts)
	if err != nil {
		return AgentDecision{}, err
	}
	matched := make([]Result, 0)
	var reasons []string
	var score float64
	action := "default"
	allowed := false
	for _, r := range all {
		if r.Matched {
			if action == "default" {
				action = r.RuleName
			}
			allowed = true
			score += r.Score
			reasons = append(reasons, r.MatchedConditions...)
			matched = append(matched, r)
		}
	}
	decision := AgentDecision{Action: action, Allowed: allowed, Score: score, Reasons: reasons, MatchedRules: matched, AllRules: all, EvaluatedAt: time.Now().UTC()}
	if e.cacheEnabled {
		e.putCached(facts, decision)
	}
	return decision, nil
}

// EvaluateWhatIf compares current facts against candidate facts.
func (e *Engine) EvaluateWhatIf(ctx context.Context, currentFacts, candidateFacts FactMap) (WhatIfResult, error) {
	cur, err := e.EvaluateDecision(ctx, currentFacts)
	if err != nil {
		return WhatIfResult{}, err
	}
	cand, err := e.EvaluateDecision(ctx, candidateFacts)
	if err != nil {
		return WhatIfResult{}, err
	}
	return WhatIfResult{Current: cur, Candidate: cand, Delta: cand.Score - cur.Score}, nil
}

func (e *Engine) getCached(facts FactMap) (AgentDecision, bool) {
	key, err := factKey(facts)
	if err != nil {
		return AgentDecision{}, false
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	d, ok := e.cache[key]
	return d, ok
}

func (e *Engine) putCached(facts FactMap, d AgentDecision) {
	key, err := factKey(facts)
	if err != nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.cache[key] = d
}

func factKey(f FactMap) ([32]byte, error) {
	b, err := json.Marshal(f)
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(b), nil
}

func evalRule(ctx context.Context, r Rule, facts FactMap) (Result, error) {
	matchedConds, ok, err := r.evaluate(ctx, facts)
	if err != nil {
		return Result{}, err
	}
	res := Result{RuleName: r.Name, RuleDescription: r.Description, Matched: ok, Score: 0, MatchedConditions: matchedConds, Tags: r.Tags, Metadata: r.Metadata}
	if ok {
		res.Score = r.Score
	}
	return res, nil
}
