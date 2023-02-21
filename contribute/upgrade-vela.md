# How to upgrade KubeVela version of VelaD

VelaD embed one KubeVela Helm chart and VelaD's build process will cache some images(e.g. vela-core). 
When KubeVela has a new release. Do these steps below to upgrade VelaD's embedded KubeVela version.

1. Upgrade go.mod
2. Upgrade vela version in makefile

### Upgrade vela version in makefile

In `Makefile`, find this two variables, upgrade them to right version.

> VelaUX sometimes don't release new version together with KubeVela, make sure VelaUX version is right.

```makefile
VELAUX_VERSION ?= v1.6.0
VELAUX_IMAGE_VERSION ?= ${VELAUX_VERSION}
```

### After upgrade

Commit all changes and make a pull request.
