# HumbleBee GUI (Wails v2 + React)

This is the standalone GUI application. It uses the same local SQLite database format as the CLI (`~/.humblebee/humblebee.db` by default), but it does not require the CLI to be installed.

## Requirements

- Go 1.22+
- Node.js 18+
- Wails v2

Install Wails:
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
```

## Run (development)

From the repo root:
```bash
cd frontend
npm install
cd ..
wails dev
```

Reason: Wails expects the GUI entrypoint in the current working directory. The repo root now provides that entrypoint, while the CLI binary remains under `cmd/humblebee`.

## Build

```bash
cd frontend
npm install
npm run build
cd ..
wails build -tags production
```

## Release Builds

GUI release assets are built by `.github/workflows/release-gui.yml` after a GitHub release is published. The CLI release remains GoReleaser-based; the GUI workflow attaches standalone Wails app downloads to the same `v*` release.

Release asset names include both the GUI marker and release tag, for example `HumbleBee_GUI_v0.2.1_darwin_arm64.zip`.

The CLI command `humblebee gui` launches an installed GUI app if one is available next to the CLI, on `PATH`, or via `HUMBLEBEE_GUI_PATH`.

## Notes

- The GUI uses the same DB as the CLI by default; you can override it for testing with `HUMBLEBEE_HOME`.
  - Example: `HUMBLEBEE_HOME="$PWD/.humblebee-test" wails dev`
- Backend bindings live in `internal/guiapp`.
- The Wails entrypoint is the repository root `main.go`.
- Production builds use embedded frontend assets through the `production` build tag.
