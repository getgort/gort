.DEFAULT_GOAL := help

############################
## Project Info
############################
PROJECT = gort
GIT_URL = github.com
GIT_ORGANIZATION = getgort
GIT_REPOSITORY = $(GIT_URL)/$(GIT_ORGANIZATION)/$(PROJECT)

############################
## Docker Registry Info
############################
REGISTRY_URL = getgort
IMAGE_NAME = $(REGISTRY_URL)/$(PROJECT)
IMAGE_TAG = $(shell grep "Version =" version/version.go | sed 's/.*Version = "\(.*\)"/\1/')

help:
	# Commands:
	# make help           - Show this message
	#
	# Dev commands:
	# make clean          - Remove generated files
	# make test           - Run Go tests
	# make build          - Build go binary
	#
	# Docker commands:
	# make image          - Build Docker image with current version tag
	# make run            - Run Docker image with current version tag
	#
	# Deployment commands:
	# make push           - Push current version tag to registry

clean:
	if [ -d "bin" ]; then rm -R bin; fi
	rm coverage.out coverage.html

test:
	@DOCKER_BUILDKIT=1 docker build --target test -t foo_$(PROJECT)_foo --network host .
	@docker rmi foo_$(PROJECT)_foo

test-local:
	@go test -count=1 -timeout 60s -cover -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

build: clean
	mkdir -p bin
	@go build -a -installsuffix cgo -o bin/$(PROJECT) $(GIT_REPOSITORY)

image: test
	@echo Building image $(IMAGE_NAME):$(IMAGE_TAG)
	@DOCKER_BUILDKIT=1 docker build --target image -t $(IMAGE_NAME):latest --network host .
	@docker tag $(IMAGE_NAME):latest $(IMAGE_NAME):$(IMAGE_TAG)

push: image
	@docker push $(IMAGE_NAME):$(IMAGE_TAG)
