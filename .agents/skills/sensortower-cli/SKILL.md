# sensortower-cli

Use this skill when editing or releasing the `sensortower` CLI.

## Scope

This project is a read-only Sensor Tower iOS CLI built with Go and Cobra. The supported command groups are:

- `publishers apps`
- `apps get`
- `charts category-rankings`

## Key files

- `cmd/`
- `internal/config/`
- `internal/sensortower/`
- `internal/output/`
- `testdata/`
- `.github/workflows/release.yml`
- `README.md`

## Validation

Run these after changes:

```bash
go test ./...
go run . publishers apps --publisher-id 1619264551 --output json
go run . apps get --app-id 6478631467 --country US --output json
go run . charts category-rankings --country US --category 0 --date 2026-04-16 --device iphone --output json
```

## Release notes

- Tagging `vX.Y.Z` triggers the release workflow.
- Homebrew artifacts are expected to match the filenames produced by `make brew-dist VERSION=X.Y.Z`.
- The Homebrew formula for this project lives in the separate tap repo `FerdiKT/homebrew-tap`.
