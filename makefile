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

deploy: financify
	kustomize build ./zarf/k8s/dev | kubectl apply -f -

remove:
	kustomize build ./zarf/k8s/dev | kubectl delete -f -

upgrade: financify
	kubectl delete pods -lapp=fin

status:
	kubectl get nodes
	kubectl describe pod -lapp=fin
	kubectl get service

logs:
	kubectl logs -lapp=fin --all-containers=true -f

tidy:
	go mod tidy
	go mod vendor

run:
	go run ./app/financify-api/main.go

runk:
	go run ./app/keygen/main.go

runa:
	go run ./app/admin/main.go

test:
# -count=1 means, don't use the cache
	go test -v ./... -count=1
