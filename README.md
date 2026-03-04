# go-ruler

[![CI](https://github.com/njchilds90/go-ruler/actions/workflows/ci.yml/badge.svg)](https://github.com/njchilds90/go-ruler/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/njchilds90/go-ruler.svg)](https://pkg.go.dev/github.com/njchilds90/go-ruler)
[![Go Report Card](https://goreportcard.com/badge/github.com/njchilds90/go-ruler)](https://goreportcard.com/report/github.com/njchilds90/go-ruler)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A declarative, zero-dependency rule engine for Go.

go-ruler lets you define named rules composed of typed conditions, evaluate them against arbitrary fact maps, and receive structured, explainable results. It is designed for:

- **Human developers** building feature flags, access control, pricing logic, and fraud detection
- **AI agents** that need deterministic, inspectable, machine-readable policy evaluation

## Features

- Zero external dependencies
- 13 built-in condition operators: `eq`, `neq`, `gt`, `gte`, `lt`, `lte`, `contains`, `not_contains`, `in`, `not_in`, `matches`, `exists`, `not_exists`
- `AND` / `OR` logical operators per rule
- Priority-ordered evaluation
- Score accumulation across matched rules
- Structured, JSON-serializable `Result` type
- Sentinel errors for programmatic handling
- `context.Context` support throughout
- Fully table-driven test suite

## Install
```bash
go get github.com/njchilds90/go-ruler
```

## Quick Start
```go
package main

import (
    "context"
    "fmt"
    "github.com/njchilds90/go-ruler"
)

func main() {
    e := ruler.NewEngine()

    e.MustAddRule(ruler.Rule{
        Name:        "premium-adult-user",
        Description: "Active premium users over 18",
        Priority:    10,
        Score:       100,
        Op:          ruler.OpAnd,
        Conditions: []ruler.Condition{
            ruler.GreaterThanEquals("age", 18),
            ruler.Equals("plan", "premium"),
            ruler.Equals("status", "active"),
        },
    })

    e.MustAddRule(ruler.Rule{
        Name:     "any-admin",
        Priority: 20,
        Score:    200,
        Op:       ruler.OpOr,
        Conditions: []ruler.Condition{
            ruler.Equals("role", "admin"),
            ruler.In("role", []any{"superuser", "root"}),
        },
    })

    facts := ruler.FactMap{
        "age":    25,
        "plan":   "premium",
        "status": "active",
        "role":   "user",
    }

    matched, err := e.EvaluateMatching(context.Background(), facts)
    if err != nil {
        panic(err)
    }

    for _, r := range matched {
        fmt.Printf("\u2705 %s (score: %.0f)\n", r.RuleName, r.Score)
        fmt.Printf("  conditions: %v\n", r.MatchedConditions)
    }
}
```

Output:
```
\u2705 premium-adult-user (score: 100)
  conditions: [age gte 18 plan eq premium status eq active]
```

## Evaluation Modes

| Method | Returns |
|---|---|
| `EvaluateAll` | Every rule result (matched + unmatched) |
| `EvaluateMatching` | Only matched rule results |
| `EvaluateFirst` | First matched result by priority |
| `TotalScore` | Sum of scores for all matched rules |

## Condition Constructors
```go
ruler.Equals("field", value)
ruler.NotEquals("field", value)
ruler.GreaterThan("field", 18)
ruler.GreaterThanEquals("field", 18)
ruler.LessThan("field", 100)
ruler.LessThanEquals("field", 100)
ruler.Contains("bio", "gopher")          // string or []any
ruler.NotContains("bio", "spam")
ruler.In("role", []any{"admin", "mod"})
ruler.NotIn("role", []any{"banned"})
ruler.Matches("email", `^.+@.+..+$`)   // regex
ruler.Exists("optional_field")
ruler.NotExists("deleted_at")
```

## Risk Scoring Example
```go
e := ruler.NewEngine()

e.MustAddRule(ruler.Rule{
    Name: "risk:new-account", Score: 30,
    Op: ruler.OpAnd,
    Conditions: []ruler.Condition{ruler.Equals("new_account", true)},
})
e.MustAddRule(ruler.Rule{
    Name: "risk:foreign-ip", Score: 50,
    Op: ruler.OpAnd,
    Conditions: []ruler.Condition{ruler.Equals("foreign_ip", true)},
})
e.MustAddRule(ruler.Rule{
    Name: "risk:invalid-email", Score: 40,
    Op: ruler.OpAnd,
    Conditions: []ruler.Condition{
        ruler.NotMatches("email", `^[a-z0-9._%+-]+@[a-z0-9.\-]+\.[a-z]{2,}$`),
    },
})

score, _ := e.TotalScore(ctx, ruler.FactMap{
    "new_account": true,
    "foreign_ip":  true,
    "email":       "legit@example.com",
})
// score == 80
```

## Error Handling

goruler uses sentinel errors wrappable with `errors.Is`:
```go
var (
    ruler.ErrInvalidRule
    ruler.ErrInvalidCondition
    ruler.ErrDuplicateRule
    ruler.ErrTypeMismatch
    ruler.ErrContextCanceled
)
```

## AI Agent Use Case

goruler is ideal for agent decision pipelines:
```go
type AgentDecision struct {
    Action string
    Score  float64
    Reason []string
}

func evaluate(ctx context.Context, e *ruler.Engine, facts ruler.FactMap) AgentDecision {
    result, matched, _ := e.EvaluateFirst(ctx, facts)
    if !matched {
        return AgentDecision{Action: "default"}
    }
    return AgentDecision{
        Action: result.RuleName,
        Score:  result.Score,
        Reason: result.MatchedConditions,
    }
}
```

## License

MIT — see [LICENSE](LICENSE)
