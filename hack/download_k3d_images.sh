#!/bin/bash

K3D_IMAGE_DIR=pkg/resources/static/k3d/images
mkdir -p "$K3D_IMAGE_DIR"

vela_images=("rancher/k3d-tools:5.2.2"
  "rancher/k3d-proxy:5.2.2"
  "rancher/k3s:v1.21.10-k3s1")

for IMG in ${vela_images[*]}; do
  IMAGE_NAME=$(echo "$IMG" | cut -f1 -d: | cut -f2 -d/)
  echo saving "$IMG" to "$K3D_IMAGE_DIR"/"$IMAGE_NAME".tar
  docker pull "$IMG"
  docker save -o "$K3D_IMAGE_DIR"/"$IMAGE_NAME".tar "$IMG"
  gzip -f "$K3D_IMAGE_DIR"/"$IMAGE_NAME".tar
done
