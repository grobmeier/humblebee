# HumbleBee (MVP)

Local-first CLI time tracking that stays out of your way.

## Requirements

- Go 1.21+
- SQLite is embedded via `modernc.org/sqlite` (no CGO)

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
