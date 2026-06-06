#!/usr/bin/env bash
set -euo pipefail

readonly COPYRIGHT=' * Copyright 2026 Grobmeier Solutions GmbH. All Rights Reserved.'

readonly HEADER="$(cat <<'EOF'
/*
 * Copyright 2026 Grobmeier Solutions GmbH. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
EOF
)"

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

added=0
updated=0
skipped=0

strip_existing_header() {
  local file="$1"
  local tmp="$2"

  awk '
    BEGIN {
      in_header = 1
      header_lines = 0
      saw_copyright = 0
      saw_license = 0
      saw_end = 0
      skip_blank_after_header = 0
    }

    in_header && header_lines == 0 && $0 == "/*" {
      header_lines++
      header[header_lines] = $0
      next
    }

    in_header && header_lines > 0 {
      header_lines++
      header[header_lines] = $0
      if ($0 ~ /^ \* Copyright [0-9][0-9][0-9][0-9] Grobmeier Solutions GmbH\. All Rights Reserved\.$/) {
        saw_copyright = 1
      }
      if ($0 ~ /^ \* Licensed under the Apache License, Version 2\.0/) {
        saw_license = 1
      }
      if ($0 ~ /^ \* limitations under the License\.$/) {
        saw_end = 1
      }
      if ($0 == " */") {
        in_header = 0
        if (saw_copyright && saw_license && saw_end) {
          skip_blank_after_header = 1
          next
        }
        for (i = 1; i <= header_lines; i++) {
          print header[i]
        }
        next
      }
      next
    }

    in_header {
      in_header = 0
    }

    skip_blank_after_header && $0 == "" {
      skip_blank_after_header = 0
      next
    }

    {
      skip_blank_after_header = 0
    }

    { print }
  ' "$file" > "$tmp"
}

while IFS= read -r file; do
  tmp="$(mktemp)"
  stripped="$(mktemp)"
  strip_existing_header "$file" "$stripped"
  { printf '%s\n\n' "$HEADER"; cat "$stripped"; } > "$tmp"

  if cmp -s "$file" "$tmp"; then
    skipped=$((skipped + 1))
    rm -f "$tmp" "$stripped"
    continue
  fi

  if cmp -s "$file" "$stripped"; then
    added=$((added + 1))
  else
    updated=$((updated + 1))
  fi

  mv "$tmp" "$file"
  rm -f "$stripped"
done < <(git ls-files '*.ts' '*.tsx')

printf 'Added license header to %d TypeScript files.\n' "$added"
printf 'Updated license header in %d TypeScript files.\n' "$updated"
printf 'Skipped %d TypeScript files with current license header.\n' "$skipped"
