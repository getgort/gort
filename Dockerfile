# This is a multi-stage Dockerfile. The first part executes a build in a Golang
# container, and the second retrieves the binary from the build container and
# inserts it into a "scratch" image.

# Part 1: Execute the tests in a containerized Golang environment
#
FROM golang:1.11 as test

WORKDIR /go/bin/

RUN rm -Rf /go/src 

COPY . /go/src/github.com/clockworksoul/cog2

# TODO Start using Go modules!
RUN go get github.com/docker/docker/api \
    && go get github.com/docker/docker/client \
    && go get github.com/gorilla/mux \
    && go get github.com/nlopes/slack \
    && go get github.com/sirupsen/logrus\
    && go get github.com/spf13/cobra \
    && go get golang.org/x/net/context \
    && go get gopkg.in/yaml.v1 \
    && go get gopkg.in/yaml.v2

RUN go test -v github.com/clockworksoul/cog2/...

# Part 2: Compile the binary in a containerized Golang environment
#
FROM golang:1.11 as builder

WORKDIR /go/bin/

RUN rm -Rf /go/src 

COPY . /go/src/github.com/clockworksoul/cog2

# TODO Start using Go modules!
RUN go get github.com/docker/docker/api \
    && go get github.com/docker/docker/client \
    && go get github.com/gorilla/mux \
    && go get github.com/nlopes/slack \
    && go get github.com/sirupsen/logrus\
    && go get github.com/spf13/cobra \
    && go get golang.org/x/net/context \
    && go get gopkg.in/yaml.v1 \
    && go get gopkg.in/yaml.v2

RUN GOOS=linux go build -a -installsuffix cgo -o cog2 github.com/clockworksoul/cog2

# Part 3: Build the Cog2 image proper
#
FROM ubuntu:16.04 as image

# Install Ansible
#
RUN apt update                                              \
  && apt-get -y --force-yes install --no-install-recommends \
    ssh                                                     \
    ca-certificates                                         \
  && apt-get clean                                          \
  && apt-get autoclean                                      \
  && apt-get autoremove                                     \
  && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN ssh-keygen -b 2048 -f /root/.ssh/id_rsa -P ''

COPY --from=builder /go/bin/cog2 .

EXPOSE 4000

CMD [ "/cog2", "start", "-v" ]
