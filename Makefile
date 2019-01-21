.DEFAULT_GOAL := help

############################
## Project Info
############################
PROJECT = cog2
GIT_URL = github.com
GIT_ORGANIZATION = clockworksoul
GIT_REPOSITORY = $(GIT_URL)/$(GIT_ORGANIZATION)/$(PROJECT)

############################
## Docker Registry Info
############################
REGISTRY_URL = clockworksoul
IMAGE_NAME = $(REGISTRY_URL)/$(PROJECT)
IMAGE_TAG = latest

help:
	# Commands:
	# make help           - Show this message
	#
	# Dev commands:
	# make clean          - Remove generated files
	# make install        - Install requirements (none atm)
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

install:
	@echo TODO Start using Go modules!
	@go get github.com/docker/docker/api
	@go get github.com/docker/docker/client
	@go get github.com/gorilla/mux
	@go get github.com/nlopes/slack
	@go get github.com/sirupsen/logrus
	@go get github.com/spf13/cobra
	@go get golang.org/x/net/context
	@go get gopkg.in/yaml.v1
	@go get gopkg.in/yaml.v2

test_begin:
	@docker stop foo_postgres | true
	@docker rm foo_postgres | true
	@docker run -d -e POSTGRES_USER=cog -e POSTGRES_PASSWORD=password -p 5432:5432 --name foo_postgres postgres:10
	
test: test_begin
	@docker build --target test -t foo_$(PROJECT)_foo --network host .
	@docker rmi foo_$(PROJECT)_foo
	@docker stop foo_postgres
	@docker rm foo_postgres

build: clean
	mkdir -p bin
	@go build -a -installsuffix cgo -o bin/$(PROJECT) $(GIT_REPOSITORY)

image: test_begin
	@docker build --target image -t $(IMAGE_NAME):$(IMAGE_TAG) --network host .
	@docker stop foo_postgres
	@docker rm foo_postgres

push:
	@docker push $(IMAGE_NAME):$(IMAGE_TAG)
