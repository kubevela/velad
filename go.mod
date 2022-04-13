module github.com/oam-dev/velad

go 1.17

require (
	github.com/oam-dev/kubevela v1.3.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.2.1
	k8s.io/utils v0.0.0-20210802155522-efc7438f0176
)

require github.com/zcalusic/sysinfo v0.9.5

replace (
	github.com/docker/cli => github.com/docker/cli v20.10.9+incompatible
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	github.com/wercker/stern => github.com/oam-dev/stern v1.13.2
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client => sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.0.24
)
