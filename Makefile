BINARY_NAME=wordcrack
DOCKER_IMAGE=wordcrack:latest
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
	go test ./... -failfast -coverprofile=coverage.out
	./scripts/check.sh

.PHONY: cover-html
cover-html:
	go test ./... -failfast -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	xdg-open coverage.html

.PHONY: cover
cover:
	go test ./... -failfast -coverprofile=coverage.out

# verbose
.PHONY: cover-v
cover-v:
	go test ./... -v -failfast -coverprofile=coverage.out

.PHONY: tests
tests:
	go test ./... -failfast

API_PKG_PATH=./internal/api/v1

# verbose
.PHONY: api-tests-v
api-tests-v:
	go test $(API_PKG_PATH) -v -failfast -coverprofile=coverage.out

.PHONY: api-tests
api-tests:
	go test $(API_PKG_PATH) -failfast -coverprofile=coverage.out

.PHONY: test-wordcrack-text-file
test-wordcrack-text-file:
	go run main.go text $(SCREENSHOT_FILEPATH) -o $(TXT_FILEPATH)

.PHONY: test-wordcrack-text-dir
test-wordcrack-text-dir:
	go run main.go text -f $(TESTDATA_DIR) -o $(TXT_FILEPATH)

.PHONY: test-wordcrack-frequency
test-wordcrack-frequency:
	go run main.go words frequency --file=$(TESTDATA_DIR)/words.txt

.PHONY: docker-build
docker-build:
	docker build --tag=$(DOCKER_IMAGE) $(DOCKER_IMAGE_PATH)

.PHONY: test-docker-wordcrack-text-1
test-docker-wordcrack-text-1: docker-build
	docker run \
	-u $(shell id -u):$(shell id -g) \
	-e $(ENV_FILEPATH) \
	-v $(TESTDATA_DIR):/testdata \
	-v $(OUTPUT_DIR):/output \
	$(DOCKER_IMAGE) text $(TESTDATA_DIR)/golang_0.png --out=$(OUTPUT_DIR) -v

.PHONY: test-docker-wordcrack-text-2
test-docker-wordcrack-text-2: docker-build
	docker run \
	-u $(shell id -u):$(shell id -g) \
	-e $(ENV_FILEPATH) \
	-v $(TESTDATA_DIR):/testdata \
	-v $(OUTPUT_DIR):/output \
	$(DOCKER_IMAGE) text -v --file=$(TESTDATA_DIR) --out=$(OUTPUT_DIR)

.PHONY: test-docker-test-wordcrack-words-frequency-analyze-file
test-docker-test-wordcrack-words-frequency-analyze-file: docker-build
	docker run \
	-u $(shell id -u):$(shell id -g) \
	-e $(ENV_FILEPATH) \
	-v $(TESTDATA_DIR):/testdata \
	-v $(OUTPUT_DIR):/output \
	$(DOCKER_IMAGE) words frequency analyze -v --file=$(TESTDATA_DIR)/words.txt --out=$(OUTPUT_DIR)

.PHONY: compose-up
compose-up:
	docker-compose up --build -d

.PHONY: compose-down
compose-down:
	docker-compose down

.PHONY: test-wordcrack-words
test-wordcrack-words:
	go run main.go words

.PHONY: clear-output-dir
clear-output-dir:
	rm -rf ./output/*

.PHONY: test-wordcrack-ping
test-wordcrack-ping: compose-up
	go run main.go ping

ANALYSIS_JSON_TEST_FILE=analysis_07_11_2024_07_47_1691.json

.PHONY: test-wordcrack-words-add-many
test-wordcrack-words-add-many: compose-up
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

.PHONY: test-wordcrack-words-frequency
test-wordcrack-words-frequency:
	go run main.go words frequency


.PHONY: docker-start-api
docker-start-api:
	docker-compose up --build -d
	go run main.go api start
