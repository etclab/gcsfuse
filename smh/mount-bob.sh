#!/bin/sh

# --foreground
# --tmp-dir

rm -rf smh/run/filecache
rm -f smh/logs/log.json


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
