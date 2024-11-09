BINARY_NAME=itcrack
DOCKER_IMAGE=itcrack:latest
DOCKER_COMPOSE_IMAGE=itcrack-app:latest
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
	go test ./... -count=1 -failfast -coverprofile=coverage.out
	./scripts/check.sh

.PHONY: go-cover-html
go-cover-html:
	go test ./... -count=1 -failfast -coverprofile=coverage.out
	go tool cover -html=coverage.out

.PHONY: go-cover
go-cover:
	go test ./... -count=1 -failfast -coverprofile=coverage.out

.PHONY: go-tests
go-tests:
	go test ./... -count=1 -failfast

.PHONY: test-itcrack-text-file
test-itcrack-text-file:
	go run main.go text $(SCREENSHOT_FILEPATH) -o $(TXT_FILEPATH)

.PHONY: test-itcrack-text-dir
test-itcrack-text-dir:
	go run main.go text -f $(TESTDATA_DIR) -o $(TXT_FILEPATH)

.PHONY: test-itcrack-frequency
test-itcrack-frequency:
	go run main.go words frequency --file=$(TESTDATA_DIR)/words.txt

.PHONY: docker-build
docker-build:
	docker build --tag=$(DOCKER_IMAGE) $(DOCKER_IMAGE_PATH)

.PHONY: test-docker-itcrack-text-1
test-docker-itcrack-text-1: docker-build
	docker run \
	-u $(shell id -u):$(shell id -g) \
	-e $(ENV_FILEPATH) \
	-v $(TESTDATA_DIR):/testdata \
	-v $(OUTPUT_DIR):/output \
	$(DOCKER_IMAGE) text -v --file=$(TESTDATA_DIR)/golang_0.png --out=$(OUTPUT_DIR)

.PHONY: test-docker-itcrack-text-2
test-docker-itcrack-text-2: docker-build
	docker run \
	-u $(shell id -u):$(shell id -g) \
	-e $(ENV_FILEPATH) \
	-v $(TESTDATA_DIR):/testdata \
	-v $(OUTPUT_DIR):/output \
	$(DOCKER_IMAGE) text -v --file=$(TESTDATA_DIR) --out=$(OUTPUT_DIR)

.PHONY: test-docker-test-itcrack-words-frequency-analyze-file
test-docker-test-itcrack-words-frequency-analyze-file: docker-build
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

.PHONY: test-itcrack-words
test-itcrack-words:
	go run main.go words

.PHONY: clear-output-dir
clear-output-dir:
	rm -rf ./output/*

.PHONY: test-itcrack-ping
test-itcrack-ping: compose-up
	go run main.go ping

ANALYSIS_JSON_TEST_FILE=analysis_07_11_2024_07_47_1691.json

.PHONY: test-itcrack-words-add-many
test-itcrack-words-add-many: compose-up
	go run main.go words add many '$(TESTDATA_DIR)/$(ANALYSIS_JSON_TEST_FILE)'


# Stop development
.PHONY: stop-all
stop-all:
	docker-compose down
	./scripts/format.sh
	go test ./... -count=1 -failfast -coverprofile=coverage.out
	./scripts/check.sh

# Start development
.PHONY: start
start:
	docker-compose up --build -d
	go run main.go ping

.PHONY: test-itcrack-words-frequency
test-itcrack-words-frequency:
	go run main.go words frequency
