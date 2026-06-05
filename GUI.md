# HumbleBee GUI (Wails v2 + React)

This is an early GUI prototype that uses the same local SQLite database as the CLI (`~/.humblebee/humblebee.db`).

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
wails build
```

## Release Builds

GUI release assets are built by `.github/workflows/release-gui.yml` after a GitHub release is published. The CLI release remains GoReleaser-based; the GUI workflow attaches Wails app downloads to the same `v*` release.

The CLI command `humblebee gui` launches an installed GUI app if one is available next to the CLI, on `PATH`, or via `HUMBLEBEE_GUI_PATH`.

## Notes

- The GUI uses the same DB as the CLI by default; you can override it for testing with `HUMBLEBEE_HOME`.
  - Example: `HUMBLEBEE_HOME="$PWD/.humblebee-test" wails dev`
- Backend bindings live in `internal/guiapp`.
- Entry point is `cmd/humblebee-gui`.
