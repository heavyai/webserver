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

echo "############################"
echo "Building Immerse"
cd ../../tools/docker/saml/
rm -rf mapd-immerse/
git clone git@github.com:heavy/mapd-immerse.git
cd mapd-immerse
npm install
sh ./scripts/build-prod.sh
echo "############################"
echo "Doing the Docker thingamajig"
docker-compose build
docker-compose up -d


echo "Keycloak at http://localhost:8081 ...... Username: admin Password: Pa55w0rd"