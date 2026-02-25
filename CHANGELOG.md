# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
