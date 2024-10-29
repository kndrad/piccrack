# syntax=docker/dockerfile:1

FROM golang:1.23.2-alpine3.20 AS build-stage
LABEL maintainer="Konrad Nowara"
WORKDIR /

# Install tesseract and dependencies for the Go binary
RUN apk add --no-cache \
    gcc \
    musl-dev \
    g++ \
    make \
    pkgconf \
    leptonica-dev \
    tesseract-ocr \
    tesseract-ocr-dev \
    tesseract-ocr-data-eng

# Build Go binary
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build -o main ./

# Export Go binary
FROM alpine:3.20.3 AS final-stage
WORKDIR /

# Once again install tesseract and dependencies to make the Go binary work
RUN apk add --no-cache \
    tesseract-ocr \
    tesseract-ocr-data-eng \
    leptonica

COPY --from=build-stage /app/main /main

ENTRYPOINT ["./main"]
