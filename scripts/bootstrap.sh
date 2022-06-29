#!/bin/bash

wait-for-url() {
    echo "Testing $1"
    timeout -s TERM 45 bash -c \
    'while [[ "$(curl --insecure -s -o /dev/null -L -w ''%{http_code}'' ${0})" != "200" ]];\
    do echo "Waiting for ${0}" && sleep 2;\
    done' ${1}
    echo "OK!"
    curl -I $1
}

wait-for-url https://localhost:4000/v2/healthz

go run . bootstrap -F -B ./scratch https://localhost:4000 --allow-insecure
