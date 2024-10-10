fmt:
	./scripts/format.sh

review: fmt
	./scripts/check.sh

main:
	go run ./cmd/main.go

words-1:
	go run main.go words --file=./internal/screenshot/testdata/golang_0.png --save=true -o=./internal/screenshot/testdata/out-file.txt

words-2:
	go run main.go words --file=./internal/screenshot/testdata/ --save=true -o=./internal/screenshot/testdata/out-dir.txt

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

