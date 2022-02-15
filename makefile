SHELL := /bin/bash

default:
	go build -o bin/financify-api ./app/financify-api/main.go

tidy:
	go mod tidy
	go mod vendor

run:
	go run ./app/financify-api/main.go
