./smh/umount.sh

rm -rf ./smh/akeso.d/abcd && rm -rf ./smh/run/filecache

./gcsfuse --config-file smh/conf/akeso-nested.yaml --implicit-dirs atp-nested-ca mnt