review:
	./scripts/format.sh
	./scripts/check.sh

run: review
	go run ./cmd/main.go

cover:
	go test ./... -count=1 -failfast -coverprofile=coverage.out
	go tool cover -html=coverage.out

tests:
	go test ./... -count=1 -failfast
