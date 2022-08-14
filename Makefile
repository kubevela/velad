include makefiles/dependency.mk

K3S_VERSION ?= v1.21.10+k3s1
STATIC_DIR := pkg/resources/static
VELA_VERSION ?= v1.5.2
VELAUX_VERSION ?= v1.5.2
LDFLAGS= "-X github.com/oam-dev/velad/version.VelaUXVersion=${VELAUX_VERSION} -X github.com/oam-dev/velad/version.VelaVersion=${VELA_VERSION}"
OS ?= linux
ARCH ?= amd64

.DEFAULT_GOAL := linux-amd64
linux-amd64 linux-arm64: download_vela_images_addons pack_vela_chart download_k3s_bin_script download_k3s_images
	GOOS=${OS} GOARCH=${ARCH} \
	go build -o bin/velad-${OS}-${ARCH} \
	-ldflags=${LDFLAGS} \
	github.com/oam-dev/velad/cmd/velad

darwin-amd64 darwin-arm64 windows-amd64: download_vela_images_addons download_k3d pack_vela_chart download_k3s_images
	GOOS=${OS} GOARCH=${ARCH} \
	go build -o bin/velad-${OS}-${ARCH} \
	-ldflags=${LDFLAGS} \
	github.com/oam-dev/velad/cmd/velad

download_vela_images_addons:
	./hack/download_vela_images.sh ${VELA_VERSION} ${VELAUX_VERSION} ${ARCH}
	./hack/download_addons.sh ${VELAUX_VERSION}

download_k3d:
	./hack/download_k3d_images.sh ${ARCH}

download_k3s_bin_script:
	mkdir -p ${STATIC_DIR}/k3s/other
	curl -Lo ${STATIC_DIR}/k3s/other/k3s https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s
	curl -Lo ${STATIC_DIR}/k3s/other/setup.sh https://get.k3s.io

download_k3s_images:
	mkdir -p ${STATIC_DIR}/k3s/images
	curl -Lo ${STATIC_DIR}/k3s/images/k3s-airgap-images.tar.gz https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s-airgap-images-${ARCH}.tar.gz

CHART_DIR := ${STATIC_DIR}/vela/charts
pack_vela_chart:
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
