#!/bin/bash

VELA_ADDON_DIR=pkg/resources/static/vela/addons
mkdir -p "$VELA_ADDON_DIR"

echo "downloading addons"

addons=("velaux-v1.3.2.tgz")
for addon in ${addons[*]}; do
  echo saving "$addon" to "$VELA_ADDON_DIR"/"$addon"
  curl -L "http://addons.kubevela.net/$addon" -o "$VELA_ADDON_DIR"/"$addon"
done
