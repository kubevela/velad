#!/bin/bash

mkdir -p vela/images

vela_images=("oamdev/vela-core:v1.3.0-alpha.1"\
	"oamdev/cluster-gateway:v1.1.7" \
	"oamdev/kube-webhook-certgen:v2.3")

for img in ${vela_images[*]}; do
	image_name=$(echo $img| cut -f1 -d:| cut -f2 -d/)
  echo saving $img to static/vela/images/"$image_name".tar
  docker pull $img
  docker save -o static/vela/images/"$image_name".tar $img
done
 
