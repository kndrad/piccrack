BINARY_NAME=wcrack
DOCKER_IMAGE=wcrack:latest
DOCKER_COMPOSE_IMAGE=wordcrack-app:latest
DOCKER_IMAGE_PATH=.

TXT_FILEPATH=./internal/textproc/testdata/words.txt
SCREENSHOT_FILEPATH=./testdata/golang_0.png
TESTDATA_DIR=./testdata
OUTPUT_DIR=./output
ENV_FILEPATH=$(shell pwd)/.env

.PHONY: build
build:
	go build -o bin/$(BINARY_NAME) ./cmd/main.go

.PHONY: format
format:
	./scripts/format.sh

.PHONY: review
review:
	./scripts/format.sh
	go clean -testcache
	go test ./... -cover -coverprofile=coverage.out
	./scripts/check.sh

.PHONY: cover-html
cover-html:
	go test ./... -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	xdg-open coverage.html

.PHONY: cover
cover:
	go test ./... -cover -coverprofile=coverage.out

API_PKG_PATH=./internal/api/v1

# verbose
.PHONY: api-tests-v
api-tests-v:
	go test $(API_PKG_PATH) -v -failfast -coverprofile=coverage.out

.PHONY: api-tests
api-tests:
	go test $(API_PKG_PATH) -failfast -coverprofile=coverage.out

.PHONY: test-wcrack-text-file
test-wcrack-text-file:
	go run main.go text $(SCREENSHOT_FILEPATH) -o $(TXT_FILEPATH)

.PHONY: test-wcrack-text-dir
test-wcrack-text-dir:
	go run main.go text -f $(TESTDATA_DIR) -o $(TXT_FILEPATH)

.PHONY: test-wcrack-frequency
test-wcrack-frequency:
	go run main.go words frequency --file=$(TESTDATA_DIR)/words.txt

.PHONY: docker-build
docker-build:
	docker build --tag=$(DOCKER_IMAGE) $(DOCKER_IMAGE_PATH)

.PHONY: test-docker-wcrack-text-1
test-docker-wcrack-text-1: docker-build
	docker run \
	-u $(shell id -u):$(shell id -g) \
	-e $(ENV_FILEPATH) \
	-v $(TESTDATA_DIR):/testdata \
	-v $(OUTPUT_DIR):/output \
	$(DOCKER_IMAGE) text $(TESTDATA_DIR)/golang_0.png --out=$(OUTPUT_DIR) -v

.PHONY: test-docker-wcrack-text-2
test-docker-wcrack-text-2: docker-build
	docker run \
	-u $(shell id -u):$(shell id -g) \
	-e $(ENV_FILEPATH) \
	-v $(TESTDATA_DIR):/testdata \
	-v $(OUTPUT_DIR):/output \
	$(DOCKER_IMAGE) text -v --file=$(TESTDATA_DIR) --out=$(OUTPUT_DIR)

.PHONY: test-docker-test-wcrack-words-frequency-analyze-file
test-docker-test-wcrack-words-frequency-analyze-file: docker-build
	docker run \
	-u $(shell id -u):$(shell id -g) \
	-e $(ENV_FILEPATH) \
	-v $(TESTDATA_DIR):/testdata \
	-v $(OUTPUT_DIR):/output \
	$(DOCKER_IMAGE) words frequency analyze -v --file=$(TESTDATA_DIR)/words.txt --out=$(OUTPUT_DIR)

.PHONY: compose-up-d
compose-up-d:
	docker-compose up --build -d

.PHONY: compose-up
compose-up:
	docker-compose up --build

.PHONY: compose-down
compose-down:
	docker-compose down

.PHONY: test-wcrack-words
test-wcrack-words:
	go run main.go words

.PHONY: clear-output-dir
clear-output-dir:
	rm -rf ./output/*

.PHONY: test-wcrack-ping
test-wcrack-ping: compose-up
	go run main.go ping

ANALYSIS_JSON_TEST_FILE=analysis_07_11_2024_07_47_1691.json

.PHONY: test-wcrack-words-add-many
test-wcrack-words-add-many: compose-up
	go run main.go words add many '$(TESTDATA_DIR)/$(ANALYSIS_JSON_TEST_FILE)'


# Stop development
.PHONY: stop-all
stop-all:
	docker-compose down
	./scripts/format.sh
	go test ./... -failfast -coverprofile=coverage.out
	./scripts/check.sh

# Start development
.PHONY: start
start:
	docker-compose up --build -d
	go run main.go ping

.PHONY: test-wcrack-words-frequency
test-wcrack-words-frequency:
	go run main.go words frequency

