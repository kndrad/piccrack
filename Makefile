fmt:
	./scripts/format.sh

review: fmt
	./scripts/check.sh

cover-html:
	go test ./... -count=1 -failfast -coverprofile=coverage.out
	go tool cover -html=coverage.out

cover:
	go test ./... -count=1 -failfast -coverprofile=coverage.out

tests:
	go test ./... -count=1 -failfast

staging:
	./scripts/format.sh
	./scripts/check.sh
	go test ./... -count=1 -failfast -coverprofile=coverage.out

main:
	go run ./cmd/main.go

words_filepath = ./internal/screenshot/testdata/words.txt
screenshot_filepath = ./internal/screenshot/testdata/golang_0.png
screenshot_testdata_filepath = ./internal/screenshot/testdata/

words-1:
	go run main.go words --file=$(screenshot_filepath)-o=$(words_filepath)

words-2:
	go run main.go words --file=./internal/screenshot/testdata/ -o=$(words_filepath)

frequency:
	go run main.go frequency --file=$(words_filepath)

all:
	./scripts/format.sh
	./scripts/check.sh
	go test ./... -count=1 -failfast -coverprofile=coverage.out
	go run main.go words --file=$(screenshot_filepath) -o=$(words_filepath)
	go run main.go frequency --file=$(words_filepath)

docker_image = itcrack-dev

docker-build:
	docker build --tag=$(docker_image) .

screenshots_dir = screenshots
output_dir = output

docker-test-1:
	docker run -v $(shell pwd)/screenshots:/screenshots -v $(shell pwd)/output:/app/output itcrack-dev:latest words --file=/screenshots/golang_0.png --out=app/output

docker-test-2:
	docker run -v $(shell pwd)/$(screenshots_dir):/$(screenshots_dir) itcrack-dev:latest words --file=$(screenshots_dir)

docker-all:
	docker build -t $(docker_image) .
	docker run -v $(shell pwd)/screenshots:/screenshots -v $(shell pwd)/output:/app/output itcrack-dev:latest words --file=/screenshots/golang_0.png --out=app/output
	docker run -v $(shell pwd)/$(screenshots_dir):/$(screenshots_dir) -v $(shell pwd)/output:/app/output itcrack-dev:latest words --file=$(screenshots_dir) --out=app/output

