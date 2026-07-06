NO_COLOR=\x1b[0m

OK_COLOR=\x1b[32;01m

ERROR_COLOR=\x1b[31;01m

WARN_COLOR=\x1b[33;01m



OK_STRING=$(OK_COLOR)[OK]$(NO_COLOR)

ERROR_STRING=$(ERROR_COLOR)[ERRORS]$(NO_COLOR)

WARN_STRING=$(WARN_COLOR)[WARNINGS]$(NO_COLOR)

.DEFAULT_GOAL := build


build:
		@echo "Building"
		cd ../../../ && \
		./scripts/build-all.sh && \
		cp build/Linux-x86_64/heavy_web_server tools/docker/headers-testing/webserver-proxy/heavy_web_server && \
		cd tools/docker/headers-testing && \
		docker-compose build && \
		docker-compose up -d
clean:
		docker-compose down
		rm -f webserver-proxy/heavy_web_server
test: 	
		curl -s -i -H "X-OmniSci-UserName: admin" localhost:6273/backend/ > /dev/null
		curl -s -i -H "X-OmnISci-UseRName: admin" localhost:6273/backend/ > /dev/null
		curl -s -i -H "x-omnisci-username: admin" localhost:6273/backend/ > /dev/null
		docker-compose logs backend | grep -i "Omnisci_Username" > /dev/null && echo "Header Found Test Failed $(ERROR_STRING)" || echo "No Header Found Test Passed $(OK_STRING)"
		