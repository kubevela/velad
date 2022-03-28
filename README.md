# velad

Setup KubeVela control plane airgapped with high availability of metadata

## Features

1. Air-gap install.
2. High Availability with an External DB. (MySQL/MariaDB, PostgreSQL, ETCD)
 
## Prerequisites

- Linux

## Quickstart

### Setup

Only one command to setup KubeVela control plane

```shell
velad install
```
```shell
Preparing cluster setup script...
Preparing k3s binary...
Successfully place k3s binary to /usr/local/bin/k3s
Preparing k3s images
Successfully prepare k3s image
Setting up cluster...
...
Successfully set up KubeVela control plane, run: export KUBECONFIG=$(vela ctrl-plane kubeconfig) to access it
```

And there you go.

### uninstall

```shell
velad uninstall
```