include makefiles/dependency.mk

K3S_VERSION ?= v1.21.10+k3s1
STATIC_DIR := pkg/resources/static
GOOS ?= linux
GOARCH ?= amd64
#VELA_VERSION := 1.3.0

.DEFAULT_GOAL := linux-amd64
linux-amd64: download_vela_images_addons download_k3s pack_vela_chart
	go build -o bin/velad github.com/oam-dev/velad/cmd/velad

darwin-amd64 windows-amd64: download_vela_images_addons download_k3d pack_vela_chart download_k3s_images
	GOOS=${GOOS} GOARCH=${GOARCH} go build -o bin/velad-${GOOS}-${GOARCH} github.com/oam-dev/velad/cmd/velad

download_vela_images_addons:
	./hack/download_vela_images.sh
	./hack/download_addons.sh

download_k3d:
	./hack/download_k3d_images.sh

download_k3s: download_k3s_images
	mkdir -p ${STATIC_DIR}/k3s/other
	curl -Lo ${STATIC_DIR}/k3s/other/k3s https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s
	curl -Lo ${STATIC_DIR}/k3s/other/setup.sh https://get.k3s.io

download_k3s_images:
	mkdir -p ${STATIC_DIR}/k3s/images
	curl -Lo ${STATIC_DIR}/k3s/images/k3s-airgap-images-amd64.tar.gz https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s-airgap-images-amd64.tar.gz

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
	rm -f ${CHART_DIR}/vela-core.tgz
	rm -f bin/velad

lint: golangci
	$(GOLANGCILINT) run ./...

staticcheck: staticchecktool
	$(STATICCHECK) ./...

fmt: goimports
	$(GOIMPORTS) -local github.com/kubevela/velad -w $$(go list -f {{.Dir}} ./...)

go-check:
	go fmt ./...
	go vet ./...

reviewable: lint staticcheck fmt go-check
	go mod tidy -compat=1.17

check-diff: reviewable
	git --no-pager diff
	git diff --quiet || (echo please run 'make reviewable' to include all changes && false)
	echo branch is clean
