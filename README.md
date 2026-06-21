# HumbleBee (MVP)

Local-first CLI time tracking that stays out of your way.

## Requirements

- Go 1.22+
- SQLite is embedded via `modernc.org/sqlite` (no CGO)

## Install (recommended)

### Homebrew (macOS/Linux)

```bash
brew tap grobmeier/tap
brew install humblebee
```

### Scoop (Windows)

```powershell
scoop bucket add grobmeier https://github.com/grobmeier/scoop-bucket
scoop install humblebee
```

## Install / Build

```bash
git clone https://github.com/grobmeier/humblebee
cd humblebee
go mod tidy
go build -o bin/humblebee ./cmd/humblebee
```

## Data storage

- Database path: `~/.humblebee/humblebee.db`
- Override base directory with `HUMBLEBEE_HOME` (database becomes `$HUMBLEBEE_HOME/humblebee.db`)

## Usage

```bash
humblebee help
```

## GUI

HumbleBee also ships a standalone local-first GUI built with Wails and React. The GUI uses the same local database format as the CLI, but it does not require the CLI to be installed.

Download the UI-first app from the GitHub release assets for your operating system. GUI assets are named with the release version, for example `HumbleBee_GUI_v0.2.1_darwin_arm64.zip`.

If the GUI app or `humblebee-gui` binary is installed next to the CLI, on your `PATH`, or configured through `HUMBLEBEE_GUI_PATH`, launch it from the CLI:

```bash
humblebee gui
```

For local GUI development, run Wails from the repository root:

```bash
cd frontend
npm install
cd ..
wails dev
```

Use `humblebee gui` to test the launcher behavior after a GUI build or installed release is available. See `GUI.md` for the full GUI development notes.

## Releases

HumbleBee uses GitHub Actions for release builds.

- CLI-first releases are created from `v*` tags through GoReleaser and publish the `humblebee` CLI archives, Homebrew formula, and Scoop manifest.
- UI-first release assets are built after the GitHub release is published and attached to the same release as standalone Wails app downloads.
- SBOM assets are generated and attached to published releases as CycloneDX JSON, for example `HumbleBee_v0.2.1_sbom.cdx.json`.

### Signed macOS GUI release

The GitHub GUI workflow keeps producing the macOS app asset, but it is not
signed or notarized there so GitHub-hosted macOS runner minutes stay limited.
For a trusted macOS download, build and replace the macOS release asset locally
from a Mac that has a Developer ID Application certificate installed:

```bash
xcrun notarytool store-credentials humblebee-notary
scripts/release-macos-app.sh v0.2.1 --no-upload
scripts/release-macos-app.sh v0.2.1
```

The first command stores Apple notarization credentials in the local keychain.
The `--no-upload` run verifies the local build/sign/notarization flow. The final
run uploads the signed and notarized app to the existing GitHub release using
the same asset name, for example `HumbleBee_GUI_v0.2.1_darwin_arm64.zip`.

See `GUI.md` for the detailed GUI release notes.

### Doctor (health check / safe repair)

`humblebee doctor` helps diagnose common issues (DB location, initialization, schema, running timer, and timezone metadata on entries).

Read-only check:
```bash
humblebee doctor
```

Safe repairs (requires no running timer):
```bash
humblebee doctor --fix
```

Notes:
- `--fix` re-runs idempotent migrations and backfills timezone fields for older (stopped) entries that predate timezone tracking.
- Use `--dry-run` to see what would change without writing:
  - `humblebee doctor --fix --dry-run`
- If you want to backfill using a specific timezone name (IANA), pass `--tz-name`:
  - `humblebee doctor --fix --tz-name America/New_York`

### Initialize

```bash
humblebee init
```

Creates:
- a default user (single-user MVP; stored in `persons`)
- a `Default` work item (timers without an explicit work item store `workitem_id = NULL`)

Non-interactive:

```bash
humblebee init --email user@example.com --workitem "Client Project A"
```

### Work items

```bash
humblebee add "Client Project A"
humblebee add "Client Project A > Feature Development"
humblebee show
humblebee remove "Client Project A"
```

Notes:
- Names are case-insensitive for lookup and uniqueness.
- `remove` archives the selected work item and its entire subtree.

### Time tracking

```bash
humblebee start "Client Project A > Feature Development"
humblebee start            # starts "Default"
humblebee stop
```

Delete time entries (interactive, hard delete):

```bash
humblebee delete
```

Notes:
- Only one running timer is allowed at a time (enforced in code and by a partial UNIQUE index).
- Cross-midnight time is split at local midnight for daily totals and reports (timestamps are stored as UTC Unix seconds).

### Reports

```bash
humblebee report          # current month
humblebee report 5        # month (1-12) in current year
humblebee report 5 2025   # month + year
```

## Color output

- Success = green, warnings = yellow, errors = red
- Disable colors with `--no-color` or by setting `NO_COLOR` to any non-empty value

## Development

Run tests:

```bash
go test ./...
```

## Security

Please report suspected vulnerabilities privately. See [SECURITY.md](SECURITY.md) for supported versions, reporting instructions, and SBOM maintenance notes.
