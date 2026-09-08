# SPDX-FileCopyrightText: Copyright (c) 2026, NVIDIA CORPORATION & AFFILIATES. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

NO_COLOR=\x1b[0m

OK_COLOR=\x1b[32;01m

ERROR_COLOR=\x1b[31;01m

WARN_COLOR=\x1b[33;01m

OK_STRING=$(OK_COLOR)[OK]$(NO_COLOR)

ERROR_STRING=$(ERROR_COLOR)[ERRORS]$(NO_COLOR)

WARN_STRING=$(WARN_COLOR)[WARNINGS]$(NO_COLOR)

.DEFAULT_GOAL := deploy

ifndef VERBOSE
.SILENT:
endif

deploy:
		docker-compose build && \
		docker-compose up -d && \
		echo "$(OK_STRING) Run webserver with jupyter flag --jupyter-url http://localhost:8000" && \
		echo "$(OK_STRING) Add \"enableJupyter\": true to servers.json"
		echo "$(OK_STRING) Default OmniSciDB Credentials: admin / HyperInteractive / omnisci"
clean:
		docker-compose down
		echo "$(OK_STRING) Jupyter Removed"

