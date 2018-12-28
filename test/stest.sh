#!/usr/bin/env bash

echo "Found $# parameters:"

{
    for p in $@; do
        echo "STDOUT: $p"
        sleep 1
    done
} &

{
    for p in $@; do
        echo "STDERR: $p" 1>&2
        sleep 1
    done
}
