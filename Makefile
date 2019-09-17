.DEFAULT_GOAL := help

.PHONY: build

all: build

build: ## Compile Go code and create zip file to upload to AWS Lambda
	env GOOS=linux GOARCH=amd64 go build -o handler main.go
	zip -j handler.zip handler

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'