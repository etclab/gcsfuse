#!/bin/sh

# these are files that I have modified
files="cmd/flags.go
cmd/legacy_main.go
cmd/mount.go
internal/akeso/config.go
internal/akeso/metadata.go
internal/akeso/pubsub.go
internal/cache/file/cache_handle.go
internal/cache/file/downloader/downloader.go
internal/cache/file/downloader/job.go
internal/fs/fs.go
internal/fs/handle/file.go
internal/fs/inode/base_dir.go
internal/fs/inode/file.go
internal/gcsx/bucket_manager.go
internal/gcsx/random_reader.go
internal/gcsx/syncer.go
internal/storage/bucket_handle.go
internal/storage/storage_handle.go"

for f in $files; do
    go fmt "$f"
done

#go vet ./...

go build
