# Contribute to HumbleBee

HumbleBee is local-first time tracking for developers and solo users. Contributions are welcome when they keep the project small, understandable, and useful without turning it into a hosted project management suite.

## Good First Contributions

Useful contributions include:

- clear bug reports with reproduction steps,
- documentation improvements,
- small UI fixes,
- tests for repository, service, or GUI app behavior,
- import edge cases for Time & Bill exports,
- release and packaging fixes,
- accessibility improvements in the GUI.

Large feature ideas are best discussed in an issue before implementation.

## Development Setup

Requirements:

- Go 1.22 or newer,
- Node.js 20 or newer for the frontend,
- Wails v2 for GUI development.

Install frontend dependencies:

```bash
cd frontend
npm install
cd ..
```

Run the Go test suite:

```bash
go test ./...
```

Build the frontend:

```bash
npm --prefix frontend run build
```

Run the GUI during development:

```bash
go run github.com/wailsapp/wails/v2/cmd/wails@v2.13.0 dev
```

Build the production GUI with embedded frontend assets:

```bash
go run github.com/wailsapp/wails/v2/cmd/wails@v2.13.0 build -clean -tags production
```

## Local Data

By default, HumbleBee writes its database to:

```text
~/.humblebee/humblebee.db
```

For development and testing, use a separate local home directory:

```bash
HUMBLEBEE_HOME="$PWD/.humblebee-test" go test ./...
```

Do not commit local `.db` or `.sqlite` files.

## Pull Requests

Before opening a pull request:

```bash
go test ./...
npm --prefix frontend run build
git diff --check
```

Keep pull requests focused. A small bug fix, one UI improvement, or one documentation improvement is easier to review than a broad mixed change.

For user-facing behavior changes, include a short explanation of:

- what changed,
- how it was tested,
- whether existing local databases are affected.

## Project Boundaries

HumbleBee should stay:

- local-first,
- usable without a cloud account,
- understandable for solo users,
- scriptable for developers,
- careful with local time tracking data.

Features that require accounts, hosted collaboration, team permissions, or server-side billing workflows usually belong in Time & Bill instead of HumbleBee.

## Security

Do not open public issues for suspected vulnerabilities. Use the reporting process in `SECURITY.md`.

## Releases

Releases are built by GitHub Actions from `v*` tags. The release process publishes CLI archives, GUI assets, checksums, and a CycloneDX SBOM.
