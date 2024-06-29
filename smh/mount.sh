#!/bin/sh

# --foreground
# --tmp-dir

rm -rf smh/run/filecache
rm -f smh/logs/log.json

rm -rf smh/run/filecache-cici
rm -f smh/logs/log-cici.json


./gcsfuse \
    --config-file smh/conf/config.yaml \
    --implicit-dirs \
    --debug_fuse_errors \
    --debug_fuse \
    --debug_fs \
    --debug_gcs \
    --debug_http \
    --debug_http \
    wmsr-test-bucket2 mnt \
&& ./gcsfuse \
    --config-file smh/conf/config-cici.yaml \
    --implicit-dirs \
    --debug_fuse_errors \
    --debug_fuse \
    --debug_fs \
    --debug_gcs \
    --debug_http \
    --debug_http \
    wmsr-test-bucket2 mnt-cici
