BINARY_NAME=piccrack
DOCKER_IMAGE=piccrack:latest
DOCKER_COMPOSE_IMAGE=w-app:latest
DOCKER_IMAGE_PATH=.

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

.PHONY: compose-up
compose-up:
	docker-compose up --build

.PHONY: compose-down
compose-down:
	docker-compose down

# Stop development
.PHONY: stop-all
stop-all:
	docker-compose down
	./scripts/format.sh
	go test ./... -failfast -coverprofile=coverage.out
	./scripts/check.sh
