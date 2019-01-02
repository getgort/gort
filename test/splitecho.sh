#!/usr/bin/env bash

echo "Found $# parameters:"

{
    for p in $@; do
        echo $p
        sleep 1
    done
}
