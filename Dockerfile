# This is a multi-stage Dockerfile. The first part executes a build in a Golang
# container, and the second retrieves the binary from the build container and
# inserts it into a "scratch" image.

# Part 1: Execute the tests in a containerized Golang environment
#
FROM golang:1.12 as test

COPY . /cog2
WORKDIR /cog2
RUN go test -v ./...

# Part 2: Compile the binary in a containerized Golang environment
#
FROM golang:1.12 as builder

COPY . /cog2
WORKDIR /cog2
RUN GOOS=linux go build -a -installsuffix cgo -o cog2 .

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
