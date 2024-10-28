# syntax=docker/dockerfile:1
FROM golang:1.23.2-alpine3.20 AS build

LABEL maintainer="Konrad Nowara"

# Install Tesseract and dependencies
RUN apk add --no-cache \
    gcc \
    musl-dev \
    g++ \
    make \
    pkgconf \
    leptonica-dev \
    tesseract-ocr \
    tesseract-ocr-dev

# Build Go binary
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build -o main ./


# Export Go binary
FROM alpine:3.20.3 AS final-stage
WORKDIR /

RUN apk add --no-cache tesseract-ocr

COPY --from=build /usr/src/app/main /main
ENTRYPOINT ["./main"]
