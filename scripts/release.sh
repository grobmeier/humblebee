#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

abort() {
  printf 'Release aborted: %s\n' "$1" >&2
  exit 1
}

require_clean_worktree() {
  if [[ -n "$(git status --porcelain)" ]]; then
    git status --short
    abort "working tree is not clean"
  fi
}

latest_release_tag() {
  git for-each-ref \
    --sort=-v:refname \
    --format='%(refname:short)' \
    'refs/tags/v[0-9]*' | head -n 1
}

safe_tag_name() {
  printf '%s' "$1" | tr '/' '-'
}

asset_suffix_for_host() {
  case "$(uname -m)" in
    arm64) printf 'darwin_arm64' ;;
    x86_64) printf 'darwin_amd64' ;;
    *) abort "unsupported macOS architecture for GUI signing: $(uname -m)" ;;
  esac
}

wait_for_github_gui_asset() {
  local tag="$1"
  local safe_tag asset_suffix asset_name timeout interval start now elapsed

  command -v gh >/dev/null 2>&1 || abort "missing required command for macOS signing: gh"

  safe_tag="$(safe_tag_name "$tag")"
  asset_suffix="$(asset_suffix_for_host)"
  asset_name="HumbleBee_GUI_${safe_tag}_${asset_suffix}.zip"
  timeout="${HUMBLEBEE_RELEASE_ASSET_WAIT_SECONDS:-3600}"
  interval="${HUMBLEBEE_RELEASE_ASSET_WAIT_INTERVAL_SECONDS:-30}"
  start="$(date +%s)"

  printf 'Waiting for GitHub release asset before uploading signed replacement: %s\n' "$asset_name"
  while true; do
    if gh release view "$tag" --json assets --jq '.assets[].name' 2>/dev/null | grep -Fxq "$asset_name"; then
      printf 'Found GitHub release asset: %s\n' "$asset_name"
      return
    fi

    now="$(date +%s)"
    elapsed="$((now - start))"
    if (( elapsed >= timeout )); then
      abort "timed out waiting for GitHub release asset ${asset_name}"
    fi

    printf 'Still waiting for %s (%ss elapsed)...\n' "$asset_name" "$elapsed"
    sleep "$interval"
  done
}

validate_version() {
  local version="$1"

  if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]; then
    abort "version must look like 0.2.1 or 0.2.1-rc.1"
  fi
}

require_clean_worktree

printf 'Fetching tags from origin...\n'
git fetch --tags --prune-tags origin

latest_tag="$(latest_release_tag || true)"
if [[ -z "$latest_tag" ]]; then
  printf 'Latest published version: none found\n'
else
  printf 'Latest published version: %s\n' "$latest_tag"
fi

read -r -p 'New release version (without leading v): ' version
version="${version#v}"
validate_version "$version"

tag="v${version}"
if git rev-parse -q --verify "refs/tags/${tag}" >/dev/null; then
  abort "tag ${tag} already exists locally"
fi

if git ls-remote --exit-code --tags origin "refs/tags/${tag}" >/dev/null 2>&1; then
  abort "tag ${tag} already exists on origin"
fi

cat <<EOF

Release plan:
  repository: ${repo_root}
  latest:     ${latest_tag:-none}
  new tag:    ${tag}

The script will now:
  1. switch to main
  2. pull origin/main with --ff-only
  3. run go test ./...
  4. run npm --prefix frontend run build
  5. create annotated tag ${tag}
  6. push ${tag} to origin

EOF

read -r -p "Type '${version}' to continue: " confirmation
if [[ "$confirmation" != "$version" ]]; then
  abort "confirmation did not match ${version}"
fi

git switch main
git pull --ff-only
require_clean_worktree

go test ./...
npm --prefix frontend run build

git tag -a "$tag" -m "Release ${tag}"
git push origin "$tag"

printf 'Release tag %s pushed. GitHub Actions will build and publish the release assets.\n' "$tag"

if [[ "$(uname -s)" == "Darwin" ]]; then
  read -r -p 'Wait for the macOS GUI asset and replace it with a signed/notarized build? [y/N]: ' sign_macos
  if [[ "$sign_macos" =~ ^[Yy]$ ]]; then
    wait_for_github_gui_asset "$tag"
    HUMBLEBEE_RELEASE_SOURCE_REF="${HUMBLEBEE_RELEASE_SOURCE_REF:-origin/main}" \
      scripts/release-macos-app.sh "$tag"
  fi
else
  printf 'Skipping local macOS GUI signing because this host is not macOS.\n'
fi
