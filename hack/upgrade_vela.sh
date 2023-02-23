#! /bin/bash

# This script is for upgrade kubevela helm charts maintained in velad repo
# Chart in this repo have one more argument(deployByPod) than that in kubevela repo.

# usage: ./hack/upgrade_vela.sh version_now version_upgrade_to
# e.g. ./hack/upgrade_vela.sh v1.3.3 v1.3.4

set -e

[ $# = 1 ] || { echo "Usage: "$0" version_to" >&2; exit 1; }

VERSION_NOW=$(cat Makefile |grep "VELA_VERSION ?=" |grep -o "v.*")
VERSION_TO=$1
PATCH_FILE_NAME=$VERSION_NOW-$VERSION_TO.patch
WORKDIR=pkg/resources/static/vela

echo "Upgrading KubeVela version From: "$VERSION_NOW" --> TO: "$VERSION_TO

echo "Upgrading go.mod version..."

sed -i "" -e "s/github.com\/oam-dev\/kubevela v.*/github.com\/oam-dev\/kubevela $VERSION_TO/g" go.mod
go mod tidy

echo "Upgrading version variable in Makefile"

sed -i "" -e "s/VELA_VERSION ?= v.*/VELA_VERSION ?= $VERSION_TO/g" Makefile