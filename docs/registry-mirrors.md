# 容器镜像源配置说明

## 概述

容器镜像源（Registry Mirrors）配置允许你为不同的容器镜像仓库指定镜像源，从而加速镜像拉取速度。这在中国大陆网络环境下尤其重要，因为直接访问 Docker Hub、Google Container Registry 等国外镜像仓库可能会很慢或无法访问。

## 工作原理

当 Kubernetes 需要拉取容器镜像时（例如 `docker.io/nginx:latest`），Talos Linux 会：

1. 检查是否为该镜像仓库配置了镜像源
2. 如果配置了镜像源，按顺序尝试从镜像源拉取
3. 如果所有镜像源都失败，回退到原始镜像仓库

## 配置示例

在 `cluster-config.yaml` 中添加：

```yaml
registry:
  mirrors:
    docker.io:
      endpoints:
        - "https://docker.mirrors.ustc.edu.cn"
        - "https://hub-mirror.c.163.com"
    k8s.gcr.io:
      endpoints:
        - "https://registry.aliyuncs.com/google_containers"
    gcr.io:
      endpoints:
        - "https://gcr.mirrors.ustc.edu.cn"
    ghcr.io:
      endpoints:
        - "https://ghcr.nju.edu.cn"
    quay.io:
      endpoints:
        - "https://quay.mirrors.ustc.edu.cn"
```

## 支持的镜像仓库

### Docker Hub (docker.io)

Docker Hub 是最常用的容器镜像仓库。

**国内镜像源：**
- 中科大：`https://docker.mirrors.ustc.edu.cn`
- 网易：`https://hub-mirror.c.163.com`
- 阿里云：`https://<your-id>.mirror.aliyuncs.com`（需要注册）

**示例：**
```yaml
docker.io:
  endpoints:
    - "https://docker.mirrors.ustc.edu.cn"
```

### Google Container Registry (gcr.io, k8s.gcr.io)

Google 的容器镜像仓库，Kubernetes 官方镜像存储在这里。

**国内镜像源：**
- 阿里云：`https://registry.aliyuncs.com/google_containers`
- 中科大：`https://gcr.mirrors.ustc.edu.cn`

**示例：**
```yaml
k8s.gcr.io:
  endpoints:
    - "https://registry.aliyuncs.com/google_containers"
gcr.io:
  endpoints:
    - "https://gcr.mirrors.ustc.edu.cn"
```

### GitHub Container Registry (ghcr.io)

GitHub 的容器镜像仓库。

**国内镜像源：**
- 南京大学：`https://ghcr.nju.edu.cn`

**示例：**
```yaml
ghcr.io:
  endpoints:
    - "https://ghcr.nju.edu.cn"
```

### Quay (quay.io)

Red Hat 的容器镜像仓库。

**国内镜像源：**
- 中科大：`https://quay.mirrors.ustc.edu.cn`

**示例：**
```yaml
quay.io:
  endpoints:
    - "https://quay.mirrors.ustc.edu.cn"
```

## 多个镜像源

你可以为同一个镜像仓库配置多个镜像源，系统会按顺序尝试：

```yaml
docker.io:
  endpoints:
    - "https://docker.mirrors.ustc.edu.cn"  # 首先尝试
    - "https://hub-mirror.c.163.com"        # 如果失败，尝试这个
    - "https://dockerhub.azk8s.cn"          # 最后尝试这个
```

## 验证配置

部署集群后，可以通过以下方式验证镜像源配置是否生效：

1. 查看 Talos 配置：
```bash
talosctl get machineconfig -o yaml
```

2. 查看镜像拉取日志：
```bash
talosctl logs kubelet
```

3. 测试拉取镜像：
```bash
kubectl run test --image=nginx:latest
kubectl describe pod test
```

## 常见问题

### Q: 镜像源配置后仍然很慢？

A: 可能的原因：
- 镜像源本身速度慢，尝试更换其他镜像源
- 镜像源不稳定，配置多个镜像源作为备份
- 网络问题，检查防火墙和代理设置

### Q: 某些镜像无法拉取？

A: 可能的原因：
- 镜像源不完整，不是所有镜像都会被同步
- 镜像源更新延迟，新发布的镜像可能需要时间同步
- 尝试使用原始镜像仓库或其他镜像源

### Q: 如何获取阿里云镜像加速器地址？

A: 
1. 登录阿里云容器镜像服务控制台
2. 在左侧导航栏选择"镜像加速器"
3. 获取专属加速器地址，格式为 `https://<your-id>.mirror.aliyuncs.com`

### Q: 镜像源配置会影响已部署的集群吗？

A: 镜像源配置在集群部署时应用到 Talos 配置中。如果需要修改已部署集群的镜像源配置，需要：
1. 修改配置文件
2. 重新生成 Talos 配置
3. 使用 `talosctl apply-config` 应用新配置

## 推荐配置

针对中国大陆网络环境的推荐配置：

```yaml
registry:
  mirrors:
    docker.io:
      endpoints:
        - "https://docker.mirrors.ustc.edu.cn"
        - "https://hub-mirror.c.163.com"
    k8s.gcr.io:
      endpoints:
        - "https://registry.aliyuncs.com/google_containers"
    gcr.io:
      endpoints:
        - "https://gcr.mirrors.ustc.edu.cn"
    ghcr.io:
      endpoints:
        - "https://ghcr.nju.edu.cn"
    quay.io:
      endpoints:
        - "https://quay.mirrors.ustc.edu.cn"
```

这个配置覆盖了最常用的镜像仓库，并为每个仓库配置了可靠的国内镜像源。

## 参考资料

- [Talos Linux Registry Configuration](https://www.talos.dev/latest/reference/configuration/#registriesconfig)
- [Docker 镜像加速器](https://yeasy.gitbook.io/docker_practice/install/mirror)
- [Kubernetes 镜像仓库](https://kubernetes.io/docs/concepts/containers/images/)
