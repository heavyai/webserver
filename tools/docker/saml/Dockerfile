# SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

FROM omnisci/omnisci-ee-cpu:latest

COPY heavy_web_server /omnisci/bin
COPY ./mapd-immerse/dist /omnisci/frontend
COPY ./servers.json.example /omnisci/frontend/servers.json