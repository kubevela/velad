#!/bin/bash

K3D_IMAGE_DIR=pkg/resources/static/k3d/images
mkdir -p "$K3D_IMAGE_DIR"

vela_images=("ghcr.io/k3d-io/k3d-tools:5.4.1"
  "ghcr.io/k3d-io/k3d-proxy:5.4.1"
  "docker.io/rancher/k3s:v1.21.10-k3s1")

for IMG in ${vela_images[*]}; do
  IMAGE_NAME=$(echo "$IMG" | cut -f1 -d: | cut -f3 -d/)
  echo saving "$IMG" to "$K3D_IMAGE_DIR"/"$IMAGE_NAME".tar
  docker pull "$IMG"
  docker save -o "$K3D_IMAGE_DIR"/"$IMAGE_NAME".tar "$IMG"
  gzip -f "$K3D_IMAGE_DIR"/"$IMAGE_NAME".tar
done
