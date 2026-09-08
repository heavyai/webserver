# SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

ifeq ($(user),)
# USER retrieved from env, UID from shell.
HOST_USER ?= $(strip $(if $(USER),$(USER),nodummy))
HOST_UID ?= $(strip $(if $(shell id -u),$(shell id -u),4000))
else
# allow override by adding user= and/ or uid=  (lowercase!).
# uid= defaults to 0 if user= set (i.e. root).
HOST_USER = $(user)
HOST_UID = $(strip $(if $(uid),$(uid),0))
endif

.DEFAULT_GOAL := build

build:
		@echo "Building"
		mkdir src && \
		cd ../../../ && \
		rsync -avz ./ tools/docker/fips140-2/src/ --exclude=tools/ --exclude=data/ --exclude=build/ && \
		cd tools/docker/fips140-2 && \
		docker build -t omnisci-webserver-fips . && \
		docker run -it -v $(CURDIR)/build-artifacts:/build-artifacts omnisci-webserver-fips:latest cp -r /webserver/build/ /build-artifacts && \
		echo "Binaries located at $(CURDIR)/build-artifacts"
clean:
		@echo "Cleaning up..."
		rm -rf build-artifacts/
		rm -rf src/
help:
		@echo ''
		@echo 'Usage: make [TARGET]'
		@echo 'Targets:'
		@echo '  build    	build fips1402 docker --image-- for current user: $(HOST_USER)(uid=$(HOST_UID))'
		@echo '  test     	test fips1402 docker image was build correctly for current user: $(HOST_USER)(uid=$(HOST_UID))'
		@echo '  clean    	remove docker --image-- for current user: $(HOST_USER)(uid=$(HOST_UID))'
		@echo ''

test:
		@echo "Starting FIPS Tests"
		cd $(CURDIR)/build-artifacts/build/Linux-x86_64 && \
		go tool nm heavy_web_server > test.txt && \
		grep '_Cfunc__goboringcrypto_' test.txt > /dev/null && \
		echo "Tests Passed, Found Correct Crypto Library in Webserver Binary"
