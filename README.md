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

- **Linux/macOS**
```shell
curl -fsSl https://static.kubevela.net/script/install-velad.sh | bash -s 1.3.6
```

- **Windows**
> Only the official release version is supported.
```shell
powershell -Command "iwr -useb https://static.kubevela.net/script/install-velad.ps1 | iex"
```

### Setup

To set up KubeVela you only need run `velad install`

```shell
velad install
```
```text
INFO[0000] portmapping '8080:80' targets the loadbalancer: defaulting to [servers:*:proxy agents:*:proxy] 
Preparing K3s images...
...(omit for brevity)

🚀  Successfully install KubeVela control plane
💻  When using gateway trait, you can access with 127.0.0.1:8080
🔭  See available commands with `vela help`
```

There you go! You have set up KubeVela. Behind the command, VelaD starts a K3d container(K3s when Linux), installs vela-core
Helm chart and setup vela CLI for you.

After install, you can follow this [example](./docs/01.simple.md) to deliver your first application.

### uninstall

```shell
velad uninstall
```

### More example

Please check [docs](./docs/) for more VelaD example