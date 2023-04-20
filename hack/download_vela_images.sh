#!/bin/bash

set -e
set -x

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

if [ -z "$2" ]; then
  echo "No VelaUX version specified, exiting"
  exit 1
elif [[ $2 == v* ]]; then
  velaux_version=$2
else
  velaux_version=v$2
fi

# optional, amd64 if not set
ARCH=$3

function set_cluster_gateway_version() {
    cluster_gateway_version=UNKNOWN
    image_tag=$(cat pkg/resources/static/vela/charts/vela-core/values.yaml | grep -A 1 oamdev/cluster-gateway | grep tag)
    cluster_gateway_version=$(echo $image_tag| cut -f2 -d:|xargs)
    echo "cluster-gateway image version detected:" $cluster_gateway_version
}

function set_certgen_version() {
    certgen_version=UNKNOWN
    image_tag=$(cat pkg/resources/static/vela/charts/vela-core/values.yaml | grep -A 1 oamdev/kube-webhook-certgen | grep tag)
    certgen_version=$(echo $image_tag| cut -f2 -d:|xargs)
    echo "kube-webhook-certgen image version detected:" $certgen_version
}

function download_images() {
  vela_images=("oamdev/vela-core:${vela_version}"
    "oamdev/cluster-gateway:${cluster_gateway_version}"
    "oamdev/kube-webhook-certgen:${certgen_version}"
    "oamdev/velaux:${velaux_version}")

  for IMG in ${vela_images[*]};
    do
      IMAGE_NAME=$(echo "$IMG" | cut -f1 -d: | cut -f2 -d/)
      echo saving "$IMG" to "$VELA_IMAGE_DIR"/"$IMAGE_NAME".tar.gz
      $DOCKER_PULL "$IMG"
      docker save -o "$VELA_IMAGE_DIR"/"$IMAGE_NAME".tar "$IMG"
      gzip -f "$VELA_IMAGE_DIR"/"$IMAGE_NAME".tar
    done
}

function determine_pull_command() {
  DOCKER_PULL="docker pull --platform=linux/amd64"
  if [ "$1" == "arm64" ]; then
      DOCKER_PULL="docker pull --platform=linux/arm64"
  fi
}


determine_pull_command "$ARCH"
set_cluster_gateway_version
set_certgen_version
download_images
