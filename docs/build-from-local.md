# Build VelaD

You can build VelaD yourself, there are two main components: k3s and KubeVela. You can specify the version
you like of these.

### k3s

To use certain k3s version, add `K3S_VERSION` environment and run `make`:

```shell
git clone https://github.com/oam-dev/velad.git
cd velad
K3S_VERSION=v1.21.10+k3s1 make
```

### KubeVela 

If you want to change the version of KubeVela, use corresponding version of VelaD is recommended.

```shell
git clone https://github.com/oam-dev/velad.git
cd velad
git checkout vx.y.z
make
```