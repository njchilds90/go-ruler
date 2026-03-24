package otelruler

import (
	"context"
	"time"

	ruler "github.com/njchilds90/go-ruler"
)

// Span is a minimal tracing abstraction allowing OpenTelemetry adapters without adding dependencies.
type Span interface {
	SetAttribute(key string, value any)
	RecordError(err error)
	End()
}

// Tracer is a minimal tracing abstraction allowing OpenTelemetry adapters without adding dependencies.
type Tracer interface {
	Start(ctx context.Context, name string) (context.Context, Span)
}

// EvaluateDecisionWithTrace wraps ruler decision evaluation with optional tracing hooks.
func EvaluateDecisionWithTrace(ctx context.Context, tracer Tracer, engine *ruler.Engine, facts ruler.FactMap) (ruler.AgentDecision, error) {
	if tracer == nil {
		return engine.EvaluateDecision(ctx, facts)
	}
	ctx, span := tracer.Start(ctx, "ruler.evaluate_decision")
	span.SetAttribute("ruler.facts.count", len(facts))
	started := time.Now()
	decision, err := engine.EvaluateDecision(ctx, facts)
	span.SetAttribute("ruler.duration_ms", float64(time.Since(started).Microseconds())/1000.0)
	span.SetAttribute("ruler.action", decision.Action)
	span.SetAttribute("ruler.allowed", decision.Allowed)
	if err != nil {
		span.RecordError(err)
	}
	span.End()
	return decision, err
}
