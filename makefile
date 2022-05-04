SHELL := /bin/bash

default:
	go build -o bin/financify-api ./app/financify-api/main.go

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
