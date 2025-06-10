#!/bin/bash

set -e

PREFIX="atp"
STRATEGY="cmek"

PWD=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

setup_cmek=false
benchmark=false

REPO=${PWD}

for cmd in "$@"; do
    case $cmd in
        setup_cmek) setup_cmek=true ;; 
        benchmark) benchmark=true ;;
        *) 
            echo "Unknown command: $cmd"
            ;;
    esac
done

if [[ "$setup_cmek" == "true" ]]; then

    echo "Setting up folders and configs"

    mkdir -p ${REPO}/smh/logs
    mkdir -p ${REPO}/smh/run
    mkdir -p ${REPO}/mnt

    # no need for cache/logging/debug during perf
    cat ${REPO}/cmek/new-config.yaml > ${REPO}/smh/conf/config.yaml
    cat ${REPO}/cmek/new-mount.sh > ${REPO}/smh/mount.sh

    cp ${REPO}/cmek/run-bench.py ${REPO}

    cd ${REPO}
    ${REPO}/smh/make.sh
    cd -
fi

if [[ "$benchmark" == "true" ]]; then
    echo "Running benchmark"

    cd ${PWD}
    cd ${REPO}
    # python3 run-bench.py <sets> <reps>
    python3 run-bench.py 1 5
    cd -
fi

