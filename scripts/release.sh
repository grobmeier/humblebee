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
