all: download_vela_images download_k3s
	go build -o bin/velad github.com/oam-dev/velad

download_vela_images:
	./download_images.sh

download_k3s:
	mkdir -p k3s
	curl -Lo k3s/k3s https://github.com/k3s-io/k3s/releases/download/v1.21.10%2Bk3s1/k3s
	curl -Lo k3s/setup.sh https://get.k3s.io
	curl -Lo k3s/k3s-airgap-images-amd64.tar.gz https://github.com/k3s-io/k3s/releases/download/v1.21.10%2Bk3s1/k3s-airgap-images-amd64.tar.gz