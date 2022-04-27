# 使用 VelaD 快速创建高可用的多集群控制平面

本文将介绍如何从零开始快速创建一个多集群控制平面，并发布一个应用

### 简介
如今，在越来越多的场景下，开发者和系统运维人员开始将应用部署在多个集群中。如何管理不同集群中的应用，如何快速
搭建一个好用的集群控制平面，成为了一个问题。

下文将展示如何借助 VelaD 工具，从零开始五分钟内创建一个多集群控制平面，并发布一个应用。

### 开始之前

1. 准备一个 Linux 系统的机器
2. 准备一个需要管理的子集群：以一个 kubeconfig 为 us-west 的子集群为例
3. （可选）准备一个数据库，以MySQL为例，其他支持的数据库见[数据库支持文档](db-connect-format.md)

### 下载 VelaD

```shell
curl -Lo velad.tar.gz https://kubevela-docs.oss-cn-beijing.aliyuncs.com/binary/velad/velad-linux-amd64-v1.3.1.tar.gz
tar -xzvf velad.tar.gz
cp linux-amd64/velad /usr/local/bin/velad
```

确认你已经安装成功：

```shell
velad version
```

### 创建多集群控制平面

最简单的情况下，创建多集群控制平面，只需要一条命令：`velad install`。你还可以使用一个数据库来保证数据的更高可用性。

该命令将为你在机器上创建一个单节点的 k3s 集群，并在其中安装 KubeVela。如果你还不熟悉 KubeVela，它
是一个现代化的应用交付与管理平台，原生支持多集群应用交付。VelaD 还帮你设置好了操作该控制平面的命令行工具 vela。

例子中 `--database-endpoint` 参数，用到了准备的数据库，将用户名、密码、以及数据库所在机器的IP地址替换为你的数据 ，
以及你想要使用的数据库（以 VelaD 为例）如果使用了这个选项，你将可以将控制平面的全部数据存在其中。即使机器故障，你
也能快速从其他机器重启控制平面。当然你也可以不使用该参数、所有的数据将存储于你的本地。

```shell
$ velad install --database-endpoint="mysql://user:password@tcp(IP:3306)/velad" 
Preparing cluster setup script...
Preparing k3s binary...
Successfully place k3s binary to /usr/local/bin/k3s
Preparing k3s images
Successfully prepare k3s image
Setting up cluster...
...
Successfully set up KubeVela control plane, run: export KUBECONFIG=$(velad kubeconfig --internal) to access it

Keep the token below in case of restarting the control plane
<TOKEN>
```

确认控制平面已经正常安装，根据 `velad install` 最后的提示：

```shell
export KUBECONFIG=$(velad kubeconfig --internal)
vela components
```

这将列出可用的组件:

```shell
NAME            DEFINITION
raw             autodetects.core.oam.dev
cron-task       cronjobs.batch
webservice      deployments.apps
k8s-objects     autodetects.core.oam.dev
ref-objects     autodetects.core.oam.dev
task            jobs.batch
worker          deployments.apps
```
 
### 连接子集群

使用配套安装好的 vela 命令行工具，将子集群加入到控制平面的管控中来。

```shell
vela cluster join <your kubeconfig path>
```

子集群加入之后，你可以使用 `vela cluster list` 来查看被管控的所有集群。

```shell
$ vela cluster list
CLUSTER                 TYPE            ENDPOINT                ACCEPTED        LABELS
local                   Internal        -                       true                  
cluster-us-west         X509Certificate <ENDPOINT_US_WEST>      true                  
```

### 部署多集群应用

这是 KubeVela 1.3 中部署多集群应用的一个例子。 你只需要使用 topology 策略来声明要部署的集群，就可以部署多集群应用了。

例如，你可以使用下面这个样例将 nginx webservice 部署在 us-west 集群中，

```shell
cat <<EOF | vela up -f -
apiVersion: core.oam.dev/v1beta1
kind: Application
metadata:
  name: basic-topology
  namespace: examples
spec:
  components:
    - name: nginx-basic
      type: webservice
      properties:
        image: nginx
      traits:
        - type: expose
          properties:
            port: [80]
  policies:
    - name: topology-us-west-clusters
      type: topology
      properties:
        clusters: ["us-west"]
EOF
```

此时你已经成功部署了一个多集群的应用！你可以使用 `vela status` 来查看部署状态

```shell
$ vela status basic-topology -n examples
About:

  Name:         basic-topology               
  Namespace:    examples                     
  Created at:   2022-04-10 14:37:54 +0800 CST
  Status:       workflowFinished             

Workflow:

  mode: DAG
  finished: true
  Suspend: false
  Terminated: false
  Steps
  - id:3mvz5i8elj
    name:deploy-topology-us-west-clusters
    type:deploy
    phase:succeeded 
    message:

Services:

  - Name: nginx-basic  
    Cluster: us-west  Namespace: examples
    Type: webservice
    Healthy Ready:1/1
    Traits:
      ✅ expose
```

当然你可以使用这个控制平面对多集群进行更多需求，例如：使用集群 labels 按组分发、在不同集群进行配置差异化等，你可以在 
[KubeVela 文档](https://kubevela.io/zh/docs/case-studies/multi-cluster) 中找到这些更多用法

### 进阶使用：提高控制平面的可用性

上面介绍的 `velad install` 将会在你的机器中将k3s注册为服务并启动，当机器重启时，服务会自动启动。
如果你在创建控制平面时，使用了一个数据库作为存储。那么当你遇到当出现更严重的问题或者其他情况时，你将拥有更高的数据可用性，例如：

1. 机器出现物理故障，至少无法再重启
2. 随着业务规模的提升，需要将控制平面迁移到更大规格的机器
3. 你运行 `velad uninstall` 卸载了控制平面

在你迁移控制平面的时候，不用担心子集群，其中所有的工作负载将不受任何影响，当控制平面迁移完毕，所有的子集群将自动回到管控当中
假设你现在使用 `--database-endpoint` 参数安装了控制平面，并且希望迁移控制平面。你可以这样做：

1. 在原机器上运行 `velad uninstall`
2. 在新机器上运行 `velad install --database-endpoint=<ENDPOINT> --token=<TOKEN> --start`

在新机器上运行的命令，需要使用与原机器上启动控制平面时相同的 `database-endpoint`，而且使用当时启动后，
提示你保存的token。最后的 `--start` 参数表示仅启动，跳过 KubeVela 安装过程，因为在数据库所保存的控制平面元数据中，
KubeVela 已经安装了，无需重复安装。

以上就是本次的全部内容，感谢你的阅读和尝试。Velad 还在持续开发，下一步将支持在 Mac/Windows 上面启动
控制平面，将给多集群管理带来更多灵活和便捷。