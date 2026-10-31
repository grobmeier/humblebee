# Releasing HumbleBee

This is the maintainer runbook. A release has CLI archives, standalone GUI
downloads, a CycloneDX SBOM, and website download links. Publishing the GitHub
release page alone does not mean all of these steps have finished.

## Release Pipeline

| Stage | Workflow | Result |
| --- | --- | --- |
| Annotated `v*` tag pushed | `release-cli` (`release.yml`) | GoReleaser publishes CLI archives and checksums, and updates Homebrew/Scoop. |
| CLI release finishes | Explicit workflow dispatches | Starts `release-gui.yml` and `sbom.yml` for the same tag. |
| GUI packaging | `release-gui` | Uploads Linux x64, Windows x64, and macOS ARM64 GUI archives. |
| Release SBOM | `sbom` | Attaches `HumbleBee_<tag>_sbom.cdx.json` to the release. |
| All GUI builds succeed | `update-website-downloads` | Waits for all expected assets, then updates the website version for the latest stable release. |
| Website version commit | Landing repository's `Deploy to S3` | Builds and deploys the website, then requests CloudFront invalidation. |

The explicit dispatches matter: a release created with `GITHUB_TOKEN` does not
automatically trigger downstream workflows through the release event. The GUI
and SBOM workflows also accept a manually published release event.

CLI downloads contain the CLI; `humblebee gui` launches an available GUI app.
Standalone GUI downloads do not require a separate CLI installation.

The macOS release and CI jobs use the standard ARM64 `macos-15` runner. Keep
their labels aligned when migrating images; `macos-latest` can change OS versions.
The retired `macos-14` image must not be used. See the
[GitHub retirement notice](https://github.blog/changelog/2026-10-01-github-actions-macos-14-runner-image-retirement/).
CI builds and packages the ARM64 app without uploading release assets; local
Developer ID signing remains a separate step.

## One-Time Setup

Configure these Actions secrets in `grobmeier/humblebee`:

| Secret | Required access |
| --- | --- |
| `TAP_GITHUB_TOKEN` | Write access to `grobmeier/homebrew-tap`. |
| `SCOOP_GITHUB_TOKEN` | Write access to `grobmeier/scoop-bucket`. |
| `WEBSITE_RELEASE_TOKEN` | Fine-grained token restricted to `grobmeier/timeandbill-landing`, with Contents read/write. |

The workflow's normal `GITHUB_TOKEN` creates releases and dispatches workflows in
HumbleBee. It cannot write to the separate website repository. Do not reuse a
broad personal token or package-manager credentials for the website job. Check
token expiry and repository access when rotating credentials; never put token
values in source, logs, issues, or release notes.

The website publisher updates only `_data/humblebee_release.yml` on `develop`.
The landing repository owns its deployment workflow and AWS credentials; they
are not copied into HumbleBee. Website links must use
`releases/download/<configured-tag>/...`, not `releases/latest/download/...`,
so the previous downloads remain valid while the next release is uploading.

For macOS signing, install a Developer ID Application certificate, Apple command
line tools, and a notarization keychain profile on the release Mac:

```bash
xcrun notarytool store-credentials humblebee-notary
```

## Prepare And Publish

1. Obtain release approval and choose the next version from the latest published
   release. Use a patch version for compatible fixes and maintenance.
2. Confirm the intended changes are merged and pushed to `main`, the working tree
   is clean, and CI is green, including macOS/Windows/Linux GUI builds. Use the Go
   version in `go.mod` and keep the Wails CLI aligned with the module dependency.
3. Install the locked frontend dependencies and validate maintenance automation:

   ```bash
   npm ci --prefix frontend
   node --test scripts/*.test.cjs
   ```

4. Run the interactive helper from the repository root:

   ```bash
   scripts/release.sh
   ```

   It shows the highest local version tag after fetching tags, asks for the new
   version and confirmation, switches to `main`, pulls with `--ff-only`, runs
   Go tests and the frontend build, then creates and pushes an annotated tag.
   Compare its tag suggestion with GitHub's published releases; a tag alone is
   not evidence of a successfully published release.

   If you intend to accept automatic macOS signing, set the source override
   before starting the helper. For example, when creating `v0.7.2`:

   ```bash
   HUMBLEBEE_RELEASE_SOURCE_REF=v0.7.2 scripts/release.sh
   ```

5. On macOS, the helper offers to wait for the hosted archive and replace it with
   a signed/notarized build. Confirm the source override matches the new tag.
   Without that override, the current helper passes `origin/main` to the signer,
   which may advance after tagging. In that case, decline automatic signing and
   use the standalone exact-tag signing command below instead.
6. Monitor all release stages and complete the verification below. No separate
   manual website version edit is normally needed.

Do not move or delete a published release tag to repair a failure. Retry the
failed stage with the original tag, or publish a new version for source changes.

## Dependency Security Checks

For security maintenance, check the compiler as well as module versions. Use a
supported, patched Go toolchain meeting the minimum in `go.mod`; a dependency
upgrade alone cannot fix vulnerabilities in the standard library used to build
the executable.

```bash
go version
go mod verify
go mod tidy -diff
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

Also scan built CLI and GUI executables with `govulncheck -mode=binary` where
supported, because production build tags and platform dependencies can differ
from the source scan. Record the tested source revision, compiler and scan
results. A clean scan is not a guarantee that no vulnerabilities exist.

Keep Wails CLI pins aligned with its module version. The maintenance tests check
CI, GUI release, macOS packaging and contributor commands. Validate combined
dependency updates with the browser suite and all three native GUI CI builds;
in particular, keep React and React DOM on compatible matching versions.

## macOS Signing

GitHub-hosted macOS builds are not Developer ID signed/notarized. To sign or
repair an existing release, build from that release's exact tag. This example
uses the already published `v0.7.1`; replace it with the intended release tag:

```bash
HUMBLEBEE_RELEASE_SOURCE_REF=v0.7.1 scripts/release-macos-app.sh v0.7.1
```

The helper builds in a temporary worktree, signs, submits to Apple, staples and
validates the ticket, checks Gatekeeper acceptance, and replaces the matching
GitHub archive. Use `--no-upload` to validate without replacing a public asset.
Multiple signing identities require `HUMBLEBEE_CODESIGN_IDENTITY`; the notarization
profile can be changed with `HUMBLEBEE_NOTARY_PROFILE`.

Website automation checks asset availability, not notarization status. It can
publish links before local signing finishes; signing replaces the same archive
name afterward. Confirm the signed upload separately before describing the
macOS download as signed and notarized.

## Verify Completion

The examples below inspect `v0.7.1`; substitute the release being verified:

```bash
gh release view v0.7.1 --repo grobmeier/humblebee --json url,assets
gh run list --repo grobmeier/humblebee --limit 15
gh run list --repo grobmeier/timeandbill-landing --workflow main.yml --branch develop --limit 5
```

- Verify the release tag points to the intended commit and the release is neither
  draft nor prerelease when publishing stable website downloads.
- Expect six CLI archives: Darwin/Linux/Windows on AMD64 and ARM64. Windows uses
  ZIP; Darwin/Linux use tar.gz. Names omit the leading `v`, for example
  `humblebee_0.7.1_windows_amd64.zip`.
- Expect three standalone GUI archives:
  `HumbleBee_GUI_v0.7.1_darwin_arm64.zip`,
  `HumbleBee_GUI_v0.7.1_linux_amd64.tar.gz`, and
  `HumbleBee_GUI_v0.7.1_windows_amd64.zip`.
- Expect `checksums.txt` for the GoReleaser CLI archives and
  `HumbleBee_v0.7.1_sbom.cdx.json` as a downloadable release asset. The unversioned
  `humblebee-sbom.cdx.json` from ordinary CI is a workflow artifact, not the
  release attachment. The SBOM is a source-manifest inventory, not an attestation
  of each binary. `checksums.txt` does not cover the separately uploaded GUIs.
- Verify Homebrew and Scoop updates landed. For signed macOS downloads, confirm
  Apple's acceptance, ticket stapling, Gatekeeper acceptance, and the final
  uploaded archive; hosted GUI build success alone does not prove signing.
- Check the website publisher and landing deployment results. Verify the three
  download links on each of `/en/humblebee/`, `/de/humblebee/`, `/es/humblebee/`,
  `/fr/humblebee/`, and `/it/humblebee/` on `https://www.timeandbill.de`.

Website publishing skips prereleases and releases that are not the latest stable
release. It waits up to ten minutes for all eleven expected assets, refuses to
roll back a newer website version, and makes same-version retries a no-op. A
successful no-op verifies release readiness and website access, but does not
exercise a new website commit or deployment.

## Release Notes

Review the changes since the preceding release and write concise user-facing
notes covering fixes, features, security updates, and relevant build requirements.
Generated commit/merge titles alone are not adequate release notes. Do not claim
that assets, signing, or deployments are complete until verified.

Save the reviewed notes to a local file and update the release:

```bash
gh release edit v0.7.1 --repo grobmeier/humblebee --notes-file /path/to/release-notes.md
```

Include a compare link to the previous version. Mention Go/toolchain requirements
as source-build requirements rather than installation requirements for GUI users.

## Recovery

| Failure | Recovery |
| --- | --- |
| CLI/Homebrew/Scoop | Inspect `release-cli` logs and token permissions; retry the failed run for the original tag. |
| Missing SBOM | Dispatch `sbom.yml` with the intended tag. |
| Website token missing/expired | Repair `WEBSITE_RELEASE_TOKEN`, then run only the standalone publisher below. |
| Website assets missing | Repair the failed asset stage, then retry the standalone publisher. |
| Website deployment failed | Inspect/rerun `Deploy to S3` in the landing repository; a version commit is not deployment proof. |
| GUI packaging failed | Inspect the failed matrix job and rerun only that job where possible. Rebuilding macOS can overwrite a signed archive; sign/upload it again if necessary. |

```bash
gh workflow run sbom.yml --repo grobmeier/humblebee -f tag=v0.7.1
gh workflow run update-website-downloads.yml --repo grobmeier/humblebee -f tag=v0.7.1
```

The standalone website workflow never rebuilds or replaces GUI archives. Use it
for website recovery instead of rerunning the entire GUI release workflow.
