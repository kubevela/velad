K3S_VERSION ?= v1.21.10+k3s1
STATIC_DIR := pkg/static
all: download_vela_images download_k3s pack_vela_chart
	go build -o bin/velad github.com/oam-dev/velad

download_vela_images:
	./download_images.sh

download_k3s:
	mkdir -p ${STATIC_DIR}/k3s
	curl -Lo ${STATIC_DIR}/k3s/k3s https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s
	curl -Lo ${STATIC_DIR}/k3s/setup.sh https://get.k3s.io
	curl -Lo ${STATIC_DIR}/k3s/k3s-airgap-images-amd64.tar.gz https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s-airgap-images-amd64.tar.gz

pack_vela_chart:
	tar -czf ${STATIC_DIR}/vela/charts/vela-core.tgz ${STATIC_DIR}/vela/charts/vela-core

.PHONY: clean
clean:
	rm ${STATIC_DIR}/vela/charts/vela-core.tgz
	rm bin/velad