# Talos Proxmox Deployer

一个用于在 Proxmox VE 上自动化部署 Talos Linux Kubernetes 集群的 Go 工具。

## 功能特性

- 🚀 交互式配置向导
- 📦 自动下载和准备 Talos 镜像
- 🖥️  自动创建和配置虚拟机
- ⚙️  自动生成和应用 Talos 配置
- 🔍 集群健康检查和验证
- 🔧 集群管理（启动、停止、重启）
- 🗑️  一键销毁集群

## 执行环境

**重要：此程序必须在 Proxmox VE 主机上执行**

程序需要直接调用 Proxmox 的 `qm` 命令来管理虚拟机，因此必须在 Proxmox VE 主机的 Shell 中运行。

### 两种执行方式

#### 方式 1：在 Proxmox Web 控制台的 Shell 中执行（推荐）

1. 登录 Proxmox Web 界面
2. 选择你的 Proxmox 节点
3. 点击 "Shell" 按钮打开终端
4. 在终端中执行部署命令

#### 方式 2：通过 SSH 连接到 Proxmox 主机执行

```bash
# 从你的电脑 SSH 连接到 Proxmox 主机
ssh root@your-proxmox-host

# 在 Proxmox 主机上执行部署命令
./talos-deployer deploy
```

## 安装

### 前置要求

**在 Proxmox VE 主机上需要安装：**

- Go 1.21+（用于编译，可选）
- `talosctl` 命令行工具
- `kubectl` 命令行工具
- `qemu-img` 工具（Proxmox 默认已安装）
- `wget` 或 `curl`（用于下载镜像）
- `xz` 工具（用于解压镜像）

### 在 Proxmox 主机上安装依赖

```bash
# 安装 talosctl
curl -sL https://talos.dev/install | sh

# 安装 kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
mv kubectl /usr/local/bin/

# 安装 xz（如果没有）
apt-get update && apt-get install -y xz-utils
```

### 编译

**选项 1：在 Proxmox 主机上编译**

```bash
# 安装 Go（如果没有）
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 编译程序
go mod download
go build -o talos-deployer
```

**选项 2：在本地编译后上传到 Proxmox**

```bash
# 在你的开发机器上交叉编译
GOOS=linux GOARCH=amd64 go build -o talos-deployer

# 上传到 Proxmox 主机
scp talos-deployer root@your-proxmox-host:/root/
scp example-config.yaml root@your-proxmox-host:/root/

# SSH 到 Proxmox 主机
ssh root@your-proxmox-host
chmod +x talos-deployer
```

## 使用方法

> 💡 **新手？** 查看 [快速开始指南](docs/quick-start.md) 获取详细的分步说明。

### 1. 初始化配置

通过交互式向导创建集群配置：

```bash
./talos-deployer init
```

这将引导你完成以下配置：
- 集群基础信息（名称、版本）
- 网络配置（网桥、DNS、网关）
- Proxmox 配置（主机、存储池）
- 节点配置（控制平面和工作节点）

配置将保存到 `cluster-config.yaml` 文件。

### 2. 部署集群

```bash
./talos-deployer deploy
```

可选参数：
- `-c, --config`: 指定配置文件路径（默认: cluster-config.yaml）
- `-s, --skip-prepare`: 跳过镜像准备步骤
- `-t, --skip-template`: 跳过模板创建步骤
- `--skip-config`: 跳过配置生成步骤
- `--skip-bootstrap`: 跳过集群引导步骤

### 3. 验证集群

```bash
./talos-deployer verify
```

### 4. 管理集群

启动集群节点：
```bash
./talos-deployer manage start
```

停止集群节点：
```bash
./talos-deployer manage stop
```

重启集群节点：
```bash
./talos-deployer manage restart
```

### 5. 销毁集群

```bash
./talos-deployer destroy
```

强制销毁（不询问确认）：
```bash
./talos-deployer destroy --force
```

## 配置文件示例

完整的配置文件示例请参考 [example-config.yaml](example-config.yaml)。

### 基础配置

```yaml
cluster_name: my-talos-cluster
talos_version: v1.6.0
kubernetes_version: "1.29"

network:
  bridge: vmbr0
  dns_server: 8.8.8.8
  gateway: 192.168.1.1
  netmask: "24"

proxmox:
  host: pve
  user: root@pam
  
  # 认证方式选择: "password" 或 "api_token"
  auth_method: api_token  # 推荐使用 api_token
  
  # API Token 认证配置（当 auth_method=api_token 时使用）
  api_token_id: "root@pam!deployer"
  api_token: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  
  # 密码认证配置（当 auth_method=password 时使用）
  # password: "your-password"
  
  storage_pool: local-lvm
  template_vm_id: 9000
  skip_tls_verify: false  # 生产环境设为 false

nodes:
  control_planes:
    - vm_id: 101
      ip_address: 192.168.1.101
      name: talos-cp-1
      cpu: 2
      memory: 2048
      disk: 20G
      role: controlplane
    - vm_id: 102
      ip_address: 192.168.1.102
      name: talos-cp-2
      cpu: 2
      memory: 2048
      disk: 20G
      role: controlplane
    - vm_id: 103
      ip_address: 192.168.1.103
      name: talos-cp-3
      cpu: 2
      memory: 2048
      disk: 20G
      role: controlplane

  workers:
    - vm_id: 201
      ip_address: 192.168.1.201
      name: talos-worker-1
      cpu: 4
      memory: 4096
      disk: 50G
      role: worker
    - vm_id: 202
      ip_address: 192.168.1.202
      name: talos-worker-2
      cpu: 4
      memory: 4096
      disk: 50G
      role: worker
```

### 代理配置（针对中国网络环境）

如果你在中国或需要通过代理访问网络，可以在配置文件中添加代理设置：

```yaml
proxy:
  enabled: true
  http_proxy: "http://proxy.example.com:8080"
  https_proxy: "http://proxy.example.com:8080"
  no_proxy: "localhost,127.0.0.1,192.168.0.0/16,10.0.0.0/8"
  # 使用 GitHub 镜像站加速 Talos 镜像下载
  mirror_url: "https://mirror.ghproxy.com/https://github.com/siderolabs/talos/releases/download"
```

常见的 GitHub 镜像站：
- `https://mirror.ghproxy.com/https://github.com/siderolabs/talos/releases/download`
- `https://ghproxy.com/https://github.com/siderolabs/talos/releases/download`
- `https://gh.api.99988866.xyz/https://github.com/siderolabs/talos/releases/download`

国内推荐 DNS 服务器：
- 阿里云：`223.5.5.5` 或 `223.6.6.6`
- 腾讯云：`119.29.29.29`
- 114DNS：`114.114.114.114`

### 容器镜像源配置（加速镜像拉取）

为了加速 Kubernetes 容器镜像的拉取，可以配置国内镜像源：

```yaml
registry:
  mirrors:
    # Docker Hub 镜像
    docker.io:
      endpoints:
        - "https://docker.mirrors.ustc.edu.cn"
        - "https://hub-mirror.c.163.com"
    # Google 容器镜像
    k8s.gcr.io:
      endpoints:
        - "https://registry.aliyuncs.com/google_containers"
    gcr.io:
      endpoints:
        - "https://gcr.mirrors.ustc.edu.cn"
    # GitHub 容器镜像
    ghcr.io:
      endpoints:
        - "https://ghcr.nju.edu.cn"
    # Quay 镜像
    quay.io:
      endpoints:
        - "https://quay.mirrors.ustc.edu.cn"
```

常用国内镜像源：

**Docker Hub 镜像：**
- 中科大：`https://docker.mirrors.ustc.edu.cn`
- 网易：`https://hub-mirror.c.163.com`
- 阿里云：`https://<your-id>.mirror.aliyuncs.com`（需要注册获取）

**Google 容器镜像：**
- 阿里云：`https://registry.aliyuncs.com/google_containers`
- 中科大：`https://gcr.mirrors.ustc.edu.cn`

**GitHub 容器镜像：**
- 南京大学：`https://ghcr.nju.edu.cn`

**Quay 镜像：**
- 中科大：`https://quay.mirrors.ustc.edu.cn`

配置镜像源后，Kubernetes 会自动使用这些镜像源来拉取容器镜像，大大提升部署速度。

## 部署流程

1. **准备镜像**: 下载并转换 Talos Linux 镜像为 qcow2 格式
2. **创建模板**: 在 Proxmox 中创建虚拟机模板
3. **创建节点**: 从模板克隆并配置所有节点
4. **生成配置**: 使用 talosctl 生成集群配置
5. **应用配置**: 将配置应用到所有节点
6. **引导集群**: 初始化 Kubernetes 集群
7. **验证**: 检查集群状态和健康度

## 访问集群

部署完成后，使用以下命令访问集群：

```bash
export KUBECONFIG=$(pwd)/my-talos-cluster-config/kubeconfig
kubectl get nodes
kubectl get pods -A
```

## 故障排除

### 认证失败
- 检查 Proxmox 认证配置是否正确
- 确认 API Token 或密码有效
- 查看详细说明：[认证配置文档](docs/authentication.md)

### 镜像下载失败
- 确保网络连接正常
- 如果在中国，建议配置代理或使用镜像站
- 可以手动下载镜像文件到当前目录
- 检查 `proxy` 配置是否正确

### 下载速度慢
- 配置 `mirror_url` 使用国内镜像站
- 配置 HTTP/HTTPS 代理
- 使用 VPN 或其他加速工具

### 节点无法启动
检查 Proxmox 资源是否充足，查看虚拟机日志。

### 配置应用失败
确保节点已完全启动，可以增加等待时间。

### 集群引导失败
检查控制平面节点网络连通性，确保端口 6443 可访问。

## 许可证

MIT License

## 相关文档

- [快速开始指南](docs/quick-start.md) - 新手入门教程
- [认证配置文档](docs/authentication.md) - Proxmox 认证配置详解
- [镜像源配置](docs/registry-mirrors.md) - 容器镜像加速配置
- [架构说明](docs/architecture.md) - 系统架构和设计
