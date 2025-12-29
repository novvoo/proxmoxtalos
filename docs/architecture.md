# 架构说明

## 执行环境架构

```
┌─────────────────────────────────────────────────────────────┐
│                      你的电脑                                 │
│  - 编写/编辑配置文件                                          │
│  - 编译程序（可选）                                           │
│  - SSH 连接到 Proxmox                                        │
└─────────────────────┬───────────────────────────────────────┘
                      │ SSH / Web Shell
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                  Proxmox VE 主机                             │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  talos-deployer 程序（在这里执行）                      │  │
│  │  - 调用 qm 命令管理虚拟机                               │  │
│  │  - 调用 talosctl 配置 Talos                            │  │
│  │  - 调用 kubectl 验证集群                               │  │
│  └───────────────────┬───────────────────────────────────┘  │
│                      │                                       │
│                      ▼                                       │
│  ┌─────────────────────────────────────────────────────┐    │
│  │           Proxmox 虚拟机管理                         │    │
│  │  - 创建模板 VM                                       │    │
│  │  - 克隆虚拟机                                        │    │
│  │  - 配置资源（CPU、内存、磁盘）                        │    │
│  │  - 启动/停止虚拟机                                   │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │  Control     │  │  Control     │  │  Control     │      │
│  │  Plane 1     │  │  Plane 2     │  │  Plane 3     │      │
│  │  (Talos VM)  │  │  (Talos VM)  │  │  (Talos VM)  │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐                        │
│  │  Worker 1    │  │  Worker 2    │                        │
│  │  (Talos VM)  │  │  (Talos VM)  │                        │
│  └──────────────┘  └──────────────┘                        │
└─────────────────────────────────────────────────────────────┘
```

## 为什么必须在 Proxmox 主机上执行？

### 1. qm 命令依赖

程序使用 Proxmox 的 `qm` 命令来管理虚拟机：

```go
// 创建虚拟机
cmd := exec.Command("qm", "create", "101", "--name", "talos-cp-1", ...)

// 克隆虚拟机
cmd := exec.Command("qm", "clone", "9000", "101", ...)

// 启动虚拟机
cmd := exec.Command("qm", "start", "101")
```

`qm` 命令只在 Proxmox VE 主机上可用，无法远程调用。

### 2. 本地文件访问

程序需要访问 Proxmox 的本地存储来导入磁盘镜像：

```go
// 导入磁盘到 Proxmox 存储
cmd := exec.Command("qm", "importdisk", "9000", "talos-v1.6.0.qcow2", "local-lvm")
```

### 3. 直接网络访问

程序需要直接访问虚拟机的 IP 地址来配置 Talos：

```go
// 应用配置到节点
cmd := exec.Command("talosctl", "apply-config", 
    "--nodes", "192.168.1.101",
    "--file", "controlplane.yaml")
```

这些 IP 通常在 Proxmox 的内部网络中，外部无法直接访问。

## 部署流程

### 阶段 1：准备阶段（在 Proxmox 主机上）

```
1. 下载 Talos 镜像
   wget https://github.com/.../metal-amd64.raw.xz
   
2. 解压镜像
   xz -d metal-amd64.raw.xz
   
3. 转换格式
   qemu-img convert -f raw -O qcow2 metal-amd64.raw talos.qcow2
```

### 阶段 2：创建模板（在 Proxmox 主机上）

```
1. 创建虚拟机
   qm create 9000 --name talos-template ...
   
2. 导入磁盘
   qm importdisk 9000 talos.qcow2 local-lvm
   
3. 配置虚拟机
   qm set 9000 --scsi0 local-lvm:vm-9000-disk-0 ...
   
4. 转换为模板
   qm template 9000
```

### 阶段 3：创建节点（在 Proxmox 主机上）

```
对每个节点：
1. 克隆模板
   qm clone 9000 101 --name talos-cp-1
   
2. 配置资源
   qm set 101 --cores 2 --memory 2048 --scsi0 ...,size=20G
   
3. 启动虚拟机
   qm start 101
```

### 阶段 4：配置集群（在 Proxmox 主机上）

```
1. 生成 Talos 配置
   talosctl gen config my-cluster https://192.168.1.101:6443
   
2. 应用配置到节点
   talosctl apply-config --nodes 192.168.1.101 --file controlplane.yaml
   
3. 引导集群
   talosctl bootstrap --nodes 192.168.1.101
   
4. 获取 kubeconfig
   talosctl kubeconfig ./kubeconfig
```

### 阶段 5：验证集群（在 Proxmox 主机上）

```
1. 检查节点状态
   kubectl get nodes
   
2. 检查 Pod 状态
   kubectl get pods -A
```

## 网络拓扑

```
Internet
    │
    ▼
┌─────────────────────────────────────┐
│  Proxmox VE 主机                     │
│  IP: 192.168.1.100                  │
│                                     │
│  ┌─────────────────────────────┐   │
│  │  vmbr0 (虚拟网桥)            │   │
│  │  连接到物理网卡              │   │
│  └──────────┬──────────────────┘   │
│             │                       │
│  ┌──────────┴──────────────────┐   │
│  │  Talos 虚拟机网络            │   │
│  │  192.168.1.101 - CP1        │   │
│  │  192.168.1.102 - CP2        │   │
│  │  192.168.1.103 - CP3        │   │
│  │  192.168.1.201 - Worker1    │   │
│  │  192.168.1.202 - Worker2    │   │
│  └─────────────────────────────┘   │
└─────────────────────────────────────┘
```

## 数据流

### 镜像下载流程

```
Internet (GitHub/镜像站)
    │
    ▼ wget/curl
Proxmox 主机文件系统
    │ talos-v1.6.0.qcow2
    ▼ qm importdisk
Proxmox 存储池 (local-lvm)
    │
    ▼ qm clone
虚拟机磁盘
```

### 配置应用流程

```
talos-deployer
    │
    ▼ talosctl gen config
配置文件 (controlplane.yaml, worker.yaml)
    │
    ▼ talosctl apply-config
Talos 节点 (通过网络)
    │
    ▼ 节点应用配置
Kubernetes 集群启动
```

## 安全考虑

### 为什么不支持远程执行？

1. **安全性**：`qm` 命令需要 root 权限，远程执行存在安全风险
2. **复杂性**：需要配置 SSH 密钥、权限管理等
3. **性能**：镜像文件较大（~1GB），远程传输耗时
4. **可靠性**：网络中断会导致部署失败

### 最佳实践

1. **使用 SSH 密钥**：避免使用密码登录 Proxmox
2. **限制访问**：只允许特定 IP 访问 Proxmox SSH
3. **定期备份**：备份配置文件和 kubeconfig
4. **监控日志**：记录所有部署操作

## 替代方案

如果你确实需要远程管理，可以考虑：

### 方案 1：Proxmox API

使用 Proxmox API 进行远程管理（需要重写程序）：

```go
// 使用 Proxmox API 而不是 qm 命令
client := proxmox.NewClient(apiURL, apiToken)
client.CreateVM(...)
```

### 方案 2：Terraform

使用 Terraform Proxmox Provider：

```hcl
resource "proxmox_vm_qemu" "talos_node" {
  name        = "talos-cp-1"
  target_node = "pve"
  clone       = "talos-template"
  ...
}
```

### 方案 3：Ansible

使用 Ansible 通过 SSH 执行命令：

```yaml
- name: Create Talos VM
  shell: qm create 101 --name talos-cp-1 ...
  delegate_to: proxmox-host
```

但这些方案都比直接在 Proxmox 主机上执行更复杂。

## 总结

- ✅ **推荐**：在 Proxmox 主机上直接执行（简单、可靠、快速）
- ⚠️ **可选**：通过 SSH 连接后执行（需要配置 SSH）
- ❌ **不推荐**：远程 API 调用（复杂、需要重写程序）

对于大多数用户，直接在 Proxmox Web Shell 中执行是最简单的方式。
