# VelaD

Lightweight Deploy tool, helps setup [KubeVela](https://github.com/kubevela/kubevela) quickly。

English | [简体中文](docs/readme-zh.md)

![E2E Test](https://github.com/kubevela/velad/actions/workflows/e2e-test.yaml/badge.svg)
![Build status](https://github.com/kubevela/velad/workflows/Go/badge.svg)
![Docker Pulls](https://img.shields.io/docker/pulls/oamdev/vela-core)
[![codecov](https://codecov.io/gh/kubevela/velad/branch/master/graph/badge.svg)](https://codecov.io/gh/kubevela/velad)
[![LICENSE](https://img.shields.io/github/license/kubevela/velad.svg?style=flat-square)](/LICENSE)
[![Releases](https://img.shields.io/github/release/kubevela/velad/all.svg?style=flat-square)](https://github.com/kubevela/velad/releases)
[![TODOs](https://img.shields.io/endpoint?url=https://api.tickgit.com/badge?repo=github.com/kubevela/velad)](https://www.tickgit.com/browse?repo=github.com/kubevela/velad)
[![Twitter](https://img.shields.io/twitter/url?style=social&url=https%3A%2F%2Ftwitter.com%2Foam_dev)](https://twitter.com/oam_dev)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/4602/badge)](https://bestpractices.coreinfrastructure.org/projects/4602)
![E2E status](https://github.com/kubevela/velad/workflows/E2E%20Test/badge.svg)
[![](https://img.shields.io/badge/KubeVela-Check%20Your%20Contribution-orange)](https://opensource.alibaba.com/contribution_leaderboard/details?projectValue=kubevela)


## Introduction

VelaD is lightweight deployment tool to set up [KubeVela](https://github.com/kubevela/kubevela).

VelaD make it very easy to set up KubeVela environment, including a cluster with KubeVela installed, VelaUX/Vela CLI prepared.

VelaD is the fastest way to get started with KubeVela.


![demo](docs/resources/demo.gif)

## Features

1. Set up KubeVela air-gapped.
2. Build KubeVela control plane with higher availability with more nodes and database(Optional).
3. Experience KubeVela multi-cluster features in one computer.

## Prerequisites

If you are using Windows/macOS, docker is needed for run VelaD

## Quickstart

### Installation

- **Linux/macOS**
```shell
curl -fsSl https://static.kubevela.net/script/install-velad.sh | bash
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

## Known issues

- Installation on darwin-arm64 (Apple chip) machine isn't fully air-gapped. Please track #64 for more info.
