package deployer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"talos-proxmox-deployer/pkg/config"
)

type Deployer struct {
	config *config.ClusterConfig
}

// getProxmoxEnv 返回配置了 Proxmox 认证的环境变量
func (d *Deployer) getProxmoxEnv() []string {
	env := os.Environ()

	// 根据配置的认证方式选择
	switch d.config.Proxmox.AuthMethod {
	case "api_token":
		// 使用 API Token 认证
		if d.config.Proxmox.APITokenID != "" && d.config.Proxmox.APIToken != "" {
			env = append(env, fmt.Sprintf("PROXMOX_TOKEN_ID=%s", d.config.Proxmox.APITokenID))
			env = append(env, fmt.Sprintf("PROXMOX_TOKEN_SECRET=%s", d.config.Proxmox.APIToken))
		} else {
			fmt.Println("⚠️  警告: auth_method 设置为 api_token，但未配置 api_token_id 或 api_token")
		}
	case "password":
		// 使用密码认证
		if d.config.Proxmox.Password != "" {
			env = append(env, fmt.Sprintf("PROXMOX_PASSWORD=%s", d.config.Proxmox.Password))
		} else {
			fmt.Println("⚠️  警告: auth_method 设置为 password，但未配置 password")
		}
	default:
		// 兼容旧配置：如果没有指定 auth_method，优先使用 API Token
		if d.config.Proxmox.APITokenID != "" && d.config.Proxmox.APIToken != "" {
			env = append(env, fmt.Sprintf("PROXMOX_TOKEN_ID=%s", d.config.Proxmox.APITokenID))
			env = append(env, fmt.Sprintf("PROXMOX_TOKEN_SECRET=%s", d.config.Proxmox.APIToken))
		} else if d.config.Proxmox.Password != "" {
			env = append(env, fmt.Sprintf("PROXMOX_PASSWORD=%s", d.config.Proxmox.Password))
		}
	}

	// Proxmox 主机和用户
	env = append(env, fmt.Sprintf("PROXMOX_HOST=%s", d.config.Proxmox.Host))
	env = append(env, fmt.Sprintf("PROXMOX_USER=%s", d.config.Proxmox.User))

	// TLS 验证
	if d.config.Proxmox.SkipTLSVerify {
		env = append(env, "PROXMOX_SKIP_TLS_VERIFY=1")
	}

	return env
}

// execProxmoxCommand 执行带认证的 Proxmox 命令
func (d *Deployer) execProxmoxCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Env = d.getProxmoxEnv()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func New(cfg *config.ClusterConfig) *Deployer {
	return &Deployer{config: cfg}
}

// getProxyEnv 返回配置了代理的环境变量
func (d *Deployer) getProxyEnv() []string {
	env := os.Environ()

	if d.config.Proxy.Enabled {
		if d.config.Proxy.HTTPProxy != "" {
			env = append(env, fmt.Sprintf("HTTP_PROXY=%s", d.config.Proxy.HTTPProxy))
			env = append(env, fmt.Sprintf("http_proxy=%s", d.config.Proxy.HTTPProxy))
		}
		if d.config.Proxy.HTTPSProxy != "" {
			env = append(env, fmt.Sprintf("HTTPS_PROXY=%s", d.config.Proxy.HTTPSProxy))
			env = append(env, fmt.Sprintf("https_proxy=%s", d.config.Proxy.HTTPSProxy))
		}
		if d.config.Proxy.NoProxy != "" {
			env = append(env, fmt.Sprintf("NO_PROXY=%s", d.config.Proxy.NoProxy))
			env = append(env, fmt.Sprintf("no_proxy=%s", d.config.Proxy.NoProxy))
		}
	}

	return env
}

func (d *Deployer) PrepareImage() error {
	fmt.Println("📦 准备 Talos 镜像...")

	imageFile := fmt.Sprintf("talos-%s.qcow2", d.config.TalosVersion)
	if _, err := os.Stat(imageFile); err == nil {
		fmt.Printf("✓ 镜像已存在: %s\n", imageFile)
		return nil
	}

	rawImage := "talos-metal-amd64.raw"
	xzFile := rawImage + ".xz"

	// 下载镜像
	if _, err := os.Stat(xzFile); os.IsNotExist(err) {
		// 确定下载 URL
		var url string
		if d.config.Proxy.Enabled && d.config.Proxy.MirrorURL != "" {
			// 使用自定义镜像站
			url = fmt.Sprintf("%s/%s/metal-amd64.raw.xz", d.config.Proxy.MirrorURL, d.config.TalosVersion)
			fmt.Printf("使用镜像站下载: %s\n", url)
		} else {
			// 使用官方源
			url = fmt.Sprintf("https://github.com/siderolabs/talos/releases/download/%s/metal-amd64.raw.xz", d.config.TalosVersion)
			fmt.Printf("下载镜像: %s\n", url)
		}

		// 构建 wget 命令
		args := []string{"-q", "--show-progress", url, "-O", xzFile}

		// 配置代理
		cmd := exec.Command("wget", args...)
		if d.config.Proxy.Enabled {
			cmd.Env = d.getProxyEnv()
			if d.config.Proxy.HTTPProxy != "" || d.config.Proxy.HTTPSProxy != "" {
				fmt.Println("✓ 使用代理下载")
			}
		}

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("下载失败: %w", err)
		}
	}

	// 解压
	fmt.Println("解压镜像...")
	cmd := exec.Command("xz", "-d", "-k", "-f", xzFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}

	// 转换为 qcow2
	fmt.Println("转换镜像格式...")
	cmd = exec.Command("qemu-img", "convert",
		"-f", "raw",
		"-O", "qcow2",
		"-c",
		"-o", "cluster_size=64k,preallocation=metadata,lazy_refcounts=on,compression_type=zlib",
		rawImage,
		imageFile,
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("转换失败: %w", err)
	}

	// 清理
	os.Remove(rawImage)
	os.Remove(xzFile)

	fmt.Printf("✓ 镜像准备完成: %s\n", imageFile)
	return nil
}

func (d *Deployer) CreateTemplate() error {
	fmt.Println("🔧 创建 Talos 模板...")

	vmID := d.config.Proxmox.TemplateVMID

	// 检查模板是否存在
	cmd := exec.Command("qm", "status", fmt.Sprintf("%d", vmID))
	cmd.Env = d.getProxmoxEnv()
	if cmd.Run() == nil {
		fmt.Printf("✓ 模板已存在 (VM ID: %d)\n", vmID)
		return nil
	}

	// 创建虚拟机
	args := []string{
		"create", fmt.Sprintf("%d", vmID),
		"--name", "talos-template",
		"--memory", "1024",
		"--cores", "1",
		"--cpu", "host",
		"--net0", fmt.Sprintf("virtio,bridge=%s", d.config.Network.Bridge),
		"--scsihw", "virtio-scsi-pci",
		"--machine", "q35",
		"--bios", "ovmf",
		"--efidisk0", fmt.Sprintf("%s:4,format=qcow2", d.config.Proxmox.StoragePool),
		"--agent", "enabled=1",
	}
	if err := d.execProxmoxCommand("qm", args...); err != nil {
		return fmt.Errorf("创建虚拟机失败: %w", err)
	}

	// 导入磁盘
	imageFile := fmt.Sprintf("talos-%s.qcow2", d.config.TalosVersion)
	if err := d.execProxmoxCommand("qm", "importdisk",
		fmt.Sprintf("%d", vmID),
		imageFile,
		d.config.Proxmox.StoragePool,
		"--format", "qcow2",
	); err != nil {
		return fmt.Errorf("导入磁盘失败: %w", err)
	}

	// 附加磁盘
	diskSpec := fmt.Sprintf("%s:vm-%d-disk-0,discard=on,cache=writeback,iothread=1,ssd=1",
		d.config.Proxmox.StoragePool, vmID)
	if err := d.execProxmoxCommand("qm", "set", fmt.Sprintf("%d", vmID), "--scsi0", diskSpec); err != nil {
		return fmt.Errorf("附加磁盘失败: %w", err)
	}

	// 设置启动顺序
	if err := d.execProxmoxCommand("qm", "set", fmt.Sprintf("%d", vmID), "--boot", "order=scsi0"); err != nil {
		return fmt.Errorf("设置启动顺序失败: %w", err)
	}

	// 转换为模板
	if err := d.execProxmoxCommand("qm", "template", fmt.Sprintf("%d", vmID)); err != nil {
		return fmt.Errorf("转换模板失败: %w", err)
	}

	fmt.Printf("✓ 模板创建完成 (VM ID: %d)\n", vmID)
	return nil
}

func (d *Deployer) CreateNodes() error {
	fmt.Println("🖥️  创建集群节点...")

	allNodes := append(d.config.Nodes.ControlPlanes, d.config.Nodes.Workers...)

	for _, node := range allNodes {
		if err := d.createNode(node); err != nil {
			return fmt.Errorf("创建节点 %s 失败: %w", node.Name, err)
		}
	}

	fmt.Println("✓ 所有节点创建完成")
	return nil
}

func (d *Deployer) createNode(node config.NodeSpec) error {
	fmt.Printf("  创建节点: %s (VM ID: %d)\n", node.Name, node.VMID)

	// 克隆模板
	if err := d.execProxmoxCommand("qm", "clone",
		fmt.Sprintf("%d", d.config.Proxmox.TemplateVMID),
		fmt.Sprintf("%d", node.VMID),
		"--name", node.Name,
		"--full", "1",
	); err != nil {
		return fmt.Errorf("克隆失败: %w", err)
	}

	// 配置资源
	diskSpec := fmt.Sprintf("%s:vm-%d-disk-0,discard=on,cache=writeback,iothread=1,ssd=1,size=%s",
		d.config.Proxmox.StoragePool, node.VMID, node.Disk)

	if err := d.execProxmoxCommand("qm", "set", fmt.Sprintf("%d", node.VMID),
		"--cores", fmt.Sprintf("%d", node.CPU),
		"--memory", fmt.Sprintf("%d", node.Memory),
		"--scsi0", diskSpec,
	); err != nil {
		return fmt.Errorf("配置资源失败: %w", err)
	}

	// 启动节点
	if err := d.execProxmoxCommand("qm", "start", fmt.Sprintf("%d", node.VMID)); err != nil {
		return fmt.Errorf("启动节点失败: %w", err)
	}

	time.Sleep(2 * time.Second)
	return nil
}

func (d *Deployer) GenerateConfig() error {
	fmt.Println("📝 生成 Talos 配置...")

	configDir := fmt.Sprintf("./%s-config", d.config.ClusterName)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 获取第一个控制平面 IP
	controlPlaneIP := d.config.Nodes.ControlPlanes[0].IPAddress
	endpoint := fmt.Sprintf("https://%s:6443", controlPlaneIP)

	// 生成基础配置
	cmd := exec.Command("talosctl", "gen", "config",
		d.config.ClusterName,
		endpoint,
		"--output", configDir,
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("生成配置失败: %w", err)
	}

	// 如果配置了镜像源，修改配置文件
	if d.config.Registry != nil && len(d.config.Registry.Mirrors) > 0 {
		if err := d.applyRegistryConfig(configDir); err != nil {
			return fmt.Errorf("应用镜像源配置失败: %w", err)
		}
	}

	fmt.Printf("✓ 配置已生成到: %s\n", configDir)
	return nil
}

// applyRegistryConfig 将镜像源配置应用到 Talos 配置文件
func (d *Deployer) applyRegistryConfig(configDir string) error {
	fmt.Println("  应用镜像源配置...")

	// 需要修改的配置文件
	configFiles := []string{
		filepath.Join(configDir, "controlplane.yaml"),
		filepath.Join(configDir, "worker.yaml"),
	}

	for _, configFile := range configFiles {
		if err := d.patchConfigFile(configFile); err != nil {
			return fmt.Errorf("修改配置文件 %s 失败: %w", configFile, err)
		}
	}

	return nil
}

// patchConfigFile 使用 talosctl patch 命令修改配置文件
func (d *Deployer) patchConfigFile(configFile string) error {
	// 构建 registry 配置的 JSON patch
	patchContent := d.buildRegistryPatch()

	// 创建临时 patch 文件
	patchFile := configFile + ".patch.json"
	if err := os.WriteFile(patchFile, []byte(patchContent), 0644); err != nil {
		return fmt.Errorf("创建 patch 文件失败: %w", err)
	}
	defer os.Remove(patchFile)

	// 使用 talosctl patch 命令
	cmd := exec.Command("talosctl", "machineconfig", "patch",
		configFile,
		"--patch", fmt.Sprintf("@%s", patchFile),
		"--output", configFile,
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("patch 配置失败: %w", err)
	}

	return nil
}

// buildRegistryPatch 构建镜像源配置的 JSON patch
func (d *Deployer) buildRegistryPatch() string {
	var mirrors []string

	for registry, mirror := range d.config.Registry.Mirrors {
		endpoints := make([]string, len(mirror.Endpoints))
		for i, ep := range mirror.Endpoints {
			endpoints[i] = fmt.Sprintf(`"%s"`, ep)
		}
		endpointsStr := strings.Join(endpoints, ",")

		mirrorConfig := fmt.Sprintf(`"%s":{"endpoints":[%s]}`, registry, endpointsStr)
		mirrors = append(mirrors, mirrorConfig)
	}

	mirrorsStr := strings.Join(mirrors, ",")

	patch := fmt.Sprintf(`[
  {
    "op": "add",
    "path": "/machine/registries",
    "value": {
      "mirrors": {%s}
    }
  }
]`, mirrorsStr)

	return patch
}

func (d *Deployer) ApplyConfig() error {
	fmt.Println("⚙️  应用 Talos 配置...")

	configDir := fmt.Sprintf("./%s-config", d.config.ClusterName)

	// 应用控制平面配置
	for _, node := range d.config.Nodes.ControlPlanes {
		fmt.Printf("  应用配置到控制平面: %s (%s)\n", node.Name, node.IPAddress)

		configFile := filepath.Join(configDir, "controlplane.yaml")
		cmd := exec.Command("talosctl", "apply-config",
			"--insecure",
			"--nodes", node.IPAddress,
			"--file", configFile,
			"--timeout", "5m",
		)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("应用配置到 %s 失败: %w", node.Name, err)
		}
		time.Sleep(5 * time.Second)
	}

	// 应用工作节点配置
	for _, node := range d.config.Nodes.Workers {
		fmt.Printf("  应用配置到工作节点: %s (%s)\n", node.Name, node.IPAddress)

		configFile := filepath.Join(configDir, "worker.yaml")
		cmd := exec.Command("talosctl", "apply-config",
			"--insecure",
			"--nodes", node.IPAddress,
			"--file", configFile,
			"--timeout", "5m",
		)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("应用配置到 %s 失败: %w", node.Name, err)
		}
		time.Sleep(5 * time.Second)
	}

	fmt.Println("✓ 配置应用完成")
	return nil
}

func (d *Deployer) Bootstrap() error {
	fmt.Println("🚀 引导 Kubernetes 集群...")

	// 配置端点
	var endpoints []string
	for _, node := range d.config.Nodes.ControlPlanes {
		endpoints = append(endpoints, node.IPAddress)
	}
	endpointStr := strings.Join(endpoints, ",")

	firstCP := d.config.Nodes.ControlPlanes[0].IPAddress

	cmd := exec.Command("talosctl", "config", "endpoint", endpointStr)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("配置端点失败: %w", err)
	}

	cmd = exec.Command("talosctl", "config", "node", firstCP)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("配置节点失败: %w", err)
	}

	// 等待节点准备
	fmt.Println("等待节点准备...")
	time.Sleep(30 * time.Second)

	// 引导集群
	cmd = exec.Command("talosctl", "bootstrap",
		"--nodes", firstCP,
		"--timeout", "5m",
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("引导失败: %w", err)
	}

	// 获取 kubeconfig
	configDir := fmt.Sprintf("./%s-config", d.config.ClusterName)
	kubeconfigPath := filepath.Join(configDir, "kubeconfig")

	cmd = exec.Command("talosctl", "kubeconfig",
		kubeconfigPath,
		"--nodes", firstCP,
		"--force",
	)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("获取 kubeconfig 失败: %w", err)
	}

	fmt.Printf("✓ 集群引导完成\n")
	fmt.Printf("✓ kubeconfig 已保存到: %s\n", kubeconfigPath)
	return nil
}

func (d *Deployer) Verify() error {
	configDir := fmt.Sprintf("./%s-config", d.config.ClusterName)
	kubeconfigPath := filepath.Join(configDir, "kubeconfig")

	os.Setenv("KUBECONFIG", kubeconfigPath)

	fmt.Println("检查节点状态...")
	cmd := exec.Command("kubectl", "get", "nodes", "-o", "wide")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("获取节点状态失败: %w", err)
	}

	fmt.Println("\n检查 Pod 状态...")
	cmd = exec.Command("kubectl", "get", "pods", "-A", "-o", "wide")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("获取 Pod 状态失败: %w", err)
	}

	fmt.Println("\n✓ 集群验证完成")
	return nil
}

func (d *Deployer) StartNodes() error {
	allNodes := append(d.config.Nodes.ControlPlanes, d.config.Nodes.Workers...)

	for _, node := range allNodes {
		fmt.Printf("  启动节点: %s (VM ID: %d)\n", node.Name, node.VMID)
		if err := d.execProxmoxCommand("qm", "start", fmt.Sprintf("%d", node.VMID)); err != nil {
			fmt.Printf("  ⚠️  节点 %s 可能已在运行\n", node.Name)
		}
	}

	fmt.Println("✓ 节点启动完成")
	return nil
}

func (d *Deployer) StopNodes() error {
	allNodes := append(d.config.Nodes.ControlPlanes, d.config.Nodes.Workers...)

	for _, node := range allNodes {
		fmt.Printf("  停止节点: %s (VM ID: %d)\n", node.Name, node.VMID)
		if err := d.execProxmoxCommand("qm", "stop", fmt.Sprintf("%d", node.VMID)); err != nil {
			fmt.Printf("  ⚠️  停止节点 %s 失败\n", node.Name)
		}
	}

	fmt.Println("✓ 节点停止完成")
	return nil
}

func (d *Deployer) Destroy() error {
	allNodes := append(d.config.Nodes.ControlPlanes, d.config.Nodes.Workers...)

	// 停止并删除所有节点
	for _, node := range allNodes {
		fmt.Printf("  销毁节点: %s (VM ID: %d)\n", node.Name, node.VMID)

		cmd := exec.Command("qm", "stop", fmt.Sprintf("%d", node.VMID))
		cmd.Env = d.getProxmoxEnv()
		cmd.Run()
		time.Sleep(1 * time.Second)

		if err := d.execProxmoxCommand("qm", "destroy", fmt.Sprintf("%d", node.VMID), "--purge"); err != nil {
			fmt.Printf("  ⚠️  删除节点 %s 失败\n", node.Name)
		}
	}

	// 删除模板
	vmID := d.config.Proxmox.TemplateVMID
	fmt.Printf("  删除模板 (VM ID: %d)\n", vmID)
	cmd := exec.Command("qm", "destroy", fmt.Sprintf("%d", vmID), "--purge")
	cmd.Env = d.getProxmoxEnv()
	cmd.Run()

	// 清理配置文件
	configDir := fmt.Sprintf("./%s-config", d.config.ClusterName)
	os.RemoveAll(configDir)

	imageFile := fmt.Sprintf("talos-%s.qcow2", d.config.TalosVersion)
	os.Remove(imageFile)

	fmt.Println("✓ 清理完成")
	return nil
}
