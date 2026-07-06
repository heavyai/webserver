#!/bin/sh
# SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


rm -rf ./build ./storage

# This is even more hacky, but i don't really wanna spend time on this

# remove all docker containers if theres at least one, else log no containers found
if [ "$(docker ps -a | wc -l)" -gt "1" ]; then
	docker rm -f $(docker ps -a -q)
else
	echo "No containers found"
fi

docker rmi docker-internal.mapd.com/mapd/mapd-cpu:latest
docker rmi webserver
docker rmi webserver_heavydb_stack

docker builder prune -f

# remove all docker volumes if theres at least one, else log no volumes found
if [ "$(docker volume ls -q)" ]; then
	docker volume rm $(docker volume ls -q)
else
	echo "No volumes found"
fi
