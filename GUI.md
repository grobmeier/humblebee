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

### Signed macOS App

The GitHub workflow still builds an unsigned macOS GUI asset. To replace it with
a signed and notarized app without spending GitHub-hosted macOS runner minutes,
run the local macOS release helper from a Mac with the Apple Developer tools
configured:

```bash
xcrun notarytool store-credentials humblebee-notary
scripts/release-macos-app.sh v0.2.1
```

The script asks which git ref to build and defaults to `origin/main`. It builds
the Wails app in a temporary worktree, signs it with a local Developer ID
Application certificate, notarizes and staples it, then uploads the same release
asset name with `gh release upload --clobber`. Use `--no-upload` to test the
local build/sign/notarization flow without replacing the GitHub release asset.

When running `scripts/release.sh` on macOS, the release script offers to wait
for the unsigned GitHub GUI asset and replace it with the local signed and
notarized build.

The CLI command `humblebee gui` launches an installed GUI app if one is available next to the CLI, on `PATH`, or via `HUMBLEBEE_GUI_PATH`.

## Notes

- The GUI uses the same DB as the CLI by default; you can override it for testing with `HUMBLEBEE_HOME`.
  - Example: `HUMBLEBEE_HOME="$PWD/.humblebee-test" wails dev`
- Backend bindings live in `internal/guiapp`.
- The Wails entrypoint is the repository root `main.go`.
- Production builds use embedded frontend assets through the `production` build tag.
