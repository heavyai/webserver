#!/usr/bin/env bash
# SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0
#
# Generates license/THIRD_PARTY_LICENSES.md by scanning the Go module cache
# for all dependencies listed in go.mod.
#
# Usage: ./scripts/generate-third-party-licenses.sh
# Requires: go (with modules), standard Unix tools

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

OUTPUT_DIR="third_party_licenses"
OUTPUT_FILE="$OUTPUT_DIR/THIRD_PARTY_LICENSES.md"

mkdir -p "$OUTPUT_DIR"

# Classify a license file by inspecting its text.
classify_license() {
  local file="$1"
  local text
  text=$(tr '[:upper:]' '[:lower:]' < "$file")

  # Apache: "Apache License" and "Version 2" may be on separate lines.
  if echo "$text" | grep -q "apache license" && echo "$text" | grep -q "version 2"; then
    echo "Apache-2.0"
  # ISC: distinctive "permission to use...and/or distribute...fee is hereby granted" phrasing.
  elif echo "$text" | grep -q "permission to use, copy, modify, and/or distribute" && echo "$text" | grep -q "fee is hereby granted"; then
    echo "ISC"
  # MIT: "Permission is hereby granted, free of charge" is the canonical MIT fingerprint.
  elif echo "$text" | grep -q "permission is hereby granted, free of charge"; then
    echo "MIT"
  elif echo "$text" | grep -q "mozilla public license" && echo "$text" | grep -q "2\.0"; then
    echo "MPL-2.0"
  elif echo "$text" | grep -q "neither the name.*nor the names\|bsd 3-clause\|bsd.*3.*clause"; then
    echo "BSD-3-Clause"
  elif echo "$text" | grep -q "redistribution.*binary.*must reproduce\|bsd 2-clause\|bsd.*2.*clause"; then
    echo "BSD-2-Clause"
  elif echo "$text" | grep -q "gnu lesser general public license" && echo "$text" | grep -q "version 2\.1"; then
    echo "LGPL-2.1"
  elif echo "$text" | grep -q "gnu lesser general public license"; then
    echo "LGPL"
  elif echo "$text" | grep -q "gnu general public license" && echo "$text" | grep -q "version 2"; then
    echo "GPL-2.0"
  elif echo "$text" | grep -q "gnu general public license" && echo "$text" | grep -q "version 3"; then
    echo "GPL-3.0"
  elif echo "$text" | grep -q "creative commons.*zero\|cc0-1\.0"; then
    echo "CC0-1.0"
  elif echo "$text" | grep -q "this is free and unencumbered software released into the public domain"; then
    echo "Unlicense"
  else
    echo "Unknown"
  fi
}

# Map SPDX identifiers to their canonical URLs.
license_url() {
  case "$1" in
    Apache-2.0)   echo "https://spdx.org/licenses/Apache-2.0.html" ;;
    MIT)          echo "https://spdx.org/licenses/MIT.html" ;;
    ISC)          echo "https://spdx.org/licenses/ISC.html" ;;
    MPL-2.0)      echo "https://spdx.org/licenses/MPL-2.0.html" ;;
    BSD-3-Clause) echo "https://spdx.org/licenses/BSD-3-Clause.html" ;;
    BSD-2-Clause) echo "https://spdx.org/licenses/BSD-2-Clause.html" ;;
    LGPL-2.1)     echo "https://spdx.org/licenses/LGPL-2.1.html" ;;
    LGPL)         echo "https://spdx.org/licenses/LGPL-2.1.html" ;;
    GPL-2.0)      echo "https://spdx.org/licenses/GPL-2.0.html" ;;
    GPL-3.0)      echo "https://spdx.org/licenses/GPL-3.0.html" ;;
    CC0-1.0)      echo "https://spdx.org/licenses/CC0-1.0.html" ;;
    Unlicense)    echo "https://spdx.org/licenses/Unlicense.html" ;;
    *)            echo "" ;;
  esac
}

echo "Scanning module cache..."

# Collect module info: "path\tversion\tdir" per line, skip the main module (no Dir).
mapfile -t MODULES < <(
  go mod download -json 2>/dev/null \
  | python3 -c "
import json, sys, re
data = sys.stdin.read()
# go mod download outputs a stream of JSON objects
decoder = json.JSONDecoder()
pos = 0
while pos < len(data):
    data_stripped = data[pos:].lstrip()
    if not data_stripped:
        break
    pos = len(data) - len(data_stripped)
    try:
        obj, end = decoder.raw_decode(data, pos)
        pos += end - pos
        if obj.get('Dir'):
            print(obj['Path'] + '\t' + obj['Version'] + '\t' + obj['Dir'])
    except json.JSONDecodeError:
        break
" \
  | sort -f
)

TOTAL=${#MODULES[@]}
echo "Found $TOTAL modules."

# Build markdown rows.
ROWS=()
UNKNOWN=()

for entry in "${MODULES[@]}"; do
  IFS=$'\t' read -r mod_path mod_version mod_dir <<< "$entry"

  # Find the license file (LICENSE, LICENSE.txt, LICENSE.md, COPYING, etc.)
  license_file=$(find "$mod_dir" -maxdepth 1 -iname 'license*' -o -maxdepth 1 -iname 'copying*' 2>/dev/null | head -1)

  if [[ -z "$license_file" ]]; then
    license_type="Unknown"
    url=""
    UNKNOWN+=("$mod_path")
  else
    license_type=$(classify_license "$license_file")
    url=$(license_url "$license_type")
    if [[ "$license_type" == "Unknown" ]]; then
      UNKNOWN+=("$mod_path")
    fi
  fi

  if [[ -n "$url" ]]; then
    url_cell="[$url]($url)"
  else
    url_cell=""
  fi

  ROWS+=("| $mod_path | $mod_version | $license_type | $url_cell |")
done

# Write the markdown file.
{
  echo "# Third-Party Licenses"
  echo ""
  echo "This file lists the third-party Go modules distributed with this project (the dependency closure of \`go.mod\`) and their licenses, in fulfillment of the attribution requirements of those licenses."
  echo ""
  echo "Generated from the installed module cache. Total packages: **$TOTAL**."
  echo ""
  echo "## Summary"
  echo ""
  echo "| Package | Version | License | URL |"
  echo "|---|---|---|---|"
  for row in "${ROWS[@]}"; do
    echo "$row"
  done
} > "$OUTPUT_FILE"

echo "Written $TOTAL packages to $OUTPUT_FILE"

if [[ ${#UNKNOWN[@]} -gt 0 ]]; then
  echo ""
  echo "Warning: could not classify licenses for ${#UNKNOWN[@]} package(s):"
  for pkg in "${UNKNOWN[@]}"; do
    echo "  - $pkg"
  done
fi
