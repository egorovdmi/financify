SHELL := /bin/bash
VERSION := 1.0

default:
	go build -o bin/financify-api ./app/financify-api/main.go

all: financify

financify:
	docker build \
		-f zarf/docker/dockerfile.financify-api \
		-t financify-api-amd64:$(VERSION) \
		--build-arg BUILD_REF=`git rev-parse --short HEAD` \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

tidy:
	go mod tidy
	go mod vendor

run:
	go run ./app/financify-api/main.go

runk:
	go run ./app/keygen/main.go

test:
# -count=1 means, don't use the cache 
	go test -v ./... -count=1
