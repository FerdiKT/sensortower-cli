# sensortower-cli

Use this skill when working on the `sensortower` CLI in this repository.

## Purpose

`sensortower` is a Go + Cobra CLI for read-only Sensor Tower iOS market data. It currently wraps three endpoint families:

- `publishers apps`
- `apps get`
- `charts category-rankings`

The design is intentionally small:
- explicit typed endpoint wrappers
- `table|json` output
- config/env support for `base_url`, `timeout_seconds`, `cookie`, and `headers`
- GitHub release artifacts and Homebrew tap distribution

## Repository layout

- `cmd/`: Cobra command tree
- `internal/config/`: config file + env override handling
- `internal/sensortower/`: HTTP client and response types
- `internal/output/`: table and JSON rendering
- `testdata/`: captured fixture payloads for decoding tests
- `.github/workflows/release.yml`: tag-based release workflow
- `assets/hero-banner.svg`: README header art

## Preferred workflow

1. Read the touched command and the corresponding client method together.
2. If changing response handling, update tests and fixture assumptions at the same time.
3. Run `gofmt -w $(find . -name '*.go' -type f)`.
4. Run `go test ./...`.
5. If CLI behavior changed, run one or more smoke checks:
   - `go run . publishers apps --publisher-id 1619264551 --output json`
   - `go run . apps get --app-id 6478631467 --country US --output json`
   - `go run . charts category-rankings --country US --category 0 --date 2026-04-16 --device iphone --output json`

## Output rules

- Prefer preserving upstream JSON fields instead of remapping them heavily.
- Keep table output compact and operationally useful.
- Do not add generic raw API passthrough commands unless explicitly requested.

## Release rules

- Build local archives with `make brew-dist VERSION=<version>`.
- GitHub releases are created by pushing tags like `v0.1.1`.
- Homebrew formula lives in `FerdiKT/homebrew-tap`.
- If the release asset names or checksums change, update the tap formula accordingly.

## Current distribution state

- Repo: `https://github.com/FerdiKT/sensortower-cli`
- Tap formula: `sensortower`
- Install path:
  - `brew tap FerdiKT/tap`
  - `brew install sensortower`
