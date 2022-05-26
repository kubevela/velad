# velad

VelaD 是一个轻量级部署工具，能帮助你快速搭建 [KubeVela](https://github.com/kubevela/kubevela) 环境。

使用 VelaD，能方便地搭建 KubeVela 环境，包括一个安装有 KubeVela 的集群、配套命令行工具 vela CLI、Web 控制面板 VelaUX

VelaD 是上手 KubeVela 的最快方式。

## 特性

1. 离线搭建 KubeVela 环境。
2. 可以连接数据库，搭建更高可用性多接点的 KubeVela 控制平面。
3. 在一台机器上轻松体验 KubeVela 多集群特性。

## 安装条件

如果你的操作系统是Windows/macOS，VelaD的运行需要[Docker](https://www.docker.com/products/docker-desktop/) 。

## 快速开始

### 安装 VelaD

- Linux/macOS
```shell
```shell
curl -fsSl https://static.kubevela.net/script/install-velad.sh | bash -s 1.3.5
```

- Windows
```shell
powershell -Command "iwr -useb https://static.kubevela.net/script/install.ps1 | iex"
```

### 使用 VelaD 部署 KubeVela

Only one command to setup KubeVela

```shell
velad install
```
```shell
INFO[0000] portmapping '8080:80' targets the loadbalancer: defaulting to [servers:*:proxy agents:*:proxy] 
Preparing K3s images...
...(omit for brevity)

🚀  Successfully install KubeVela control plane
💻  When using gateway trait, you can access with 127.0.0.1:8080
🔭  See available commands with `vela help`
```
恭喜！你已经搭建好一个 KubeVela 的环境了。在这条命令背后，VelaD启动了一个 K3s 容器（如果在 Linux 上，则是 K3s 进程），在其中安装了 vela-core，
并在你的机器上设置了vela CLI。

你可以查看这个[例子](01.simple.md)，使用 KubeVela 来部署你的第一个应用

### 卸载 KubeVela

```shell
velad uninstall
```

### 更多案例

查看[文档](../docs)获取更多 VelaD 的使用方法和案例。
