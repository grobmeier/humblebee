# HumbleBee Maintenance

Follow the workspace instructions as well as this repository guidance.

## Release Work

- Use [RELEASE.md](RELEASE.md) as the canonical maintainer runbook. Keep README,
  GUI and contribution documentation linked to it; avoid divergent release steps.
- Release workflows, `go.mod`, `.goreleaser.yaml` and the release scripts are the
  source of truth for toolchain versions, assets, credentials and triggers.
  Confirm them before changing or reporting release behaviour.
- Keep Wails CLI and module versions aligned. Native builds may generate binding
  whitespace, dependency fingerprints and platform files; inspect the diff and
  preserve unrelated user changes rather than committing build churn.
- Build signed macOS replacements from the exact release tag, not a newer main.
  Never rerun GUI packaging merely to test website publishing: it can replace a
  signed/notarized archive. Use `update-website-downloads.yml` for that purpose.
- Website version changes are automated after assets are ready. Keep website
  links pinned to the configured tag, preserve latest-stable/idempotency/rollback
  guards, and keep website credentials separate from Homebrew/Scoop credentials.
- Validate publisher changes with `node --test scripts/update-website-release.test.cjs`.
  Lint modified workflows and run `bash -n` on changed shell scripts. Pure
  documentation changes need link/consistency checks rather than native rebuilds.
- Report source integration, CI, published assets, signing, website deployment
  and live verification separately. A no-op publishing test does not prove a new
  website commit or deployment. Curate release notes; do not rely on merge titles.
- Never move/delete published tags as routine recovery, expose credential values,
  or publish a new release solely to test maintenance changes. Existing workspace
  approval requirements still apply to commits, pushes, merges and releases.
