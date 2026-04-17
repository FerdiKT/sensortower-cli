# SensorTowerCli Agent Notes

This repository contains a small Go CLI named `sensortower`.

Primary workflows:
- Run tests: `go test ./...`
- Local smoke checks:
  - `go run . publishers apps --publisher-id 1619264551 --output json`
  - `go run . apps get --app-id 6478631467 --country US --output json`
  - `go run . charts category-rankings --country US --category 0 --date 2026-04-16 --device iphone --output json`
- Build release binary: `make build VERSION=0.1.1`
- Build release archives: `make brew-dist VERSION=0.1.1`

Release and distribution:
- GitHub repo: `FerdiKT/sensortower-cli`
- Homebrew tap repo: `FerdiKT/homebrew-tap`
- Formula file: `Formula/sensortower.rb`

Important conventions:
- Keep the CLI read-only unless the project scope changes explicitly.
- Preserve the public endpoint wrappers:
  - `publishers apps`
  - `apps get`
  - `charts category-rankings`
- JSON output should stay close to upstream Sensor Tower payloads.
- Table output should remain compact and high-signal.
