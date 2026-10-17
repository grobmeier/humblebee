# HumbleBee GUI (Wails v2 + React)

This is the standalone GUI application. It uses the same local SQLite database format as the CLI (`~/.humblebee/humblebee.db` by default), but it does not require the CLI to be installed.

## Requirements

- Go 1.22+
- Node.js 18+
- Wails v2

Install Wails:
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0
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

## Browser UI Tests

The Playwright suite runs the real React UI against deterministic Wails binding
mocks. It covers browser-visible workflows without reading a personal database
or contacting Time & Bill.

```bash
cd frontend
npm install
npx playwright install chromium
npm run test:e2e
```

Use `npm run test:e2e:headed` to watch Chromium run the suite. Use
`npm run test:e2e:debug` to step through each action in the Playwright Inspector,
or `npm run test:e2e:ui` to select and rerun scenarios interactively. The suite verifies
manual-entry defaults, recent-note replacement, duplication, restored report
filters, database-backup feedback, and explicit news loading. Go tests remain
responsible for SQLite snapshots, preferences, and RSS parsing; native dialogs
and packaged Wails applications still need platform acceptance testing.

## Release Builds

The GUI workflow attaches standalone Wails downloads to the same release as the
GoReleaser CLI archives. Names include the GUI marker and tag, for example
`HumbleBee_GUI_v0.7.1_darwin_arm64.zip`. The GUI does not require a CLI installation.

Use [RELEASE.md](RELEASE.md) for release approval, tagging, local macOS signing
from the exact tag, automated website updates, asset verification and recovery.
The hosted macOS archive is not Developer ID signed/notarized; verify the local
signed replacement separately. Retrying website publishing must use its
standalone workflow so it does not overwrite that replacement.

The CLI command `humblebee gui` launches an installed GUI app if one is available
next to the CLI, on `PATH`, or via `HUMBLEBEE_GUI_PATH`.

## Notes

### Everyday Workflows

- The database filename next to the database-switch button identifies the active
  workspace. Hover for its full path, or open the dialog to select the path text.
- Report filters and the last successfully booked manual project/task are remembered
  separately for each database and profile. Moving a database starts fresh preferences.
- Expand **Recent notes** below the note field to reuse a previous note for the selected
  task. Replacing existing text requires confirmation; selecting a note never books time.
- The duplicate icon on a booked entry opens a new entry with its project, task, and
  note. Dates and times use normal new-entry defaults, and saving is still required.
- **Back up database** in the database dialog creates a consistent SQLite snapshot,
  including committed WAL data. Choose a new filename; existing files are not overwritten.
  Open the backup through the normal database-switch dialog. GUI preferences and the
  news cache are stored separately and are not included. A snapshot also preserves any
  running stopwatch state at backup time; it is not a continuously updated copy.
- The news button fetches HumbleBee news from Time & Bill only when opened or refreshed.
  There are no startup checks or background polling. The website receives ordinary
  connection information, including the IP address, but no workspace data or installation
  identifier. Previously fetched news remains available when offline. Article links open
  in the default browser.

- The GUI uses the same DB as the CLI by default; you can override it for testing with `HUMBLEBEE_HOME`.
  - Example: `HUMBLEBEE_HOME="$PWD/.humblebee-test" wails dev`
- Backend bindings live in `internal/guiapp`.
- The Wails entrypoint is the repository root `main.go`.
- Production builds use embedded frontend assets through the `production` build tag.
