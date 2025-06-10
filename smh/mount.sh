#!/bin/bash

PWD=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

rm -rf smh/run/filecache
rm -f smh/logs/log.json

mkdir -p $PWD/../mnt
mkdir -p smh/logs/
mkdir -p smh/run/filecache/

./gcsfuse \
    --config-file smh/conf/config.yaml \
    --implicit-dirs \
    --debug_fuse_errors \
    --debug_fuse \
    --debug_fs \
    --debug_gcs \
    --debug_http \
    --debug_http \
   atp-master mnt