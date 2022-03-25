K3S_VERSION ?= v1.21.10+k3s1

all: download_vela_images download_k3s
	go build -o bin/velad github.com/oam-dev/velad

download_vela_images:
	./download_images.sh

download_k3s:
	mkdir -p static/k3s
	curl -Lo static/k3s/k3s https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s
	curl -Lo static/k3s/setup.sh https://get.k3s.io
	curl -Lo static/k3s/k3s-airgap-images-amd64.tar.gz https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s-airgap-images-amd64.tar.gz

test:
	@echo https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s
