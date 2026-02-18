# Cloudflare Integration Plan

## Summary

This document records the roadmap to bring this repository up to the standards used by Cloudflare (example: `matched-data-cli`). It is intended to be picked up later and worked incrementally.

## Goals

- Make CI and quality gates match Cloudflare expectations (`golangci-lint`, `go test`, coverage, `go vet`).
- Reach a practical test coverage baseline: 70% (80% stretch).
- Ensure non-interactive core is well-tested and stable.
- Provide reproducible releases (binaries) via `goreleaser` and CI.
- Improve developer and contributor docs (README, CONTRIBUTING, tests/run guide).

## Targets

- Coverage: 70% target, 80% stretch
- Linters: `golangci-lint` run clean, `staticcheck` addressed
- CI: GitHub Actions running tests, linters, coverage upload, and release
- Releases: `goreleaser` config + release workflow

## High-level Steps (prioritized)

1. Gap analysis (deliverable: checklist comparing this repo to `matched-data-cli`).
2. Align CI: add GitHub Actions to run `golangci-lint`, `go test -coverprofile`, `go vet`, and upload coverage.
3. Fix linter/staticcheck issues (fix root causes, not just silencing).
4. Increase test coverage to 70%: focus on `pkg/utils`, `internal/ui` rendering and model logic, and CLI flows.
5. Refactor TUI to separate rendering from interactive I/O for deterministic testing.
6. Add `goreleaser` and release workflow; add signing if desired.
7. Improve docs: README examples, CONTRIBUTING, RUNNING_TESTS, RELEASE notes, changelog.
8. Run security/dependency checks (`gosec`, `go list -m -u all`) and remediate.
9. Developer tooling: Makefile targets, pre-commit hooks, dev setup instructions.
10. Finalize PR to Cloudflare with checklist and CI green.

## Short-term next step

- Start with the gap analysis and CI alignment (1–2 days). This yields an exact list of changes required by Cloudflare and provides immediate CI coverage.

## Notes

- TUI interactive behavior should be tested by exercising model methods and view outputs rather than attempting headful UI tests in CI.
- Prefer temporary fixtures (t.TempDir) and table-driven tests for repeatability.

Saved: This plan was generated and saved to allow resuming work later.
