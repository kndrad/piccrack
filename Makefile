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

words-file:
	go run main.go words --file=./internal/screenshot/testdata/golang_0.png --save=true -o=./internal/screenshot/testdata/words_out.txt

words-dir:
	go run main.go words --file=./internal/screenshot/testdata/ --save=true -o=./internal/screenshot/testdata/words_out.txt

text-analysis:
	go run main.go frequency --file=./internal/screenshot/testdata/words_out.txt

all:
	./scripts/format.sh
	./scripts/check.sh
	go test ./... -count=1 -failfast -coverprofile=coverage.out
	go run main.go words --file=./internal/screenshot/testdata/golang_0.png --save=true -o=./internal/screenshot/testdata/words_out.txt
	go run main.go frequency --file=./internal/screenshot/testdata/words_out.txt
