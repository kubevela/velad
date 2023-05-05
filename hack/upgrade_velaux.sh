#! /bin/bash

# This script is for upgrade VelaUX

set -e

# If one parameter is passed, use it as the version to upgrade to.
# If two parameters are passed, use the second one as VelaUX image version. Sometimes, we skip the VelaUX image, only upgrade the VelaUX addon.

if [ $# = 1 ]; then
    VERSION_TO=$1
    IMAGE_VERSION=$1
elif [ $# = 2 ]; then
    VERSION_TO=$1
    IMAGE_VERSION=$2
else
    echo "Usage: "$0" version_to [image_version]" >&2
    exit 1
fi

VERION_TO=$1
IMAGE_VERSION=$2


VERSION_NOW=$(cat Makefile |grep "VELAUX_VERSION ?=" |grep -o "v.*")


PATCH_FILE_NAME=$VERSION_NOW-$VERSION_TO.patch
WORKDIR=pkg/resources/static/vela

echo "Upgrading VelaUX version From: "$VERSION_NOW" --> TO: "$VERSION_TO,
if [ -n "$IMAGE_VERSION" ]; then
    echo "Upgrading VelaUX image version to: ""$IMAGE_VERSION"
else
    echo "VelaUX image version is the same as VelaUX addon version"
    IMAGE_VERSION=$VERSION_TO
fi

sed -i "" -e "s/VELAUX_VERSION ?= v.*/VELAUX_VERSION ?= $VERSION_TO/g" Makefile
sed -i "" -e "s/VELAUX_IMAGE_VERSION ?= .*/VELAUX_IMAGE_VERSION ?= $IMAGE_VERSION/g" Makefile
