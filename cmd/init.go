package cmd

import (
	"fmt"
	"os"
	"strconv"

	"talos-proxmox-deployer/pkg/config"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "交互式初始化集群配置",
	Long:  `通过交互式问答创建集群配置文件`,
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 Talos Proxmox 集群配置向导")
	fmt.Println("================================")
	fmt.Println()

	cfg := &config.ClusterConfig{}

	// 集群基础配置
	if err := promptClusterBasics(cfg); err != nil {
		return err
	}

	// 网络配置
	if err := promptNetworkConfig(cfg); err != nil {
		return err
	}

	// Proxmox 配置
	if err := promptProxmoxConfig(cfg); err != nil {
		return err
	}

	// 节点配置
	if err := promptNodesConfig(cfg); err != nil {
		return err
	}

	// 代理配置
	if err := promptProxyConfig(cfg); err != nil {
		return err
	}

	// 容器镜像源配置
	if err := promptRegistryConfig(cfg); err != nil {
		return err
	}

	// 保存配置
	return saveConfig(cfg)
}

func promptClusterBasics(cfg *config.ClusterConfig) error {
	fmt.Println("📋 集群基础配置")
	fmt.Println("----------------")

	prompt := promptui.Prompt{
		Label:   "集群名称",
		Default: "talos-proxmox-cluster",
	}
	name, err := prompt.Run()
	if err != nil {
		return err
	}
	cfg.ClusterName = name

	prompt = promptui.Prompt{
		Label:   "Talos 版本",
		Default: "v1.6.0",
	}
	version, err := prompt.Run()
	if err != nil {
		return err
	}
	cfg.TalosVersion = version

	prompt = promptui.Prompt{
		Label:   "Kubernetes 版本",
		Default: "1.29",
	}
	k8sVersion, err := prompt.Run()
	if err != nil {
		return err
	}
	cfg.KubernetesVersion = k8sVersion

	fmt.Println()
	return nil
}

func promptNetworkConfig(cfg *config.ClusterConfig) error {
	fmt.Println("🌐 网络配置")
	fmt.Println("------------")

	prompt := promptui.Prompt{
		Label:   "网络桥接",
		Default: "vmbr0",
	}
	bridge, err := prompt.Run()
	if err != nil {
		return err
	}
	cfg.Network.Bridge = bridge

	prompt = promptui.Prompt{
		Label:   "DNS 服务器",
		Default: "8.8.8.8",
	}
	dns, err := prompt.Run()
	if err != nil {
		return err
	}
	cfg.Network.DNSServer = dns

	prompt = promptui.Prompt{
		Label:   "网关",
		Default: "192.168.1.1",
	}
	gateway, err := prompt.Run()
	if err != nil {
		return err
	}
	cfg.Network.Gateway = gateway

	prompt = promptui.Prompt{
		Label:   "子网掩码位数",
		Default: "24",
	}
	netmask, err := prompt.Run()
	if err != nil {
		return err
	}
	cfg.Network.Netmask = netmask

	fmt.Println()
	return nil
}

func promptProxmoxConfig(cfg *config.ClusterConfig) error {
	fmt.Println("🖥️  Proxmox 配置")
	fmt.Println("----------------")

	prompt := promptui.Prompt{
		Label:   "Proxmox 主机",
		Default: "pve",
	}
	host, err := prompt.Run()
	if err != nil {
		return err
	}
	cfg.Proxmox.Host = host

	prompt = promptui.Prompt{
		Label:   "Proxmox 用户",
		Default: "root@pam",
	}
	user, err := prompt.Run()
	if err != nil {
		return err
	}
	cfg.Proxmox.User = user

	// 认证方式选择
	fmt.Println()
	fmt.Println("💡 认证方式说明:")
	fmt.Println("   - password: 使用密码认证（简单，但不推荐用于生产环境）")
	fmt.Println("   - api_token: 使用 API Token 认证（推荐，更安全）")
	fmt.Println("   创建 API Token: Datacenter -> Permissions -> API Tokens -> Add")
	fmt.Println()

	selectPrompt := promptui.Select{
		Label: "选择认证方式",
		Items: []string{"password", "api_token"},
	}
	_, authMethod, err := selectPrompt.Run()
	if err != nil {
		return err
	}
	cfg.Proxmox.AuthMethod = authMethod

	if authMethod == "password" {
		prompt = promptui.Prompt{
			Label: "Proxmox 密码",
			Mask:  '*',
		}
		password, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.Proxmox.Password = password
	} else {
		prompt = promptui.Prompt{
			Label:   "API Token ID",
			Default: "root@pam!deployer",
		}
		tokenID, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.Proxmox.APITokenID = tokenID

		prompt = promptui.Prompt{
			Label: "API Token Secret",
		}
		token, err := prompt.Run()
		if err != nil {
			return err
		}
		cfg.Proxmox.APIToken = token
	}

	prompt = promptui.Prompt{
		Label:   "存储池",
		Default: "local-lvm",
	}
	storage, err := prompt.Run()
	if err != nil {
		return err
	}
	cfg.Proxmox.StoragePool = storage

	prompt = promptui.Prompt{
		Label:   "模板 VM ID",
		Default: "9000",
	}
	templateID, err := prompt.Run()
	if err != nil {
		return err
	}
	id, _ := strconv.Atoi(templateID)
	cfg.Proxmox.TemplateVMID = id

	// TLS 验证选项
	fmt.Println()
	selectPrompt = promptui.Select{
		Label: "跳过 TLS 证书验证（仅开发环境）",
		Items: []string{"否（推荐）", "是"},
	}
	_, tlsResult, err := selectPrompt.Run()
	if err != nil {
		return err
	}
	cfg.Proxmox.SkipTLSVerify = (tlsResult == "是")

	fmt.Println()
	return nil
}

func promptNodesConfig(cfg *config.ClusterConfig) error {
	fmt.Println("🖧  节点配置")
	fmt.Println("------------")

	// 控制平面节点
	prompt := promptui.Prompt{
		Label:   "控制平面节点数量",
		Default: "3",
	}
	cpCount, err := prompt.Run()
	if err != nil {
		return err
	}
	cpNum, _ := strconv.Atoi(cpCount)

	cfg.Nodes.ControlPlanes = make([]config.NodeSpec, cpNum)
	for i := 0; i < cpNum; i++ {
		fmt.Printf("\n控制平面节点 %d:\n", i+1)
		node := config.NodeSpec{
			Role:   "controlplane",
			CPU:    2,
			Memory: 2048,
			Disk:   "20G",
		}

		prompt := promptui.Prompt{
			Label:   "VM ID",
			Default: fmt.Sprintf("%d", 101+i),
		}
		vmID, _ := prompt.Run()
		node.VMID, _ = strconv.Atoi(vmID)

		prompt = promptui.Prompt{
			Label:   "IP 地址",
			Default: fmt.Sprintf("192.168.1.%d", 101+i),
		}
		node.IPAddress, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "主机名",
			Default: fmt.Sprintf("talos-cp-%d", i+1),
		}
		node.Name, _ = prompt.Run()

		cfg.Nodes.ControlPlanes[i] = node
	}

	// 工作节点
	prompt = promptui.Prompt{
		Label:   "工作节点数量",
		Default: "2",
	}
	workerCount, err := prompt.Run()
	if err != nil {
		return err
	}
	workerNum, _ := strconv.Atoi(workerCount)

	cfg.Nodes.Workers = make([]config.NodeSpec, workerNum)
	for i := 0; i < workerNum; i++ {
		fmt.Printf("\n工作节点 %d:\n", i+1)
		node := config.NodeSpec{
			Role:   "worker",
			CPU:    4,
			Memory: 4096,
			Disk:   "50G",
		}

		prompt := promptui.Prompt{
			Label:   "VM ID",
			Default: fmt.Sprintf("%d", 201+i),
		}
		vmID, _ := prompt.Run()
		node.VMID, _ = strconv.Atoi(vmID)

		prompt = promptui.Prompt{
			Label:   "IP 地址",
			Default: fmt.Sprintf("192.168.1.%d", 201+i),
		}
		node.IPAddress, _ = prompt.Run()

		prompt = promptui.Prompt{
			Label:   "主机名",
			Default: fmt.Sprintf("talos-worker-%d", i+1),
		}
		node.Name, _ = prompt.Run()

		cfg.Nodes.Workers[i] = node
	}

	fmt.Println()
	return nil
}

func promptProxyConfig(cfg *config.ClusterConfig) error {
	fmt.Println("🌍 代理配置（可选）")
	fmt.Println("------------------")
	fmt.Println("如果你在中国或需要通过代理访问网络，请配置以下选项")
	fmt.Println()

	selectPrompt := promptui.Select{
		Label: "是否启用代理",
		Items: []string{"否", "是"},
	}
	_, result, err := selectPrompt.Run()
	if err != nil {
		return err
	}

	if result == "否" {
		cfg.Proxy.Enabled = false
		fmt.Println()
		return nil
	}

	cfg.Proxy.Enabled = true

	// HTTP 代理
	prompt := promptui.Prompt{
		Label:   "HTTP 代理地址（留空跳过）",
		Default: "",
	}
	httpProxy, err := prompt.Run()
	if err != nil {
		return err
	}
	if httpProxy != "" {
		cfg.Proxy.HTTPProxy = httpProxy
	}

	// HTTPS 代理
	prompt = promptui.Prompt{
		Label:   "HTTPS 代理地址（留空则使用 HTTP 代理）",
		Default: "",
	}
	httpsProxy, err := prompt.Run()
	if err != nil {
		return err
	}
	if httpsProxy != "" {
		cfg.Proxy.HTTPSProxy = httpsProxy
	} else if httpProxy != "" {
		cfg.Proxy.HTTPSProxy = httpProxy
	}

	// No Proxy
	prompt = promptui.Prompt{
		Label:   "不使用代理的地址（逗号分隔，留空跳过）",
		Default: "localhost,127.0.0.1,192.168.0.0/16,10.0.0.0/8",
	}
	noProxy, err := prompt.Run()
	if err != nil {
		return err
	}
	if noProxy != "" {
		cfg.Proxy.NoProxy = noProxy
	}

	// 镜像站 URL
	fmt.Println()
	fmt.Println("💡 提示：国内用户可以使用镜像站加速 Talos 镜像下载")
	fmt.Println("   常见镜像站：")
	fmt.Println("   - https://mirror.ghproxy.com/https://github.com/siderolabs/talos/releases/download")
	fmt.Println("   - https://ghproxy.com/https://github.com/siderolabs/talos/releases/download")
	fmt.Println()

	prompt = promptui.Prompt{
		Label:   "Talos 镜像下载地址（留空使用官方源）",
		Default: "",
	}
	mirrorURL, err := prompt.Run()
	if err != nil {
		return err
	}
	if mirrorURL != "" {
		cfg.Proxy.MirrorURL = mirrorURL
	}

	fmt.Println()
	return nil
}

func promptRegistryConfig(cfg *config.ClusterConfig) error {
	fmt.Println("🐳 容器镜像源配置（可选）")
	fmt.Println("------------------------")
	fmt.Println("配置国内镜像源可以大幅加速容器镜像拉取")
	fmt.Println()

	selectPrompt := promptui.Select{
		Label: "是否配置容器镜像源",
		Items: []string{"否", "是（推荐国内用户）"},
	}
	_, result, err := selectPrompt.Run()
	if err != nil {
		return err
	}

	if result == "否" {
		fmt.Println()
		return nil
	}

	// 初始化 Registry 配置
	cfg.Registry = &config.RegistryConfig{
		Mirrors: make(map[string]config.RegistryMirror),
	}

	// Docker Hub 镜像
	fmt.Println()
	fmt.Println("Docker Hub 镜像源（常用镜像源）:")
	fmt.Println("  1. docker.mirrors.ustc.edu.cn (中科大)")
	fmt.Println("  2. hub-mirror.c.163.com (网易)")
	fmt.Println()

	selectPrompt = promptui.Select{
		Label: "配置 Docker Hub 镜像源",
		Items: []string{"是", "否"},
	}
	_, dockerResult, err := selectPrompt.Run()
	if err != nil {
		return err
	}

	if dockerResult == "是" {
		cfg.Registry.Mirrors["docker.io"] = config.RegistryMirror{
			Endpoints: []string{
				"https://docker.mirrors.ustc.edu.cn",
				"https://hub-mirror.c.163.com",
			},
		}
	}

	// Google 容器镜像
	fmt.Println()
	selectPrompt = promptui.Select{
		Label: "配置 Google 容器镜像源（k8s.gcr.io, gcr.io）",
		Items: []string{"是（推荐）", "否"},
	}
	_, gcrResult, err := selectPrompt.Run()
	if err != nil {
		return err
	}

	if gcrResult == "是（推荐）" {
		cfg.Registry.Mirrors["k8s.gcr.io"] = config.RegistryMirror{
			Endpoints: []string{
				"https://registry.aliyuncs.com/google_containers",
			},
		}
		cfg.Registry.Mirrors["gcr.io"] = config.RegistryMirror{
			Endpoints: []string{
				"https://gcr.mirrors.ustc.edu.cn",
			},
		}
	}

	// GitHub 容器镜像
	fmt.Println()
	selectPrompt = promptui.Select{
		Label: "配置 GitHub 容器镜像源（ghcr.io）",
		Items: []string{"是", "否"},
	}
	_, ghcrResult, err := selectPrompt.Run()
	if err != nil {
		return err
	}

	if ghcrResult == "是" {
		cfg.Registry.Mirrors["ghcr.io"] = config.RegistryMirror{
			Endpoints: []string{
				"https://ghcr.nju.edu.cn",
			},
		}
	}

	// Quay 镜像
	fmt.Println()
	selectPrompt = promptui.Select{
		Label: "配置 Quay 镜像源（quay.io）",
		Items: []string{"是", "否"},
	}
	_, quayResult, err := selectPrompt.Run()
	if err != nil {
		return err
	}

	if quayResult == "是" {
		cfg.Registry.Mirrors["quay.io"] = config.RegistryMirror{
			Endpoints: []string{
				"https://quay.mirrors.ustc.edu.cn",
			},
		}
	}

	fmt.Println()
	return nil
}

func saveConfig(cfg *config.ClusterConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	filename := "cluster-config.yaml"
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("保存配置文件失败: %w", err)
	}

	fmt.Printf("✅ 配置已保存到: %s\n", filename)
	fmt.Println()
	fmt.Println("下一步:")
	fmt.Println("  1. 检查配置文件: cat cluster-config.yaml")
	fmt.Println("  2. 开始部署: talos-deployer deploy")

	return nil
}
