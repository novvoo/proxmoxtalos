package config

type ClusterConfig struct {
	ClusterName       string          `yaml:"cluster_name"`
	TalosVersion      string          `yaml:"talos_version"`
	KubernetesVersion string          `yaml:"kubernetes_version"`
	Network           NetworkConfig   `yaml:"network"`
	Proxmox           ProxmoxConfig   `yaml:"proxmox"`
	Nodes             NodesConfig     `yaml:"nodes"`
	Proxy             ProxyConfig     `yaml:"proxy,omitempty"`
	Registry          *RegistryConfig `yaml:"registry,omitempty"` // 容器镜像仓库配置
}

type NetworkConfig struct {
	Bridge    string `yaml:"bridge"`
	DNSServer string `yaml:"dns_server"`
	Gateway   string `yaml:"gateway"`
	Netmask   string `yaml:"netmask"`
}

type ProxmoxConfig struct {
	Host         string `yaml:"host"`
	User         string `yaml:"user"`
	StoragePool  string `yaml:"storage_pool"`
	TemplateVMID int    `yaml:"template_vm_id"`
}

type NodesConfig struct {
	ControlPlanes []NodeSpec `yaml:"control_planes"`
	Workers       []NodeSpec `yaml:"workers"`
}

type NodeSpec struct {
	VMID      int    `yaml:"vm_id"`
	IPAddress string `yaml:"ip_address"`
	Name      string `yaml:"name"`
	CPU       int    `yaml:"cpu"`
	Memory    int    `yaml:"memory"`
	Disk      string `yaml:"disk"`
	Role      string `yaml:"role"`
}

type ProxyConfig struct {
	Enabled    bool   `yaml:"enabled"`
	HTTPProxy  string `yaml:"http_proxy,omitempty"`
	HTTPSProxy string `yaml:"https_proxy,omitempty"`
	NoProxy    string `yaml:"no_proxy,omitempty"`
	MirrorURL  string `yaml:"mirror_url,omitempty"` // 用于 Talos 镜像下载的镜像站
}

// RegistryConfig 容器镜像仓库配置
type RegistryConfig struct {
	Mirrors map[string]RegistryMirror `yaml:"mirrors,omitempty"` // 镜像源配置
}

// RegistryMirror 镜像源配置
type RegistryMirror struct {
	Endpoints []string `yaml:"endpoints"` // 镜像源地址列表
}
