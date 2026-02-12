# Releasing HumbleBee

This repo publishes releases via GitHub Actions + GoReleaser on semver tags (`v*`).

## One-time setup

1) These repositories need write access for releasing:
- `grobmeier/humblebee` (this repo)
- `grobmeier/homebrew-tap` (tap)
- `grobmeier/scoop-bucket` (Scoop bucket)

2) Add GitHub Actions secrets in `grobmeier/humblebee`:
- `TAP_GITHUB_TOKEN`: GitHub PAT with write access to `grobmeier/homebrew-tap`
- `SCOOP_GITHUB_TOKEN`: GitHub PAT with write access to `grobmeier/scoop-bucket`

Notes:
- The workflow uses `secrets.GITHUB_TOKEN` for creating the GitHub Release.
- The PATs are only for pushing formula/manifest commits to the other repos.

3) Ensure the release workflow + GoReleaser config are on `main`:
- `.github/workflows/release.yml`
- `.goreleaser.yaml`

## Release steps

0) (Optional) Validate the GoReleaser config:
```bash
goreleaser check
```

1) Update `main` and ensure it’s green locally:
```bash
git checkout main
git pull
go test ./...
```

2) Choose the version:
- First release example: `v0.1.0`

3) Create an annotated tag:
```bash
git tag -a v0.1.0 -m "v0.1.0"
```

4) Push the tag:
```bash
git push origin v0.1.0
```

5) Watch GitHub Actions:
- In `grobmeier/humblebee` → Actions → `release`
- The job runs GoReleaser and will:
  - create a GitHub Release with artifacts + `checksums.txt`
  - commit/update the Homebrew formula in `grobmeier/homebrew-tap`
  - commit/update the Scoop manifest in `grobmeier/scoop-bucket`

6) Verify the release outputs:
- GitHub Release contains:
  - `humblebee_<version>_windows_amd64.zip` (and other OS/arch archives)
  - `checksums.txt`
- Homebrew: check the formula commit landed in `grobmeier/homebrew-tap`
- Scoop: check the manifest commit landed in `grobmeier/scoop-bucket`

## Optional: local dry run (recommended before tagging)

Requires GoReleaser installed:
```bash
goreleaser release --snapshot --clean
```

This builds artifacts locally without publishing.

## Troubleshooting

- Homebrew/Scoop publishing fails: re-check PAT permissions and that the secret names match
  - `TAP_GITHUB_TOKEN`
  - `SCOOP_GITHUB_TOKEN`
- Wrong artifacts format on Windows: ensure `.goreleaser.yaml` has a `format_overrides` entry for `windows -> zip`
- Build metadata: `humblebee help` prints `Version: <version> (<commit>)` after release builds
- Release fails, remove tags local and remote:

    git tag -d v0.x.x
    git push --delete origin v0.x.x
    