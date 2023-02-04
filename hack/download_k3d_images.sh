#!/bin/bash

set -e
set -x

K3D_IMAGE_DIR=pkg/resources/static/k3d/images
mkdir -p "$K3D_IMAGE_DIR"

ARCH=$1

function download_k3d_images() {
  k3d_images=(
  "$(cat pkg/apis/types.go| grep "K3dImageK3s" |tail -n1 | cut -f2 -d'"')"
  "$(cat pkg/apis/types.go| grep "K3dImageTools" |tail -n1 | cut -f2 -d'"')"
  "$(cat pkg/apis/types.go| grep "K3dImageProxy" |tail -n1 | cut -f2 -d'"')"
  )

  for IMG in ${k3d_images[*]}; do
    IMAGE_NAME=$(echo "$IMG" | cut -f1 -d: | awk -F '/' '{print $NF}')
    echo saving "$IMG" to "$K3D_IMAGE_DIR"/"$IMAGE_NAME".tar
    $DOCKER_PULL "$IMG"
    docker save -o "$K3D_IMAGE_DIR"/"$IMAGE_NAME".tar "$IMG"
    gzip -f "$K3D_IMAGE_DIR"/"$IMAGE_NAME".tar
  done
}

function determine_pull_command() {
  DOCKER_PULL="docker pull --platform=linux/amd64"
  if [ "$1" == "arm64" ]; then
      DOCKER_PULL="docker pull --platform=linux/arm64"
  fi
}
determine_pull_command "$ARCH"
download_k3d_images