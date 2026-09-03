# SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

FROM goboring/golang:1.13.4b4

RUN mkdir /webserver
COPY ./src /webserver
RUN /webserver/scripts/build-all.sh
