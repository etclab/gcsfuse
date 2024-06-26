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
    --akeso_strategy akeso \
    --akeso_dir smh/akeso.d \
    --akeso_project ornate-flame-397517 \
    --akeso_topic smh-akeso-strawman \
    --akeso_sub smh-akeso-strawman-sub1 \
    wmsr-test-bucket2 mnt
