fmt:
	./scripts/format.sh

review: fmt
	./scripts/check.sh

run: review
	go run ./cmd/main.go

cover-html:
	go test ./... -count=1 -failfast -coverprofile=coverage.out
	go tool cover -html=coverage.out

cover:
	go test ./... -count=1 -failfast -coverprofile=coverage.out

tests:
	go test ./... -count=1 -failfast
