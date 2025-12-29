# 快速开始指南

本指南将帮助你在 Proxmox VE 上快速部署一个 Talos Linux Kubernetes 集群。

## 前提条件

- 一台运行 Proxmox VE 的服务器
- 至少 16GB 内存和 100GB 可用存储空间
- 网络连接（用于下载镜像和工具）

## 步骤 1：连接到 Proxmox 主机

### 方式 A：使用 Web Shell（推荐新手）

1. 打开浏览器，访问 Proxmox Web 界面：`https://your-proxmox-ip:8006`
2. 使用 root 账户登录
3. 在左侧树形菜单中，点击你的 Proxmox 节点名称
4. 点击右侧的 "Shell" 按钮，打开终端

### 方式 B：使用 SSH

```bash
ssh root@your-proxmox-ip
```

## 步骤 2：安装依赖工具

在 Proxmox 主机的终端中执行：

```bash
# 安装 talosctl
curl -sL https://talos.dev/install | sh

# 安装 kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
mv kubectl /usr/local/bin/

# 安装 xz 工具（用于解压镜像）
apt-get update && apt-get install -y xz-utils

# 验证安装
talosctl version --client
kubectl version --client
```

## 步骤 3：获取部署工具

### 选项 A：下载预编译版本（推荐）

如果有预编译版本，直接下载：

```bash
# 下载程序（替换为实际下载地址）
wget https://github.com/your-repo/talos-proxmox-deployer/releases/download/v1.0.0/talos-deployer
chmod +x talos-deployer

# 下载示例配置
wget https://raw.githubusercontent.com/your-repo/talos-proxmox-deployer/main/example-config.yaml
```

### 选项 B：从源码编译

```bash
# 安装 Go
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 克隆仓库
git clone https://github.com/your-repo/talos-proxmox-deployer.git
cd talos-proxmox-deployer

# 编译
go build -o talos-deployer
```

### 选项 C：本地交叉编译后上传

在你的开发机器上：

```bash
# 交叉编译为 Linux 版本
GOOS=linux GOARCH=amd64 go build -o talos-deployer

# 上传到 Proxmox
scp talos-deployer root@your-proxmox-ip:/root/
scp example-config.yaml root@your-proxmox-ip:/root/
```

## 步骤 4：创建配置文件

使用交互式���导创建配置：

```bash
./talos-deployer init
```

向导会询问以下信息：

1. **集群名称**：例如 `my-k8s-cluster`
2. **Talos 版本**：例如 `v1.6.0`
3. **Kubernetes 版本**：例如 `1.29`
4. **网络配置**：
   - 网桥名称（通常是 `vmbr0`）
   - DNS 服务器（中国推荐 `223.5.5.5`）
   - 网关地址（例如 `192.168.1.1`）
   - 子网掩码（例如 `24`）
5. **Proxmox 配置**：
   - 主机名（通常是 `pve`）
   - 用户（通常是 `root@pam`）
   - 存储池（例如 `local-lvm`）
   - 模板 VM ID（例如 `9000`）
6. **节点配置**：
   - 控制平面节点数量和配置
   - 工作节点数量和配置

配置将保存到 `cluster-config.yaml`。

## 步骤 5：编辑配置（可选）

如果你在中国或需要使用代理，编辑配置文件添加代理和镜像源：

```bash
vi cluster-config.yaml
```

添加以下配置：

```yaml
# 代理配置
proxy:
  enabled: true
  http_proxy: "http://your-proxy:7890"
  https_proxy: "http://your-proxy:7890"
  no_proxy: "localhost,127.0.0.1,192.168.0.0/16,10.0.0.0/8"
  mirror_url: "https://mirror.ghproxy.com/https://github.com/siderolabs/talos/releases/download"

# 容器镜像源
registry:
  mirrors:
    docker.io:
      endpoints:
        - "https://docker.mirrors.ustc.edu.cn"
    k8s.gcr.io:
      endpoints:
        - "https://registry.aliyuncs.com/google_containers"
```

## 步骤 6：部署集群

执行部署命令：

```bash
./talos-deployer deploy
```

部署过程包括：

1. 📦 下载并准备 Talos 镜像（约 5-10 分钟）
2. 🔧 创建虚拟机模板（约 2 分钟）
3. 🖥️  创建集群节点（约 3-5 分钟）
4. 📝 生成 Talos 配置（约 1 分钟）
5. ⚙️  应用配置到节点（约 5-10 分钟）
6. 🚀 引导 Kubernetes 集群（约 5-10 分钟）

整个过程大约需要 20-40 分钟，具体取决于网络速度和硬件性能。

## 步骤 7：验证集群

部署完成后，验证集群状态：

```bash
./talos-deployer verify
```

你应该看到所有节点都处于 Ready 状态。

## 步骤 8：访问集群

设置 kubeconfig 环境变量：

```bash
export KUBECONFIG=$(pwd)/my-k8s-cluster-config/kubeconfig
```

使用 kubectl 管理集群：

```bash
# 查看节点
kubectl get nodes -o wide

# 查看所有 Pod
kubectl get pods -A

# 部署测试应用
kubectl create deployment nginx --image=nginx
kubectl expose deployment nginx --port=80 --type=NodePort
kubectl get svc nginx
```

## 常见问题

### 下载镜像很慢怎么办？

1. 配置代理（如果有）
2. 使用镜像站（配置 `mirror_url`）
3. 手动下载镜像文件到当前目录

### 节点无法启动？

检查 Proxmox 资源是否充足：

```bash
# 查看内存使用
free -h

# 查看存储空间
df -h

# 查看虚拟机状态
qm list
```

### 配置应用失败？

1. 确保节点已完全启动（等待 2-3 分钟）
2. 检查网络连通性
3. 查看虚拟机控制台输出

### 如何重新部署？

如果部署失败，可以销毁集群后重新部署：

```bash
# 销毁集群
./talos-deployer destroy --force

# 重新部署
./talos-deployer deploy
```

## 下一步

- 阅读 [完整文档](../README.md)
- 了解 [镜像源配置](registry-mirrors.md)
- 学习 [集群管理命令](../README.md#4-管理集群)

## 获取帮助

如果遇到问题：

1. 查看详细日志输出
2. 检查 Proxmox 虚拟机控制台
3. 使用 `talosctl logs` 查看节点日志
4. 提交 Issue 到项目仓库
