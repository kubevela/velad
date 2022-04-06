# velad

Lightweight KubeVela that runs as Daemon in single node with high availability.

## Features

1. Air-gap install.
2. High Availability with an External DB. (MySQL/MariaDB, PostgreSQL, ETCD)
 
## Prerequisites

- Linux

## Quickstart

### Installation

```shell
curl -Lo velad.tar.gz https://github.com/oam-dev/velad/releases/download/v1.3.0/velad-v1.3.0-linux-amd64.tar.gz
tar -xzvf velad.tar.gz
cp linux-amd64/velad /usr/local/velad
```

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
Successfully set up KubeVela control plane, run: export KUBECONFIG=$(velad kubeconfig) to access it
```
There you go! You have set up a KubeVela control plane. See available components:

```shell
export KUBECONFIG=$(velad kubeconfig --internal)
vela comp
```
```shell
NAME                    DEFINITION
config-dex-connector    autodetects.core.oam.dev
config-image-registry   autodetects.core.oam.dev
k8s-objects             autodetects.core.oam.dev
my-stateful             statefulsets.apps
raw                     autodetects.core.oam.dev
ref-objects             autodetects.core.oam.dev
snstateful              statefulsets.apps
task                    jobs.batch
webservice              deployments.apps
worker                  deployments.apps
```

### uninstall

```shell
velad uninstall
```

## More Options

### Setup with high availability

If you run `velad install`, all metadata will be lost when `velad uninstall`. This section describes how to setup a 
high-availability KubeVela control plane with an external database.

1. Prepare a database, MySQL/MariaDB, PostgreSQL, ETCD are both OK. Choose one as you like.
2. Run velad with database connection string.

> **Make sure you keep the token. It is required when restart the control plane**
```shell
velad install --database-endpoint="mysql://USER:PASSWORD@tcp(HOST:3306)/velad" --token="TOKEN"
```

You can find more database endpoint format in this [doc](docs/db-connect-format.md)

3. Now you have a KubeVela control plane which keep all the data in database. 
 
If this control plane is shut down for some reason, or you run `velad uninstall`, you can restart it with `--start` flag and the same token.

```shell
velad install --database-endpoint="mysql://USER:PASSWORD@tcp(HOST:3306)/velad" --token="TOKEN" --start
```

### Access from remote

By default, you can only access this control plane on this node. Typically, you run `export KUBECONFIG=$(velad kubeconfig)`
to access the control plane.

You can also make it accessible outside the node.
1. add `--bind-ip=NODE_IP` when `velad install`, which helps to generate the kubeconfig that can be used outside.
2. `velad kubeconfig` (note without `--internal`) will print the kubeconfig position.
3. copy this file to other machine, setup `KUBECONFIG`, and you can access KubeVela control plane remotely.
