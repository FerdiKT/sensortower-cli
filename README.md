<p align="center">
  <img src="assets/hero-banner.svg" alt="sensortower CLI" width="820" />
</p>

<h1 align="center">sensortower</h1>

<p align="center">
  <strong>A JSON-first CLI for Sensor Tower iOS market data</strong><br />
  Read-only · Scriptable · CI-friendly · Built for market intelligence workflows
</p>

<p align="center">
  <a href="#-status"><img src="https://img.shields.io/badge/status-beta-yellow?style=flat-square" alt="Beta" /></a>
  <a href="#-quickstart"><img src="https://img.shields.io/badge/quickstart-3_commands-brightgreen?style=flat-square" alt="Quickstart" /></a>
  <a href="#-installation"><img src="https://img.shields.io/badge/go_install-ready-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go Install" /></a>
  <a href="#-configuration"><img src="https://img.shields.io/badge/config-cookie_header_ready-blue?style=flat-square" alt="Config" /></a>
  <a href="#-license"><img src="https://img.shields.io/badge/license-MIT-lightgrey?style=flat-square" alt="License" /></a>
</p>

---

## ⚠️ Status

> **This project is in early public beta.** The current release is intentionally small and read-only, focused on a few high-value Sensor Tower iOS endpoints. The HTTP/config layer already supports optional cookie/header injection for future session-backed usage, but no login automation is included in v1.

---

## ✨ Why sensortower?

> Pull Sensor Tower data from your terminal without rebuilding ad-hoc scripts every time.

| | |
|---|---|
| 📊 **Core market workflows** | Publisher apps, app details, rankings, competitors, and ASO helpers |
| 🧰 **JSON-first output** | Pipe to `jq`, store snapshots, build quick analyses |
| 🪶 **Small surface area** | Explicit commands for known endpoints, no generic wrapper noise |
| 🔐 **Session-ready design** | Optional cookie + header injection via config/env |
| 🚀 **Automation-friendly** | Retry, cache, contexts, batch reads, export files, Homebrew workflow |

---

## 📦 Installation

<details open>
<summary><strong>Option 1 — Homebrew</strong> (recommended)</summary>

```bash
brew tap FerdiKT/tap
brew install sensortower
```

</details>

<details>
<summary><strong>Option 2 — Go install</strong></summary>

```bash
go install github.com/ferdikt/sensortower-cli@latest
```

</details>

<details>
<summary><strong>Option 3 — Build from source</strong></summary>

```bash
git clone https://github.com/FerdiKT/sensortower-cli.git
cd sensortower-cli
make tidy
make build VERSION=0.1.0
./bin/sensortower version
```

</details>

---

## 🚀 Quickstart

### 1️⃣ List a publisher's apps

```bash
sensortower publishers apps \
  --publisher-id 1619264551 \
  --limit 25 \
  --offset 0 \
  --sort-by downloads
```

### 2️⃣ Fetch a single app

```bash
sensortower apps get \
  --app-id 6478631467 \
  --country US
```

### 3️⃣ Fetch iOS category rankings

```bash
sensortower charts category-rankings \
  --country US \
  --category 0 \
  --date 2026-04-16 \
  --device iphone \
  --limit 25
```

### JSON mode

```bash
sensortower charts category-rankings \
  --country US \
  --category 0 \
  --date 2026-04-16 \
  --device iphone \
  --output json | jq '.data.free[0]'
```

### Find fresh earners (default: last 1 month, >= $10k)

```bash
sensortower workflow fresh-earners --output json
```

Custom window/threshold:

```bash
sensortower workflow fresh-earners \
  --months 2 \
  --min-revenue-usd 25000 \
  --categories 0 \
  --country US \
  --output json
```

---

## 🗺️ Command Map

| Group | Commands | Purpose |
|---|---|---|
| `search` | `apps` · `publishers` | Search app and publisher autocomplete endpoints |
| `publishers` | `apps` | List apps for a publisher |
| `apps` | `get` | Fetch a single iOS app detail payload |
| `charts` | `category-rankings` | Fetch free/grossing/paid rankings |
| `contexts` | `add` · `list` · `use` | Manage named configs for multiple setups |
| `workflow` | `competitors` · `fresh-earners` | Pull rankings, dedupe competitors, enrich app metadata, and find recently released high-earning apps |
| `aso` | `metadata-audit` · `keyword-gap` | Generate ASO-oriented diagnostics |
| `agent` | `install-skill` · `link-skill` · `show-skill-path` | Install or link the bundled Codex skill |
| `version` | — | Print binary version |

---

## ⚙️ Configuration

Default config path:

```text
~/Library/Application Support/sensortower/config.json
```

Example:

```json
{
  "base_url": "https://app.sensortower.com",
  "timeout_seconds": 30,
  "output": "table",
  "cookie": "sensor_tower_session=...",
  "headers": {
    "X-Custom-Header": "value"
  }
}
```

Environment overrides:

```bash
export SENSORTOWER_BASE_URL="https://app.sensortower.com"
export SENSORTOWER_TIMEOUT_SECONDS=30
export SENSORTOWER_OUTPUT=json
export SENSORTOWER_COOKIE="sensor_tower_session=..."
export SENSORTOWER_HEADERS_JSON='{"X-Custom-Header":"value"}'
```

Current global flags:

```bash
sensortower --config /path/to/config.json --output json ...
```

Useful automation flags:

```bash
sensortower --context team-a \
  --retry-429 --retry-max 8 --retry-wait 60 \
  --cache-ttl 300 \
  --output-format jsonl \
  --output-file ./out.jsonl ...
```

---

## 🔎 Batch, Workflow, ASO

Autocomplete search:

```bash
sensortower search apps --term ferdi --output json
sensortower search publishers --term ferdi --output json
```

Batch app metadata:

```bash
sensortower apps get \
  --app-ids-file ids.txt \
  --country US \
  --fields name,subtitle,description.full_description \
  --output-format jsonl \
  --output-file ./apps.jsonl
```

Competitor workflow:

```bash
sensortower workflow competitors \
  --country US \
  --categories 7018,7019 \
  --top 200 \
  --output-format json \
  --output-file ./competitors.json
```

ASO helpers:

```bash
sensortower aso metadata-audit --app-id 6478631467 --country US --output json
sensortower aso keyword-gap --app-id 6478631467 --competitor-ids-file competitor_ids.txt --country US --output json
```

Contexts:

```bash
sensortower contexts add --name team-a --cookie 'sensor_tower_session=...' --headers-json '{"X-Custom":"1"}'
sensortower contexts use --name team-a
sensortower contexts list --output json
```

---

## 🤖 Agent Workflow

Repo-local guidance lives in [`AGENTS.md`](.agents/AGENTS.md) and the bundled skill at [`skills/sensortower-cli/SKILL.md`](skills/sensortower-cli/SKILL.md).

To install the bundled Codex skill directly from the CLI:

```bash
sensortower agent install-skill
```

For local development, if you want a symlink instead of a copied install:

```bash
sensortower agent link-skill --source ./skills/sensortower-cli
```

To inspect the target install path:

```bash
sensortower agent show-skill-path
```

---

## 🧪 Development

```bash
make tidy
make test
make build VERSION=0.1.0
make dist VERSION=0.1.0
```

Smoke checks:

```bash
go run . publishers apps --publisher-id 1619264551 --output json
go run . apps get --app-id 6478631467 --country US --output json
go run . charts category-rankings --country US --category 0 --date 2026-04-16 --device iphone --output json
```

---

## 📄 License

MIT
