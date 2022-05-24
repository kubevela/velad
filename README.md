# VelaD

Lightweight KubeVela that runs as Daemon in single node with k3s

English | [简体中文](docs/readme-zh.md)

![Build](https://github.com/kubevela/velad/actions/workflows/build.yaml/badge.svg)

## Introduction

VelaD is lightweight deployment tool to set up [KubeVela](https://github.com/kubevela/kubevela).

VelaD make it very easy to set up KubeVela environment, including a cluster with KubeVela installed, VelaUX/Vela CLI prepared.

VelaD is the fastest way to get started with KubeVela.


![demo](docs/resources/demo.gif)

## Features

1. Set up KubeVela air-gapped
2. Build KubeVela control plane with higher availability (Optional)

## Prerequisites

If you are using Windows/macOS, docker is needed for run VelaD

## Quickstart

### Installation

- Linux/macOS
```shell
```shell
curl -fsSl https://static.kubevela.net/script/install-velad.sh | bash -s 1.3.4
```

- Windows
```shell
powershell -Command "iwr -useb https://static.kubevela.net/script/install.ps1 | iex"
```

### Setup

Only one command to setup KubeVela

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

Successfully install KubeVela control plane! Try: vela components
```
There you go! You have set up a KubeVela. See available components:

```shell
vela comp
```
```shell
NAME                    DEFINITION
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

### Setup with database

If you run `velad install`, all metadata will be lost when `velad uninstall`. You may need to keep the metadata for migration
This section describes how to setup a KubeVela control plane with an external database.

1. Prepare a database, MySQL/MariaDB, PostgreSQL, ETCD are both OK. Choose one as you like.
2. Run velad with database connection string.

> **Make sure you keep the token. It is required when restart KubeVela using this database**
```shell
velad install --database-endpoint="mysql://USER:PASSWORD@tcp(HOST:3306)/velad" --token="TOKEN"
```

You can find more database endpoint format in this [doc](docs/db-connect-format.md)

3. Now you have a KubeVela control plane which keeps all the data in database. 
 
If this machine is shut down for some reason, or you run `velad uninstall`, you can restart it with `--start` flag and the same token.

```shell
velad install --database-endpoint="mysql://USER:PASSWORD@tcp(HOST:3306)/velad" --token="TOKEN" --start
```

### Access from remote

By default, you can only access this control plane on this node. Typically, you run `export KUBECONFIG=$(velad kubeconfig --internal)`
to access the control plane.

You can also make it accessible outside the machine.
1. add `--bind-ip=NODE_IP` when `velad install`, which helps to generate the kubeconfig that can be used outside. `NODE_IP`
is IP of machine where run the `velad`。
2. `velad kubeconfig` (note without `--internal`) will print the kubeconfig position.
3. copy this file to other machine, setup `KUBECONFIG`, and you can access KubeVela control plane remotely.


## Build

You could build velad yourself. This requires:

- docker
- go >= 1.17

```shell
git clone https://github.com/oam-dev/velad.git
cd velad
make
```

See more options [here](docs/build-from-local.md)
