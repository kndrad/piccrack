# syntax=docker/dockerfile:1
FROM golang:1.23.2-alpine3.20

# Install build dependencies and Tesseract
RUN apk add --no-cache \
    gcc \
    musl-dev \
    g++ \
    make \
    pkgconfig \
    leptonica-dev \
    tesseract-ocr \
    tesseract-ocr-dev

# Continue to build the Go binary
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .
ENTRYPOINT ["./main"]
