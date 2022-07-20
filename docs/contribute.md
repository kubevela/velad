# Contribution Guide

This guild helps you get started developing VelaD

### Prerequisites

1. Golang version 1.17+
2. Docker (for non-linux user)
3. golangci-lint 1.38.0+, it will install automatically if you run make, you can install it [manually](https://golangci-lint.run/usage/install/#local-installation) if the installation is too slow.

### Build

1. Clone this project
```shell
git clone https://github.com/kubevela/velad.git
cd velad
```
2. Build VelaD

```shell
make
```
This will build amd64-linux version of VelaD by default. To build other version, you need to specify `OS` and `ARCH`
and the target. For example, you can build a darwin-amd64 version by:

```shell
OS=darwin ARCH=amd64 make darwin-amd64
```

### Debug

When use IDE to debug VelaD, you need to do several things

1. Download resources needed

If you want build linux version, run 
```shell
VELAUX_VERSION=v1.x.y VELA_VERSION=v1.z.w make download_vela_images_addons 
make download_k3s_bin_script 
make download_k3s_images
make pack_vela_chart
```

If you want to build non-linux version, run

```shell
VELAUX_VERSION=v1.x.y VELA_VERSION=v1.z.w make download_vela_images_addons 
make download_k3d 
make pack_vela_chart 
make download_k3s_images
```

`VELAUX_VERSION=v1.x.y VELA_VERSION=v1.z.w` is optional environment variables if you want to change the default version in makefile.

2. Build VelaD

If you are using macOS with intel chip, the complete build command is like:

```shell
OS=darwin ARCH=amd64 \
go build -ldflags="-X github.com/oam-dev/velad/version.VelaVersion=v1.x.y -X github.com/oam-dev/velad/version.VelaUXVersion=v1.x.y"  \
-o bin/velad \
cmd/velad/main.go
```

> Ldflags can help to inject vela-core and VelaUX version. (Can be different)
> If you are using IDE to debug, remember to add `-ldflags="-X github.com...` part to build option.


### Create a pull request

Before you submit a PR, run this command to ensure it is ready:

```shell
make reviewable
```
For other PR things you can check the document [here](https://kubevela.net/docs/contributor/code-contribute#create-a-pull-request).