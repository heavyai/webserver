#!/bin/bash
# SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

echo "############################"
echo "Pulling latest omnisci"
docker pull heavy/heavy-ee-cpu:latest
echo "############################"

echo "Building new webserver"
cd ../../../scripts
./build-all.sh
cd ../build/Linux-x86_64
cp heavy_web_server ../../tools/docker/saml/

cd ../../tools/docker/saml/

docker-compose stop heavy-server
docker-compose rm heavy-server
docker-compose build heavy-server
docker-compose up -d heavy-server
