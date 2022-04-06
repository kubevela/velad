# velad

轻量级 KubeVela，在单节点作为守护进程运行且具备高可用性。

## 特性

1. 离线安装
2. 通过外部数据库维持高可用性. (MySQL/MariaDB, PostgreSQL, ETCD)

## 安装条件

- Linux

## 快速开始

### 安装 velad

```shell
curl -Lo velad.tar.gz https://kubevela-docs.oss-cn-beijing.aliyuncs.com/binary/velad/velad-linux-amd64-v1.3.0.tar.gz
tar -xzvf velad.tar.gz
cp linux-amd64/velad /usr/local/bin/velad
```

### 启动 KubeVela

Only one command to setup KubeVela

```shell
velad install
```
```shell
Preparing cluster setup script...
Preparing k3s binary...
Successfully place k3s binary to /usr/local/bin/k3s
Preparing k3s images
Successfully prepare k3s image
Setting up cluster...
...
Successfully set up KubeVela control plane, run: export KUBECONFIG=$(velad kubeconfig) to access it
```
恭喜！你已经设置好 KubeVela 了。用以下命令查看可用的组件。
There you go! You have set up a KubeVela. See available components:

```shell
export KUBECONFIG=$(velad kubeconfig --internal)
vela comp
```
```shell
NAME                    DEFINITION
config-dex-connector    autodetects.core.oam.dev
config-image-registry   autodetects.core.oam.dev
k8s-objects             autodetects.core.oam.dev
my-stateful             statefulsets.apps
raw                     autodetects.core.oam.dev
ref-objects             autodetects.core.oam.dev
snstateful              statefulsets.apps
task                    jobs.batch
webservice              deployments.apps
worker                  deployments.apps
```

### 卸载 KubeVela

```shell
velad uninstall
```

## 更多选项

### 启动高可用的 KubeVela

如果你使用 `velad install` 启动控制平面，那么当你运行 `velad uninstall` 的时候，所有的数据都将丢失。这部分介绍如何用外部数据库
启动高可用的 KubeVela。

1. 准备一个数据库，MySQL/MariaDB, PostgreSQL, ETCD 都可以。按照你的喜好和熟悉程度选择其一。
2. 运行velad并传入数据库链接字符串。

> **保存好 token. 这是连接到这个数据库并重启 KubeVela 所必须的**
```shell
velad install --database-endpoint="mysql://USER:PASSWORD@tcp(HOST:3306)/velad" --token="TOKEN"
```

可以在[这里](db-connect-format.md)找到更多数据库端点的格式

3. 现在你已经启动了 KubeVela，所有数据存在数据库中。

如果这台机器因为某些原因关机了，或者你运行了 `velad uninstall`。你可以用同样的命令加上 `--start` 标志重启 KubeVela。 

```shell
velad install --database-endpoint="mysql://USER:PASSWORD@tcp(HOST:3306)/velad" --token="TOKEN" --start
```

### 从远端访问

velad 默认只提供了 kubeconfig 供你从本地访问。例如你运行 `export KUBECONFIG=$(velad kubeconfig --internal)` 使 vela CLI 能访问 KubeVela

你可以通过如下操作，使得可以从其他机器访问 KubeVela（例如在服务器使用 velad 部署 KubeVela，在本地访问）
1. 当运行`velad install`的时候，添加 `--bind-ip=NODE_IP` 参数，velad会帮助生成在其他机器使用的 kubeconfig。其中的 NODE_IP 是运行 velad 所在机器的 IP。
2. 运行 `velad kubeconfig` (注意到没有 `--internal`) 将会打印可以外部使用 kubeconfig 的位置。
3. 将该文件复制到其他文件。设置好 `KUBECONFIG`，你就可以从远端访问 KubeVela 了！
