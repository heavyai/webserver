#!/usr/bin/env bash
# SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

set -e
cd "$(dirname "${BASH_SOURCE[0]}")/.."
source ./scripts/util.sh

if ! type -t go >/dev/null; then
  echo '!!! Error: could not find `go` command'
  exit 1
fi


##########################
# Formatting checker
##########################

print_headline "Checking Formatting"
GOIMPORTS_PATH="$(find_goimports)"
SRC_FILES="$(find_src_files)"

GOIMPORTS_OUTPUT="$($GOIMPORTS_PATH -l -local 'github.com/heavyai/webserver' $SRC_FILES)"

if [[ -z $GOIMPORTS_OUTPUT ]]; then
  echo "...no formatting problems found"
  echo
else
  echo
  echo "!!! There were formatting problems with the following files:"
  echo "$GOIMPORTS_OUTPUT"
  echo
  echo "Run the following command to see the incorrect formatting:"
  echo "  goimports -d -local 'github.com/heavyai/webserver' <filename>"
  echo "Run the following command to automatically fix the file:"
  echo "  goimports -w -local 'github.com/heavyai/webserver' <filename>"
  exit 2
fi


##########################
# Linter
##########################

print_headline "Checking Linting"
GOLINT_PATH="$(find_golint)"
SRC_DIRS="./cmd/... ./handlers/... ./middlewares/... ./models/... ./internal/constants/... ./internal/util/..."

if ! $GOLINT_PATH -set_exit_status $SRC_DIRS; then
  echo
  echo "!!! Linting error detected. Exiting."
  exit 10
fi


echo
print_headline "Done"
echo "Code verified successfully"
