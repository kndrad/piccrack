# Binary and build configuration
BINARY_NAME=itcrack
DOCKER_IMAGE=itcrack-dev
DOCKER_IMAGE_PATH=.

# File paths
WORDS_FILEPATH=./internal/textproc/testdata/words.txt
SCREENSHOT_FILEPATH=./internal/screenshot/testdata/golang_0.png
SCREENSHOT_TESTDATA_FILEPATH=./internal/screenshot/testdata/
SCREENSHOTS_DIR=screenshots
OUTPUT_DIR=output

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

.PHONY: staging
staging: fmt review cover

.PHONY: run
run:
	go run ./cmd/main.go

.PHONY: text-1
text-1:
	go run main.go text --file=$(SCREENSHOT_FILEPATH) -o=$(WORDS_FILEPATH)

.PHONY: text-2
text-2:
	go run main.go text --file=$(SCREENSHOT_TESTDATA_FILEPATH) -o=$(WORDS_FILEPATH)

.PHONY: frequency
frequency:
	go run main.go frequency --file=$(WORDS_FILEPATH)

.PHONY: all
all: fmt review cover text-1 frequency

.PHONY: docker-build
docker-build:
	docker build --tag=$(DOCKER_IMAGE) $(DOCKER_IMAGE_PATH)

.PHONY: docker-test-1
docker-test-1:
	docker run -v $(shell pwd)/screenshots:/screenshots -v $(shell pwd)/output:/app/output $(DOCKER_IMAGE):latest words --file=/screenshots/golang_0.png --out=app/output

.PHONY: docker-test-2
docker-test-2:
	docker run -v $(shell pwd)/$(SCREENSHOTS_DIR):/$(SCREENSHOTS_DIR) $(DOCKER_IMAGE):latest words --file=$(SCREENSHOTS_DIR)

.PHONY: docker-all
docker-all: docker-build docker-test-1 docker-test-2

.PHONY: compose-up
compose-up:
	docker-compose up --build -d

.PHONY: compose-down
compose-down:
	docker-compose down

.PHONY: ping-db
ping-db:
	go run main.go ping
