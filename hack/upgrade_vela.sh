#! /bin/bash

# This script is for upgrade kubevela helm charts maintained in velad repo
# Chart in this repo have one more argument(deployByPod) than that in kubevela repo.

# usage: ./hack/upgrade_vela.sh version_now version_upgrade_to
# e.g. ./hack/upgrade_vela.sh v1.3.3 v1.3.4

set -e

[ $# = 2 ] || { echo "Usage: "$0" version_now version_to" >&2; exit 1; }

VERSION_NOW=$1
VERSION_TO=$2
PATCH_FILE_NAME=$VERSION_NOW-$VERSION_TO.patch
WORKDIR=pkg/resources/static/vela

git clone git@github.com:kubevela/kubevela.git

pushd kubevela
git diff refs/tags/"$VERSION_NOW"...refs/tags/"$VERSION_TO" charts/vela-core > "$PATCH_FILE_NAME"
popd

mv kubevela/"$PATCH_FILE_NAME" .

echo "Patching charts..."
git apply -v --check --reject --apply --directory $WORKDIR "$PATCH_FILE_NAME"
echo "Patching done"

rm "$PATCH_FILE_NAME"
rm -rf kubevela