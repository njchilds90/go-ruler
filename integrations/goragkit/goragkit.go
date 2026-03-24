package goragkit

import (
	"context"

	ruler "github.com/njchilds90/go-ruler"
)

// RAGEnvelope is a neutral envelope for retrieval and generation metadata.
type RAGEnvelope struct {
	Query      string         `json:"query"`
	TopK       int            `json:"top_k"`
	SourceIDs  []string       `json:"source_ids,omitempty"`
	Model      string         `json:"model,omitempty"`
	Confidence float64        `json:"confidence,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// ToFacts converts an envelope to a rule engine FactMap for policy evaluation.
func ToFacts(env RAGEnvelope) ruler.FactMap {
	return ruler.FactMap{
		"query":      env.Query,
		"top_k":      env.TopK,
		"source_ids": env.SourceIDs,
		"model":      env.Model,
		"confidence": env.Confidence,
	}
}

// EvaluateAgentPolicy evaluates a goragkit envelope with go-ruler.
func EvaluateAgentPolicy(ctx context.Context, engine *ruler.Engine, env RAGEnvelope) (ruler.AgentDecision, error) {
	return engine.EvaluateDecision(ctx, ToFacts(env))
}
