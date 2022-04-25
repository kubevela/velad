#!/bin/bash

VELA_IMAGE_DIR=pkg/resources/static/vela/images
mkdir -p "$VELA_IMAGE_DIR"

vela_images=("oamdev/vela-core:v1.3.2"
  "oamdev/cluster-gateway:v1.3.2"
  "oamdev/kube-webhook-certgen:v2.3"
  "oamdev/velaux:v1.3.2"
  "oamdev/vela-apiserver:v1.3.2")

for IMG in ${vela_images[*]}; do
	IMAGE_NAME=$(echo "$IMG" | cut -f1 -d: | cut -f2 -d/)
  echo saving "$IMG" to "$VELA_IMAGE_DIR"/"$IMAGE_NAME".tar
  docker pull "$IMG"
  docker save -o "$VELA_IMAGE_DIR"/"$IMAGE_NAME".tar "$IMG"
done

