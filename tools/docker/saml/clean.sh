#!/bin/bash
# SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


docker-compose down
rm -rf ./mapd-immerse
rm -rf ./omniscidata
rm heavy_web_server
docker volume rm saml_mysql_data