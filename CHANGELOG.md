# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2026-03-24

### Added
- High-level fluent builder API via `NewRule(...).Condition(...).Build()`.
- `AgentDecision` and `WhatIfResult` outputs for AI-agent-friendly, deterministic policy traces.
- In-memory deterministic cache option (`WithCache`) and engine freezing (`Freeze`).
- JSON rule file lifecycle helpers: `LoadRulesFile` and `SaveRulesFile`.
- Advanced operators: `starts_with`, `ends_with`, and `between`.
- Optional subpackages:
  - `guardrails` deny/allow helper.
  - `integrations/goragkit` envelope-to-facts adapter.
  - `otelruler` tracing bridge interface for OpenTelemetry adapters.
- New CLI binary (`cmd/go-ruler`) with `eval`, `load`, and `serve` commands.
- Benchmarks and integration-style tests for what-if and loader flows.

### Changed
- CI workflow now includes `golangci-lint`, race tests, and atomic coverage reporting.
- Evaluation ordering is now deterministic by `(priority desc, rule name asc)`.

## [1.0.0] - 2026-02-25

### Added
- `Engine` type with `AddRule`, `MustAddRule`, `RuleCount`, `RuleNames`
- Four evaluation modes: `EvaluateAll`, `EvaluateMatching`, `EvaluateFirst`, `TotalScore`
- `Rule` type with `Name`, `Description`, `Priority`, `Score`, `Op`, `Conditions`, `Tags`, `Metadata`
- `Condition` type and 13 constructor helpers: `Equals`, `NotEquals`, `GreaterThan`, `GreaterThanEquals`, `LessThan`, `LessThanEquals`, `Contains`, `NotContains`, `In`, `NotIn`, `Matches`, `Exists`, `NotExists`
- `AND` / `OR` logical operators per rule
- Structured `Result` type with JSON tags — machine-readable by design
- Sentinel error types: `ErrInvalidRule`, `ErrInvalidCondition`, `ErrDuplicateRule`, `ErrTypeMismatch`, `ErrContextCanceled`
- `context.Context` support on all evaluation methods
- Full table-driven test suite with race detector support
- GitHub Actions CI across Go 1.21, 1.22, 1.23
- GoDoc examples on all exported functions
- Zero external runtime dependencies
