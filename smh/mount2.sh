#!/bin/sh

# --foreground
# --tmp-dir

rm -rf smh/run/filecache-2
rm -f smh/logs/log-2.json


./gcsfuse \
    --config-file smh/conf/config-2.yaml \
    --implicit-dirs \
    --debug_fuse_errors \
    --debug_fuse \
    --debug_fs \
    --debug_gcs \
    --debug_http \
    --debug_http \
    atp-fig11 mnt-2