# This is a multi-stage Dockerfile.
# The first part retrieves Go dependencies, the second compiles the binary in
# a Go container, and the second retrieves the binary from the build container and
# inserts it into a "scratch" image.
# An additional image is included for executing tests.

# Part 1: Create a layer for Go module dependencies
#
FROM golang:1.17 as gomodules

WORKDIR /gort
COPY go.mod go.sum /gort/
RUN go mod download

# Part 1a: Execute quick tests in a containerized Golang environment
#
FROM gomodules as test

WORKDIR /gort
COPY . /gort
RUN --mount=type=cache,target=/root/.cache/go-build \
  go test -v -short ./...

# Part 3: Compile the binary in a containerized Golang environment
#
FROM gomodules as builder

COPY . /gort
WORKDIR /gort
RUN --mount=type=cache,target=/root/.cache/go-build \
  GOOS=linux go build -a -installsuffix cgo -o gort .

# Part 4: Build the Gort image proper
#
FROM ubuntu:20.04 as image

# Install Ansible
#
RUN apt update                                              \
  && apt-get -y --force-yes install --no-install-recommends \
  ssh                                                       \
  ca-certificates                                           \
  && apt-get clean                                          \
  && apt-get autoclean                                      \
  && apt-get autoremove                                     \
  && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN ssh-keygen -b 2048 -f /root/.ssh/id_rsa -P ''

COPY --from=builder /gort/gort /bin

ENTRYPOINT [ "/bin/gort" ]

EXPOSE 4000

CMD [ "start", "-v" ]
