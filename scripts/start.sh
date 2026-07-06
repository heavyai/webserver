#!/bin/bash
# SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0


# Hacky script to hack around heavydb conatiner not liking /var/lib/heavyai/storage being volumed

CLEAN=$1

STORAGE_PATH_CONTAINER=/var/lib/heavyai/storage
STORAGE_PATH_CONTAINER_TMP=/var/lib/heavyai/tmp
STORAGE_PATH_LOCAL=$(pwd)/storage

checkForHeavyImage() {
	VPN_AVAILABLE=$(curl -s -o /dev/null -w "%{http_code}" https://kali.mapd.com:10043/login)
	IMAGE_AVAILABLE=$(docker images | grep mapd/mapd-cpu | wc -l) # returns 1 if image is available, 0 if not

	if [ "$VPN_AVAILABLE" != "200" ] && [ "$IMAGE_AVAILABLE" == "0" ]; then
		echo "ERROR: docker-internal.mapd.com/mapd/mapd-cpu image is not available. Try again with a VPN connection."
		exit 1
	fi
}

seedStorageDir() {
	echo "INFO: Seeding storage directory"
	docker run --rm -d --name webserver -v "$STORAGE_PATH_LOCAL:$STORAGE_PATH_CONTAINER_TMP" webserver:latest

	while true; do
		echo "INFO: Waiting for the webserver to start..."
		if grep -q "http server started on" <(docker logs webserver); then
			echo "INFO: seeding the storage directory..."
			docker exec -it webserver bash -c "cp -r $STORAGE_PATH_CONTAINER/* $STORAGE_PATH_CONTAINER_TMP"
			echo "INFO: Storage directory seeded"
			docker stop webserver
			break
		fi
		sleep 2
	done
}

main() {
	if [ "$CLEAN" == "clean" ]; then
		echo "INFO: Cleaning up"
		./scripts/clean.sh
	fi

	checkForHeavyImage
	docker build --pull --rm -f "Dockerfile.local" -t webserver:latest "."

	if [ ! -d "$STORAGE_PATH_LOCAL/data" ]; then
		seedStorageDir
	fi

	# Let'sa go!
	docker-compose up --remove-orphans
}

main
