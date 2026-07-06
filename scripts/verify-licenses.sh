#!/usr/bin/env bash
# SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

##############################################################################
# Check the license of each Go dependency to ensure we aren't shipping
# code that is GPL (or any other license type we don't want).

# This uses https://github.com/mitchellh/golicense to do the heavy lifting.
# Note this checks the repo license via the GH API.

# The concatenated list of licenses will be output to build/*/third-party-licenses.txt

# GITHUB_KEY=<some api key> ./verify-licenses.sh
##############################################################################

# Running without a GH API key will result in rate limiting so bail out
# if GITHUB_TOKEN isn't available.

set -e
cd "$(dirname "${BASH_SOURCE[0]}")/.."
source ./scripts/util.sh

ERROR_CODE=1

if [[ -z ${GITHUB_TOKEN+x} ]]; then
  echo "GITHUB_TOKEN not found! Run with \`GITHUB_TOKEN=<some api key> ./verify-licenses.sh\` or export the env var"
  exit $ERROR_CODE
fi

GOLICENSE_BINARY="$(find_golicense)"
OUTPUT_FILENAME="third-party-licenses.txt"
mapfile -t BINARIES < <(find ./build/*/ -name 'heavy_web_server')

print_headline "Checking licenses"

for ((i = 0; i < ${#BINARIES[@]}; i++)); do
  OUTPUT_FILEPATH="$(dirname "${BINARIES[$i]}")/$OUTPUT_FILENAME"
  printf '\nAnalyzing binary: %s\n' "${BINARIES[$i]}"

  GITHUB_TOKEN=${GITHUB_TOKEN} $GOLICENSE_BINARY -plain --out-licensefile "$OUTPUT_FILEPATH" golicense.cfg.json "${BINARIES[$i]}"
  GOLICENSE_EXIT_CODE=$?

  if [[ $GOLICENSE_EXIT_CODE -ne 0 ]]; then
    printf "Golicense failed!\n"
    exit $GOLICENSE_EXIT_CODE
  fi

  if [ ! -s "$OUTPUT_FILEPATH" ]; then
    printf "No licenses detected in binary!\n"
    exit $ERROR_CODE
  fi
done
