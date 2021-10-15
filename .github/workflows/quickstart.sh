#!/bin/bash

function cleanup() {
    ARG=$?
    LAST=$previous_command
    echo -e "\n=== Cleanup ==="
    rm gort
    docker-compose down
    echo -e "\n"

    if [ $ARG = 0 ]; then
        echo "Passed!"
    else
        echo "Last command: $LAST"
        echo "Failed!"
    fi

    exit $ARG
}
trap 'previous_command=$this_command; this_command=$BASH_COMMAND' DEBUG
trap cleanup EXIT

# Exit on failure
set -e

# Create your Configuration File
cp config.yml development.yml

# Create a Slack Bot User
# This can be skipped as Gort will run when it cannot connect to Slack
# But could this be mocked in future?

# Build the Gort Image
make image

# Starting Containerized Gort
docker-compose up -d

# Bootstrapping Gort
wait-port -t 10000 4000
go build -o gort
./gort bootstrap --allow-insecure localhost:4000

# Using Gort.
# TODO: Worth having a hidden command to emulate this?
