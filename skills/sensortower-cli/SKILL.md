---
name: sensortower-cli
description: Use this skill when working with the local `sensortower` CLI for Sensor Tower iOS market data, category rankings, app details, publisher app lookups, JSON-first workflows, and repo-local release or Homebrew tap maintenance.
---

# SensorTower CLI

Use this skill for repository-local Sensor Tower CLI work.

## Workflow

1. Prefer public endpoint reads first.
2. Preserve upstream JSON fields unless a change is clearly needed.
3. Prefer JSON output for agent workflows.
4. Keep table output compact and operationally useful.
5. Validate with repo smoke checks after CLI changes.

## Read Pattern

- Use `sensortower publishers apps` for publisher-level app lookups.
- Use `sensortower apps get` for a full app detail payload.
- Use `sensortower charts category-rankings` for free, grossing, and paid rankings.
- Use `sensortower workflow fresh-earners` to find newly released apps above a revenue threshold (defaults: last 1 month and >= $10k).
- Add `--output json` for agent consumption.

## Config Pattern

- Default config is loaded from the user config directory.
- Override with `--config` when testing alternate setups.
- Use env vars like `SENSORTOWER_COOKIE` and `SENSORTOWER_HEADERS_JSON` when session-backed requests are needed.

## Release Pattern

- Run `go test ./...` before release work.
- Build archives with `make brew-dist VERSION=<version>`.
- Push tags like `v0.1.2` to trigger GitHub releases.
- Keep the Homebrew tap formula in sync with the latest release artifacts and checksums.
