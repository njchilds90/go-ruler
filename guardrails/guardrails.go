package guardrails

import (
	"context"

	ruler "github.com/njchilds90/go-ruler"
)

// Outcome represents the allow/block result from guardrail policy checks.
type Outcome struct {
	Allowed  bool                `json:"allowed"`
	Decision ruler.AgentDecision `json:"decision"`
}

// Evaluate executes a policy engine and blocks on deny:* actions.
func Evaluate(ctx context.Context, engine *ruler.Engine, facts ruler.FactMap) (Outcome, error) {
	decision, err := engine.EvaluateDecision(ctx, facts)
	if err != nil {
		return Outcome{}, err
	}
	allowed := decision.Allowed && len(decision.Action) >= 5 && decision.Action[:5] != "deny:"
	return Outcome{Allowed: allowed, Decision: decision}, nil
}
