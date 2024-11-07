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

.PHONY: fmt
fmt:
	./scripts/format.sh

.PHONY: review
review: fmt
	./scripts/check.sh
	go test ./... -count=1 -failfast -coverprofile=coverage.out

.PHONY: cover-html
cover-html:
	go test ./... -count=1 -failfast -coverprofile=coverage.out
	go tool cover -html=coverage.out

.PHONY: cover
cover:
	go test ./... -count=1 -failfast -coverprofile=coverage.out

.PHONY: tests
tests:
	go test ./... -count=1 -failfast

.PHONY: run
run:
	go run ./cmd/main.go

.PHONY: itcrack-text-file
itcrack-text-file:
	go run main.go text --file=$(SCREENSHOT_FILEPATH) -o=$(TXT_FILEPATH)

.PHONY: itcrack-text-dir
itcrack-text-dir:
	go run main.go text --file=$(TESTDATA_DIR) -o=$(TXT_FILEPATH)

.PHONY: itcrack-frequency
itcrack-frequency:
	go run main.go words frequency --file=$(TESTDATA_DIR)/words.txt

.PHONY: docker-build
docker-build:
	docker build --tag=$(DOCKER_IMAGE) $(DOCKER_IMAGE_PATH)

.PHONY: docker-itcrack-text-1
docker-itcrack-text-1: docker-build
	docker run \
	-u $(shell id -u):$(shell id -g) \
	-e $(ENV_FILEPATH) \
	-v $(TESTDATA_DIR):/testdata \
	-v $(OUTPUT_DIR):/output \
	$(DOCKER_IMAGE) text -v --file=$(TESTDATA_DIR)/golang_0.png --out=$(OUTPUT_DIR)

.PHONY: docker-itcrack-text-2
docker-itcrack-text-2: docker-build
	docker run \
	-u $(shell id -u):$(shell id -g) \
	-e $(ENV_FILEPATH) \
	-v $(TESTDATA_DIR):/testdata \
	-v $(OUTPUT_DIR):/output \
	$(DOCKER_IMAGE) text -v --file=$(TESTDATA_DIR) --out=$(OUTPUT_DIR)

.PHONY: docker-itcrack-words-freq-file
docker-itcrack-words-freq-file: docker-build
	docker run \
	-u $(shell id -u):$(shell id -g) \
	-e $(ENV_FILEPATH) \
	-v $(TESTDATA_DIR):/testdata \
	-v $(OUTPUT_DIR):/output \
	$(DOCKER_IMAGE) words frequency -v --file=$(TESTDATA_DIR)/words.txt --out=$(OUTPUT_DIR)

.PHONY: compose-up
compose-up:
	docker-compose up --build -d

.PHONY: compose-down
compose-down:
	docker-compose down

.PHONY: ping
ping:
	go run main.go ping

.PHONY: itcrack-words
itcrack-words:
	go run main.go words

.PHONY: clear-output-dir
clear-output-dir:
	rm -rf ./output/*

.PHONY: itcrack-ping
itcrack-ping: compose-up
	go run main.go ping
