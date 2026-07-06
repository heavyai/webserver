#!/usr/bin/env bash
# SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

set -e
cd "$(dirname "${BASH_SOURCE[0]}")/.."
source ./scripts/util.sh

BIN_FILENAME="heavy_web_server"
OUT_DIR="${1:-build}"

if [ -z ${BUILD_NUMBER+x} ]; then
  JENKINS=0
else
  JENKINS=1
fi

GIT_HASH="$(git rev-parse --short HEAD)"

## Commented-out until we can add safety checks to ensure you're not deleting `/` or something
# [[ -e "$OUT_DIR" ]] && rm -rf "$OUT_DIR"
mkdir -p "$OUT_DIR"

set -x
go build -ldflags "-X github.com/heavyai/webserver/internal/config.commit=${GIT_HASH}" -o "${OUT_DIR}/${BIN_FILENAME}" ./cmd/immerse
