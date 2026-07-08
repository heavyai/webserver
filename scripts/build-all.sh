#!/usr/bin/env bash
# SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

set -e
cd "$(dirname "${BASH_SOURCE[0]}")/.."
source ./scripts/util.sh

print_headline "Building for all environments"

# The build directory names were chosen to match the cmake environment variable names when run on
# the corresponding platform
echo "* Building Linux amd64"
GOOS=linux GOARCH=amd64 ./scripts/build.sh "build/Linux-x86_64"
echo -e "\n* Building Linux arm64"
GOOS=linux GOARCH=arm64 ./scripts/build.sh "build/Linux-aarch64"
echo

print_headline "Compressing binaries"
for bin in ./build/*/heavy_web_server; do
    # Get the file's bottom-level directory name, which (in this case) is the OS/arch name
    TAR_FILENAME="$(basename "$(dirname "$bin")")-heavy_web_server.tar.gz"
    echo "* Compressing $bin to ./build/${TAR_FILENAME}"
    (tar -czf "./build/${TAR_FILENAME}" -C "$(dirname "$bin")" "$(basename "$bin")")
done
echo

print_headline "All builds completed"
