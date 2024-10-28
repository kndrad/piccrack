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
	go run main.go words --file=$(screenshot_filepath) --save=true -o=$(words_filepath)

words-2:
	go run main.go words --file=./internal/screenshot/testdata/ --save=true -o=$(words_filepath)

frequency:
	go run main.go frequency --file=$(words_filepath)

all:
	./scripts/format.sh
	./scripts/check.sh
	go test ./... -count=1 -failfast -coverprofile=coverage.out
	go run main.go words --file=$(screenshot_filepath) --save=true -o=$(words_filepath)
	go run main.go frequency --file=$(words_filepath)

docker_image = itcrack-dev

docker-build:
	docker build --tag=$(docker_image) .

docker-run:
	docker run $(docker_image)
