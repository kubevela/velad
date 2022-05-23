#!/bin/bash

set -e

VELA_ADDON_DIR=pkg/resources/static/vela/addons
rm -rf "$VELA_ADDON_DIR"
mkdir -p "$VELA_ADDON_DIR"

if [ -z "$1" ]; then
  echo "No addon(VelaUX) version specified, exiting"
  exit 1
elif [[ $1 == v* ]]; then
  velaux_version=$1
else
  velaux_version=v$1
fi

echo "downloading addons"

addons=("velaux-$velaux_version.tgz")
for addon in ${addons[*]}; do
  echo saving "$addon" to "$VELA_ADDON_DIR"/"$addon"
  curl -L "http://addons.kubevela.net/$addon" -o "$VELA_ADDON_DIR"/"$addon"
done
