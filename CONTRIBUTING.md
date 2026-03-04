# Contributing to go-ruler

Thank you for your interest in contributing!

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/go-ruler`
3. Create a feature branch: `git checkout -b feat/my-feature`
4. Make your changes with tests
5. Run `go test -race ./...` and `go vet ./...`
6. Open a pull request against `main`

## Guidelines

- All new features must include table-driven tests
- All exported symbols must have GoDoc comments with at least one example
- Zero external runtime dependencies — this is a hard constraint
- Follow standard Go naming conventions
- Keep the public API minimal; prefer extending via new operators or options over breaking changes

## Reporting Bugs

Open a GitHub Issue with:
- Go version (`go version`)
- Minimal reproduction case
- Expected vs. actual behavior

## Semantic Versioning

- Patch (`v1.0.x`): bug fixes, no API changes
- Minor (`v1.x.0`): new operators, new methods, backward-compatible
- Major (`v2.0.0`): breaking API changes (rare, discussed in issues first)

## Release & Verification
RELEASE STEPS (GitHub Web UI only)
====================================

1. VERIFY CI IS GREEN
   - Go to: https://github.com/njchilds90/go-ruler/actions
   - Confirm the latest push shows all green checkmarks before tagging.

2. CREATE THE TAG
   - Go to: https://github.com/njchilds90/go-ruler/releases/new
   - In "Choose a tag" → type: v1.0.0 → click "Create new tag: v1.0.0 on publish"
   - Target branch: main

3. FILL IN RELEASE DETAILS
   Title:
     v1.0.0 — Initial Release

   Release notes:
   ---
   ## go-ruler v1.0.0

   A declarative, zero-dependency rule engine for Go.

   ### Highlights
   - 13 condition operators (eq, neq, gt, gte, lt, lte, contains, not_contains, in, not_in, matches, exists, not_exists)
   - AND / OR logical operators per rule
   - Four evaluation modes: EvaluateAll, EvaluateMatching, EvaluateFirst, TotalScore
   - Priority-ordered evaluation + score accumulation
   - Structured, JSON-serializable Result type
   - Sentinel errors for programmatic handling
   - context.Context support throughout
   - Zero external runtime dependencies
   - Full table-driven test suite, race-detector clean
   - CI: Go 1.21 / 1.22 / 1.23

   ### Install
   go get github.com/njchilds90/go-ruler@v1.0.0
   ---

4. CHECK "Set as the latest release"
5. Click "Publish release"

SEMANTIC VERSIONING GUIDANCE
=============================
- v1.0.x  → Bug fixes only, no API changes
- v1.x.0  → New condition operators, new Engine methods (backward-compatible)
- v2.0.0  → Breaking API changes (discuss in issues first)

VERIFICATION
=============
After ~10 minutes, visit:
https://pkg.go.dev/github.com/njchilds90/go-ruler

If it hasn't appeared yet, force indexing by visiting:
https://pkg.go.dev/github.com/njchilds90/go-ruler@v1.0.0

Or run from any machine with Go installed:
GONOSUMCHECK=* GOFLAGS=-mod=mod go get github.com/njchilds90/go-ruler@v1.0.0