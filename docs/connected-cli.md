# Time & Bill Connected CLI Pilot

This is an online-only pilot, not a change to HumbleBee's local-first default.
It requires Time & Bill connected-client contract version 1, including durable
mutation recovery and conditional stopwatches. A normal app token alone is not
enough: incompatible or older servers are rejected before data commands run.

The corresponding server hardening must be deployed and accepted before this
pilot ships. Source integration and HTTP fixture tests are not deployment or
authenticated web/CLI acceptance. GUI connected mode and offline sync are not
implemented.

## Connect

Create a scoped, expiring app token in your Time & Bill account. Use your actual
application origin, not the marketing website. The `.example` origin below is
a placeholder and must be replaced.

```sh
humblebee remote connect --name work --url https://your-timeandbill-origin.example
humblebee remote status --profile work
```

The interactive token prompt disables terminal echo. Never pass a token as an
argument, URL, note, issue attachment or log. For a pipe from a secret manager,
use `--token-stdin`; token input is bounded and no plaintext fallback is saved.

Credentials use the native secure store: macOS Keychain, Windows Credential
Manager or Linux Secret Service. Linux needs an available session D-Bus and
unlocked Secret Service collection. A secure-store failure is an error, not
permission to write credentials to a JSON file.

For headless use, supply `HUMBLEBEE_TAB_TOKEN` from the process environment and
connect with `--ephemeral`:

```sh
humblebee remote connect --name ci --url https://your-timeandbill-origin.example --ephemeral
humblebee remote status --profile ci --json
```

The token must already be set through your secure process/CI environment. It is
not persisted. Use `--profile` explicitly with this override, even with one
profile. Without an override, an omitted profile selects the sole stored profile;
multiple profiles always require a selection. There is no fallback to another
account or the local database.

An existing profile cannot be replaced silently. `connect --replace` confirms a
new server/account binding after verification. Every data command verifies the
current server identity and contract before its operation. Account mismatches
fail without rebinding the profile.

HTTPS certificate verification is mandatory. Redirects are never followed.
Development-only HTTP requires `--allow-loopback-http` on connect and a literal
loopback IP or `localhost`. This permission is stored with that profile; it does
not enable HTTP to arbitrary hosts or disable TLS verification.

## Read Data

```sh
humblebee remote projects --profile work
humblebee remote projects --profile work --all
humblebee remote tasks --profile work --project "$PROJECT_UUID"
humblebee remote tasks --profile work --project "$PROJECT_UUID" --all
humblebee remote day --profile work
humblebee remote day --profile work --date 2026-10-19
humblebee remote stopwatches --profile work
```

Copy UUIDs from the project/task listings into your own variables. Names and
local SQLite IDs are not remote identifiers. Default lists hide inactive projects
and completed/inactive tasks. `--all` exposes their availability but never makes
them bookable. Special absence projects are not part of the booking pilot.

The default day is today in the account timezone, not necessarily the computer's
timezone. Human-readable data output includes profile, account and timezone.
Explicit response timestamps are Unix seconds and durations are milliseconds;
legacy misleadingly named totals are not used.

## Book And Control Timers

```sh
humblebee remote book --profile work --project "$PROJECT_UUID" --task "$TASK_UUID" \
  --date 2026-10-19 --start 09:00 --end 10:00 --note-file ./note.txt

humblebee remote start --profile work --task "$TASK_UUID"
humblebee remote stop --profile work --id 123
humblebee remote discard --profile work --id 123 --yes
```

Manual bookings require an explicit project, task and completed interval on one
account-local date, using `YYYY-MM-DD` and 24-hour `HH:mm`. End `24:00` means
next midnight. Server-side subscription, ownership, task availability, booking
closure, overlap and DST checks still apply. Ambiguous/nonexistent DST times are
rejected. Bookings do not complete tasks or change browser project selection.

Use either `--note` or a UTF-8 `--note-file` (`-` reads stdin), with at most
1,000 UTF-16 code units, matching the server's text limit. Unicode and line breaks
are sent unchanged. Human output escapes terminal control/format characters;
JSON escapes preserve the original text when decoded, except actual credential
values, which are redacted at display boundaries.

Starting the already-running task reuses its timer without resetting it. Another
running task must be stopped/booked or discarded first. Every timer write sends
the exact observed state token. Stale state is rejected without retrying against
a newer timer. Stop always sends explicit booking intent and may infer its ID
only when exactly one timer runs. Discard requires an exact ID and confirmation;
non-interactive use must pass `--yes`. Discard never books elapsed time.

Booked-entry editing/deletion and project/task administration remain web-app
workflows. The pilot does not upload local databases or imported exports.

## Recover An Uncertain Write

Before dispatch, each deliberate write saves a random UUID and minimal local
attempt metadata. The server commits its result and deduplication receipt
together. The client never retries a write automatically, including after a
timeout, connection loss or restart.

If the response is lost, the CLI returns exit 6, prints the original mutation ID
and a recovery command:

```sh
humblebee remote operation --profile work "$MUTATION_ID"
```

This command only reads the existing outcome. It does not send the booking again.
`applied` confirms the committed operation and resource identifiers;
`not_applied` is a recorded rejection. `unknown`/expired/in-progress outcomes
remain uncertain (exit 6), not proof of failure. Do not resubmit with a new ID.
Inspect the web application and seek support if the original outcome cannot be
recovered. Recovery is bounded by the server's advertised window (at least seven
days); local metadata does not extend it or bypass token expiry/revocation.

## Files And Disconnection

Remote state is separate from SQLite, under `~/.humblebee/remote/` (or
`$HUMBLEBEE_HOME/remote/`):

- `profiles.json`: versioned names, normalized origins, identity bindings,
  timezone and opaque secure-store references. No app tokens.
- `operations/<mutation-id>.json`: profile/origin/identity, operation, creation
  time and minimal known outcome/resource IDs. No note or replayable payload.
- `profiles.lock`: exclusive configuration-edit lock. If left after a crash,
  verify no other HumbleBee process is changing profiles before removing it.

New configuration files use private permissions and atomic replacement. Recovery
metadata is not an offline queue or an account mirror. Do not delete it while
an operation is unresolved. Local `start`, `stop`, `show`, `report`, `import`,
`init` and `gui` neither contact Time & Bill nor implicitly switch modes.

```sh
humblebee remote disconnect --profile work
```

Disconnect removes the local profile/credential. It does not revoke the token,
stop timers or delete server data. Revoke the token separately in Time & Bill.

## Scripting

All remote commands accept `--json`. Data commands emit one object on stdout:
`version`, `ok`, `profile`, `accountIdentity`, `timezone`, `data`. Failures emit
`version`, `ok: false`, `error` with a stable `code`, safe `message` and
`mutationId` when relevant. Diagnostics/prompts use stderr. JSON mode does not
consent to discard; use `--yes` explicitly.

| Exit | Meaning |
| --- | --- |
| 0 | Success |
| 2 | Invalid arguments or input; confirmation/cancellation |
| 3 | Authentication, authorization, subscription or secure-store failure |
| 4 | Booking conflict, unavailable task, stale state or mutation conflict |
| 5 | Transport, server, local-state or protocol failure |
| 6 | Uncertain write; recover the original mutation ID |
| 7 | Unsupported/incomplete server safety contract |

Legacy local-command failures retain exit 1. HTTP 200 alone is never treated as
success: JSON-RPC errors, MCP `isError`, required fields, IDs, receipts and explicit
time units are validated. HTTP responses and token input are bounded; requests
have a 15-second overall timeout. No background polling, telemetry or automatic
network activity is added to local commands.

See [connected-cli-test-plan.md](connected-cli-test-plan.md) for verification and
remaining release/acceptance gates.
