#!/bin/bash

set -e

VELA_IMAGE_DIR=pkg/resources/static/vela/images
rm -rf "$VELA_IMAGE_DIR"
mkdir -p "$VELA_IMAGE_DIR"

if [ -z "$1" ]; then
  echo "No kubevela version specified, exiting"
  exit 1
elif [[ $1 == v* ]]; then
  vela_version=$1
else
  vela_version=v$1
fi

vela_images=("oamdev/vela-core:${vela_version}"
  "oamdev/cluster-gateway:v1.3.2"
  "oamdev/kube-webhook-certgen:v2.3"
  "oamdev/velaux:${vela_version}"
  "oamdev/vela-apiserver:${vela_version}")

for IMG in ${vela_images[*]}; do
  IMAGE_NAME=$(echo "$IMG" | cut -f1 -d: | cut -f2 -d/)
  echo saving "$IMG" to "$VELA_IMAGE_DIR"/"$IMAGE_NAME".tar.gz
  docker pull "$IMG"
  docker save -o "$VELA_IMAGE_DIR"/"$IMAGE_NAME".tar "$IMG"
  gzip -f "$VELA_IMAGE_DIR"/"$IMAGE_NAME".tar
done
