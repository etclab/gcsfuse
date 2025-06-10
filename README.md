## cmek and cmek-hsm `gcsfuse` strategies
Please look into the paper for details about `cmek` and `cmek-hsm` client-side strategies.

Below we outline the dependencies, build and run instructions for the `cmek` branch.

- Dependencies
    - If you already installed the dependencies for one of the strategies, skip this step.
    - Running `gcsfuse` requires `go` and `fuse3`
    - The required packages can be installed using the command below (note: please skip `./common/install-go.sh` if you already have `Go` installed - as it'll replace the `Go` on your path, and `./common/install-gcloud.sh` if you already have gcloud cli installed):
        ```bash
        ./common/install-dependencies.sh && ./common/install-go.sh && ./common/install-gcloud.sh && source ~/.bashrc
        ``` 

- Config Setup
    - This sets up the required configs and builds the binary
        ```bash
        # for cmek-hsm run: ./cmek-hsm.sh setup_cmek_hsm
        ./cmek.sh setup_cmek
        ```

- Run benchmarks
    - Setup Service Account (SA) credentials to access the `atp-cmek` and `atp-cmek-hsm` bucket. This SA credentials has been configured with read and write access to the buckets. Also, note the software and hsm keys has been configured with the two buckets and the SA credentials has access to it.
        ```bash
        # adjust the service account key path accordingly
        export GOOGLE_APPLICATION_CREDENTIALS=$HOME/downloads/serviceAccount-ae-pets25-alice.json
        gcloud auth activate-service-account --key-file=$GOOGLE_APPLICATION_CREDENTIALS
        ```

    - Run the benchmarks (runs the benchmarks: `read_full_file` and `write_to_gcs`) five times
        ```bash
        # for cmek-hsm run: ./cmek-hsm.sh benchmark
        ./cmek.sh benchmark
        ```


<details>
  <summary>Original gcsfuse Readme</summary>

[![codecov](https://codecov.io/gh/GoogleCloudPlatform/gcsfuse/graph/badge.svg?token=vNsbSbeea2)](https://codecov.io/gh/GoogleCloudPlatform/gcsfuse)

# Current status

Starting with V1.0, Cloud Storage FUSE is Generally Available and supported by Google, provided that it is used within its documented supported applications, platforms, and limits. Support requests, feature requests, and general questions should be submitted as a support request via Google Cloud support channels or via GitHub [here](https://github.com/GoogleCloudPlatform/gcsfuse/issues).

Cloud Storage FUSE is open source software, released under the 
[Apache license](https://github.com/GoogleCloudPlatform/gcsfuse/blob/master/LICENSE).

## _New_ Cloud Storage FUSE V2
Cloud Storage FUSE V2 provides important stability, functionality, and performance enhancements, including the introduction of a file cache that allows repeat file reads to be served from a local, faster cache storage of choice, such as a Local SSD, Persistent Disk, or even in-memory /tmpfs. The Cloud Storage FUSE file cache makes AI/ML training faster and more cost-effective by reducing the time spent waiting for data, with up to _**2.3x faster training time and 3.4x higher throughput**_ observed in training runs. This is especially valuable for multi epoch training and can serve small and random I/O operations significantly faster. The file cache feature is disabled by default and is enabled by passing a directory to 'cache-dir'. See [overview of caching](https://cloud.google.com/storage/docs/gcsfuse-cache) for more details. 

# ABOUT
## What is Cloud Storage FUSE?

Cloud Storage FUSE is an open source FUSE adapter that lets you mount and access Cloud Storage buckets as local file systems. For a technical overview of Cloud Storage FUSE, see https://cloud.google.com/storage/docs/gcs-fuse.

## Cloud Storage FUSE for machine learning

To learn about the benefits of using Cloud Storage FUSE for machine learning projects, see https://cloud.google.com/storage/docs/gcsfuse-integrations#machine-learning.

## Limitations and key differences from POSIX file systems

To learn about limitations and differences between Cloud Storage FUSE and POSIX file systems, see https://cloud.google.com/storage/docs/gcs-fuse#differences-and-limitations.

## Pricing for Cloud Storage FUSE

For information about pricing for Cloud Storage FUSE, see https://cloud.google.com/storage/docs/gcs-fuse#charges.

# CSI Driver

Using the [Cloud Storage FUSE CSI driver](https://github.com/GoogleCloudPlatform/gcs-fuse-csi-driver), users get the declarative nature of Kubernetes
with all infrastructure fully managed by GKE in combination with Cloud Storage. This CSI
driver relies on Cloud Storage FUSE to mount Cloud storage buckets as file systems on the
GKE nodes, with the Cloud Storage FUSE deployment and management fully handled by GKE, 
providing a turn-key experience.

# Support

## Supported operating system and validated ML frameworks 

To see supported operating system and ML frameworks that have been validated with Cloud Storage FUSE, see [here](https://cloud.google.com/storage/docs/gcs-fuse#supported-frameworks-os).

## Getting support

You can get support, submit general questions, and request new features by [filing issues in GitHub](https://github.com/GoogleCloudPlatform/gcsfuse/issues). You can also get support by using one of [Google Cloud's official support channels](https://cloud.google.com/support-hub).

See [Troubleshooting](https://github.com/GoogleCloudPlatform/gcsfuse/blob/master/docs/troubleshooting.md) for common issue handling.
</details>