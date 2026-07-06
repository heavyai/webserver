#!/usr/bin/env bash
# SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

set -Eeuo pipefail

####################################################################################################
# This files includes utility functions, and is meant to be sourced by other scripts, not run
# directly.
####################################################################################################


# Returns a list of first-party *.go source code files, which can then be used an an argument for
# lists for commands to operate on. Assumes we are in the repo root.
function find_src_files() {
    # TODO: this is probably done better with xargs (at the least, so we can chunk the list when it
    # gets really long and execute multiple passes of the command for each chunk).
    find . \( -path './internal/db_thrift_client' -or -path './archive' \) -prune -or -name '*.go' -print
}


find_gotool() {
    if [[ $# -ne 1 ]] || [[ -z $1 ]]; then
        return 1
    fi

    local TOOL_FULLNAME="$1"
    local TOOL="$(basename "$TOOL_FULLNAME")"
    local TOOL_PATH="$(go env GOPATH)/bin/${TOOL}"
    if [[ ! -e "$TOOL_PATH" ]]; then
        go get "$TOOL_FULLNAME" >/dev/null || return 2
    fi
    echo "$TOOL_PATH"
}

# Prints the path to the `goimports` tool, installing it first if necessary
function find_goimports() {
    find_gotool "golang.org/x/tools/cmd/goimports"
}

function find_golicense() {
    find_gotool "github.com/heavyai/golicense"
}

function find_golint() {
    find_gotool "golang.org/x/lint/golint"
}


function print_headline() {
    printf '################## [%s]\n' "$1"
}

# Returns 0 if the script is currently running under Jenkins, 1 otherwise
function is_jenkins() {
    if [ -z ${BUILD_NUMBER+x} ]; then
        return 1
    else
        return 0
    fi
}
