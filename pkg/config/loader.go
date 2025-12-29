package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Load(filename string) (*ClusterConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg ClusterConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &cfg, nil
}

func (c *ClusterConfig) Validate() error {
	if c.ClusterName == "" {
		return fmt.Errorf("集群名称不能为空")
	}
	if c.TalosVersion == "" {
		return fmt.Errorf("Talos 版本不能为空")
	}
	if len(c.Nodes.ControlPlanes) < 1 {
		return fmt.Errorf("至少需要 1 个控制平面节点")
	}

	// 验证 Proxmox 认证配置
	if c.Proxmox.Password == "" && (c.Proxmox.APITokenID == "" || c.Proxmox.APIToken == "") {
		return fmt.Errorf("必须配置 Proxmox 认证信息：password 或 (api_token_id + api_token)")
	}

	// 如果同时配置了密码和 API Token，优先使用 API Token
	if c.Proxmox.Password != "" && c.Proxmox.APITokenID != "" {
		fmt.Println("⚠️  同时配置了密码和 API Token，将优先使用 API Token")
	}

	return nil
}
