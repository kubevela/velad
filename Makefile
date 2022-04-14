K3S_VERSION ?= v1.21.10+k3s1
STATIC_DIR := pkg/static
#VELA_VERSION := 1.3.0

all: download_vela_images download_k3s pack_vela_chart
	go build -o bin/velad github.com/oam-dev/velad

download_vela_images:
	./download_images.sh

download_k3s:
	mkdir -p ${STATIC_DIR}/k3s
	curl -Lo ${STATIC_DIR}/k3s/k3s https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s
	curl -Lo ${STATIC_DIR}/k3s/setup.sh https://get.k3s.io
	curl -Lo ${STATIC_DIR}/k3s/k3s-airgap-images-amd64.tar.gz https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s-airgap-images-amd64.tar.gz

CHART_DIR := ${STATIC_DIR}/vela/charts
pack_vela_chart:
	#curl -Lo ${CHART_DIR}/vela-core-${VELA_VERSION}.tgz https://kubevelacharts.oss-cn-hangzhou.aliyuncs.com/core/vela-core-${VELA_VERSION}.tgz
	#tar -xzf ${CHART_DIR}/vela-core-${VELA_VERSION}.tgz -C ${CHART_DIR}
	#patch -s -p1 -t -D ${CHART_DIR}/vela-core-${VELA_VERSION} < ${CHART_DIR}/vela-core.patch

	cp -r ${STATIC_DIR}/vela/charts/vela-core .
	tar -czf ${STATIC_DIR}/vela/charts/vela-core.tgz vela-core
	rm -r vela-core

.PHONY: clean
clean:
	rm -rf ${CHART_DIR}/vela-core-*
	rm -rf ${CHART_DIR}/vela-core
	#rm ${STATIC_DIR}/vela/charts/vela-core.tgz
	rm -f bin/velad