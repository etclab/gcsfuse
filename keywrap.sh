#!/bin/bash

set -e

PREFIX="atp"
STRATEGY="keywrap"

PWD=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

setup_keywrap=false
benchmark=false

REPO=${PWD}

for cmd in "$@"; do
    case $cmd in
        setup_keywrap) setup_keywrap=true ;; 
        benchmark) benchmark=true ;;
        *) 
            echo "Unknown command: $cmd"
            ;;
    esac
done

if [[ "$setup_keywrap" == "true" ]]; then

    echo "Setting up folders and configs"

    mkdir -p ${REPO}/smh/logs
    mkdir -p ${REPO}/smh/run
    mkdir -p ${REPO}/mnt

    # no need for cache/logging/debug during perf
    cat ${REPO}/keywrap/new-config.yaml > ${REPO}/smh/conf/config.yaml
    cat ${REPO}/keywrap/new-mount.sh > ${REPO}/smh/mount.sh

    cp ${REPO}/keywrap/run-bench.py ${REPO}

    sed -i "s/GCP_PROJECT_ID/${GCP_PROJECT_ID}/g" ${REPO}/smh/conf/config.yaml

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
