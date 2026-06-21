#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

abort() {
  printf 'macOS app release aborted: %s\n' "$1" >&2
  exit 1
}

usage() {
  cat <<'EOF'
Usage:
  scripts/release-macos-app.sh <tag> [--no-upload]

Builds the HumbleBee Wails app on the local Mac, signs it with a Developer ID
Application certificate, notarizes it with Apple, staples the app, zips it, and
uploads the zip to the existing GitHub release asset name used by release-gui.

Required one-time local setup:
  1. Install a Developer ID Application certificate in the macOS login keychain.
  2. Store notarization credentials:
       xcrun notarytool store-credentials humblebee-notary
  3. Authenticate GitHub CLI:
       gh auth login

Environment overrides:
  HUMBLEBEE_CODESIGN_IDENTITY   Exact Developer ID Application identity.
  HUMBLEBEE_NOTARY_PROFILE      notarytool keychain profile. Default: humblebee-notary
  HUMBLEBEE_RELEASE_SOURCE_REF  Git ref to build. Default: <tag>
  WAILS_VERSION                 Wails CLI version. Default: v2.12.0

Examples:
  scripts/release-macos-app.sh v0.2.1
  scripts/release-macos-app.sh v0.2.1 --no-upload
  HUMBLEBEE_RELEASE_SOURCE_REF=main scripts/release-macos-app.sh v0.2.1
EOF
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || abort "missing required command: $1"
}

safe_tag_name() {
  printf '%s' "$1" | tr '/' '-'
}

asset_suffix_for_host() {
  case "$(uname -m)" in
    arm64) printf 'darwin_arm64' ;;
    x86_64) printf 'darwin_amd64' ;;
    *) abort "unsupported macOS architecture: $(uname -m)" ;;
  esac
}

find_codesign_identity() {
  local count
  count="$(
    security find-identity -v -p codesigning |
      awk -F'"' '/Developer ID Application:/ { count++ } END { print count + 0 }'
  )"

  if [[ "$count" == "0" ]]; then
    abort "no Developer ID Application certificate found in the keychain"
  fi

  if [[ "$count" != "1" ]]; then
    abort "multiple Developer ID Application certificates found; set HUMBLEBEE_CODESIGN_IDENTITY"
  fi

  security find-identity -v -p codesigning |
    awk -F'"' '/Developer ID Application:/ { print $2; exit }'
}

tag="${1:-}"
if [[ -z "$tag" || "$tag" == "-h" || "$tag" == "--help" ]]; then
  usage
  exit 0
fi
shift

upload=1
while [[ $# -gt 0 ]]; do
  case "$1" in
    --no-upload)
      upload=0
      shift
      ;;
    *)
      abort "unknown argument: $1"
      ;;
  esac
done

[[ "$tag" =~ ^v[0-9]+ ]] || abort "tag must look like v0.2.1"

require_command git
require_command go
require_command npm
require_command security
require_command codesign
require_command xcrun
require_command ditto
require_command spctl
if [[ "$upload" == "1" ]]; then
  require_command gh
fi

notary_profile="${HUMBLEBEE_NOTARY_PROFILE:-humblebee-notary}"
source_ref="${HUMBLEBEE_RELEASE_SOURCE_REF:-$tag}"
wails_version="${WAILS_VERSION:-v2.12.0}"
sign_identity="${HUMBLEBEE_CODESIGN_IDENTITY:-}"
if [[ -z "$sign_identity" ]]; then
  sign_identity="$(find_codesign_identity)"
fi

safe_tag="$(safe_tag_name "$tag")"
asset_suffix="$(asset_suffix_for_host)"
asset_name="HumbleBee_GUI_${safe_tag}_${asset_suffix}.zip"
workdir="$(mktemp -d "${TMPDIR:-/tmp}/humblebee-macos-release.XXXXXX")"
source_dir="$workdir/source"
release_dir="$repo_root/release"
archive_path="$release_dir/$asset_name"

cleanup() {
  git -C "$repo_root" worktree remove --force "$source_dir" >/dev/null 2>&1 || true
  rm -rf "$workdir"
}
trap cleanup EXIT

printf 'Preparing macOS app release asset\n'
printf '  release tag: %s\n' "$tag"
printf '  source ref:  %s\n' "$source_ref"
printf '  asset:       %s\n' "$asset_name"
printf '  identity:    %s\n' "$sign_identity"
printf '  notarizer:   %s\n' "$notary_profile"

if [[ "$upload" == "1" ]]; then
  gh release view "$tag" >/dev/null || abort "GitHub release not found for $tag"
fi

git fetch --tags --prune-tags origin
git worktree add --detach "$source_dir" "$source_ref"

(
  cd "$source_dir"
  go run "github.com/wailsapp/wails/v2/cmd/wails@$wails_version" build -clean -tags production
)

app_path="$source_dir/build/bin/HumbleBee.app"
[[ -d "$app_path" ]] || abort "Wails app was not produced at $app_path"

codesign --force --deep --timestamp --options runtime --sign "$sign_identity" "$app_path"
codesign --verify --deep --strict --verbose=2 "$app_path"

mkdir -p "$release_dir"
rm -f "$archive_path"
ditto -c -k --sequesterRsrc --keepParent "$app_path" "$archive_path"

xcrun notarytool submit "$archive_path" --keychain-profile "$notary_profile" --wait
xcrun stapler staple "$app_path"
xcrun stapler validate "$app_path"
spctl --assess --type execute --verbose "$app_path"

rm -f "$archive_path"
ditto -c -k --sequesterRsrc --keepParent "$app_path" "$archive_path"

if [[ "$upload" == "1" ]]; then
  gh release upload "$tag" "$archive_path" --clobber
  printf 'Uploaded signed and notarized macOS asset: %s\n' "$asset_name"
else
  printf 'Created signed and notarized macOS asset without upload: %s\n' "$archive_path"
fi
