# Connected CLI Test Plan

## Automated Local Checks

Use the Go version in `go.mod`; all remote tests use disposable HTTP fixtures,
temporary configuration directories and fake credential stores. They do not
contact a real account, write the user's OS vault or open a customer database.

```sh
go test ./...
go test -race ./internal/connected ./internal/remotecli ./internal/cli
go vet ./...
go build -o /tmp/humblebee-connected-pilot ./cmd/humblebee
```

Coverage includes authenticated MCP mapping; origin/redirect restrictions;
incompatible, malformed, oversized and error responses; profile identity binding;
secure-store separation/unavailability; unavailable tasks; explicit time units;
Unicode notes; exact timer targets and booking intent; mutation persistence before
dispatch; dropped-response recovery through a new client/store instance; cancellation
before dispatch; subprocess JSON/parser exit codes; and local-database isolation.

Compile the native credential-store adapters for supported target platforms:

```sh
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c \
  -o /tmp/humblebee-connected-windows.test.exe ./internal/connected
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go test -c \
  -o /tmp/humblebee-connected-linux.test ./internal/connected
```

Cross-compilation is not Windows/Linux runtime or native-vault acceptance. The
lost-response fixture proves a new client can recover using durable local metadata;
it is not a literal server-process restart or production upgrade rehearsal.

## Local Verification Record

On 2026-10-19, source revision `bee76f1` passed the complete Go suite:
10 test packages, 171 tests/subtests, no failures or skips. The connected client,
remote CLI and CLI packages also passed race checks; `go vet ./...`, module
verification and the local CLI build passed. Windows/Linux credential-adapter
tests cross-compiled successfully. No real OS-vault write or live account was
used. In particular, the subprocess test drops a booking response, launches a
second CLI process and recovers the original ID with exactly one server write,
while a fixture local database remains byte-for-byte unchanged.

## Manual CLI Checks

Use a disposable `HUMBLEBEE_HOME` and an explicitly approved test account. Do not
use a customer database, production projects or a token with unnecessary scopes.

1. Connect with a limited token; confirm status shows the expected account,
   origin and timezone. Test native vault storage on macOS, Windows and Linux.
   Inspect profile/operation files: no token, note or full request payload.
2. Repeat with an unavailable/locked vault: no plaintext fallback or saved usable
   profile. Test explicit ephemeral mode and required `--profile` selection.
3. Configure two profiles. Omitted selection must fail without contacting either
   account. Reconnect to another account only with explicit `--replace`.
4. List projects/tasks, including `--all`. Compare completed/inactive states with
   the web app and confirm the web's selected project did not change.
5. Read a day near the account's midnight from a computer in another timezone.
   Compare totals with the web app; ensure seconds/milliseconds are not mixed.
6. Book one hour using a note containing `Grüße`, `•`, `│`, line breaks,
   `&` and `<`. Verify actual Unicode characters and note lines in Time & Bill.
   Confirm the task remains unfinished and returned entry UUID is correct.
7. Reject an overlap, completed/inactive task, closed project and invalid or
   ambiguous/nonexistent DST interval without any new entry or timer change.
8. Start a task, repeat start and verify the original timestamp remains. Start a
   different task while it runs: fail without auto-booking or switching.
9. Stop/book once and confirm one web entry. Discard a different timer and confirm
   no entry. Non-interactive discard without `--yes` must fail without discarding.
10. Change/restart a timer from a second client between observation and write.
    Reject stale stop/discard without touching the replacement timer. Use an HTTP
    proxy/fixture under your control to hold the request at the precondition boundary.
11. Drop a response after the server commits, exit the CLI, and recover its saved
    mutation ID with `remote operation`. Verify one booking and no write replay.
    Repeat after an actual server-process restart in disposable staging.
12. Revoke/expire the app token. Subsequent reads, writes and receipt lookup must
    be denied. Verify permitted recovery/discard after subscription expiry without
    granting new booking rights. Disconnect locally; server timers/data stay intact.
13. Exercise `--json`, invalid flags, unavailable server and cancelled requests.
    Assert one JSON object, stable exit category, safe terminal controls and no
    credential exposure. Preserve Unicode/line breaks when decoding JSON.
14. Compare a fixture local SQLite database byte-for-byte before/after remote
    reads, failures and writes. With no local DB, none may be created. Retest
    normal local CLI/GUI workflows and offline use.

## Release Gates

- Deploy the corresponding server contract only with separate approval. Reconcile
  its schema/mutex integration with other Time & Bill branches first, rehearse a
  populated upgrade and test literal server-restart/lost-response recovery.
- All serving replicas must support contract version 1 before clients are enabled.
  Disable connected clients before rolling back to legacy server code; legacy
  tools may ignore safety arguments. Mixed-version routing is unsafe.
- Complete authenticated staging web/CLI interoperability and native-vault checks
  on the supported operating systems. Record exact revisions and evidence.
- Review and integrate the HumbleBee pilot separately. Do not present fixture,
  compilation or local Go-test success as live-account acceptance or a release.
- Keep GUI connected mode, offline queues/sync and local database upload out of
  this release. No hosted account is required for existing local features.
