# How to upgrade KubeVela version of VelaD

VelaD embed one KubeVela Helm chart and VelaD's build process will cache some images(e.g. vela-core). 
When KubeVela has a new release. Do these steps below to upgrade VelaD's embedded KubeVela version.

1. Upgrade vela-core helm chart
2. Upgrade go.mod
3. Upgrade vela version in makefile

### Upgrade vela-core helm chart.

First check the now vela-core version:

```shell
cat pkg/resources/static/vela/charts/vela-core/Chart.yaml | grep version:
```

Output like:
```text
version: v1.4.2
```

Then use upgrade script, for example if you want to upgrade vela to v1.4.3, use `v1.4.2` and `v1.4.3`
as the first and the second argument:

```shell
./hack/upgrade_vela.sh v1.4.2 v1.4.3
```

This script will clone the KubeVela repo and make diff between tag v1.4.2 and v1.4.3. Then try to patch the diff in VelaD's
embedded vela-core chart in `pkg/resources/static/vela/charts/vela-core`

If there are conflict, you have to resolve them manually.

### Upgrade go.mod

In go.mod file, find this item:

```text
github.com/oam-dev/kubevela v1.x.y
```

Then change the `v1.x.y` to version you want to upgrade to, then run:

```shell
go mod tidy -compat=v1.17
```

### Upgrade vela version in makefile

In `Makefile`, find this two variables, upgrade them to right version.

> VelaUX sometimes don't release new version together with KubeVela, make sure VelaUX version is right.

```makefile
VELA_VERSION ?= v1.4.2
VELAUX_VERSION ?= v1.4.2
```

### After upgrade

Commit all changes and make a pull request.
