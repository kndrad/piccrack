# syntax=docker/dockerfile:1


# Building Go binary stage
FROM golang:1.23.2-alpine3.20 AS build-stage
LABEL maintainer="Konrad Nowara"
WORKDIR /

# Install tesseract and it's dependencies
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

# Build Go binary in /app
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main ./

# Test stage - run cover task from Makefile
FROM build-stage AS test-stage
WORKDIR /app

RUN apk add --no-cache make
COPY --from=build-stage . .
RUN make cover

# Run Go binary
FROM alpine:3.20.3
WORKDIR /

# Once again install tesseract and dependencies to make the Go binary work in this stage
RUN apk add --no-cache \
    tesseract-ocr \
    tesseract-ocr-data-eng \
    leptonica

COPY --from=build-stage /app/main /main
COPY --from=build-stage /app/.env /.env

ENTRYPOINT [ "./main" ]
CMD [ "api", "start" ]
