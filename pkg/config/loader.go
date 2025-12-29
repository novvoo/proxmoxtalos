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
	switch c.Proxmox.AuthMethod {
	case "password":
		if c.Proxmox.Password == "" {
			return fmt.Errorf("auth_method 设置为 password，但未配置 password 字段")
		}
	case "api_token":
		if c.Proxmox.APITokenID == "" || c.Proxmox.APIToken == "" {
			return fmt.Errorf("auth_method 设置为 api_token，但未配置 api_token_id 或 api_token 字段")
		}
	case "":
		// 兼容旧配置：如果没有指定 auth_method，检查是否至少配置了一种认证方式
		if c.Proxmox.Password == "" && (c.Proxmox.APITokenID == "" || c.Proxmox.APIToken == "") {
			return fmt.Errorf("必须配置 Proxmox 认证信息：设置 auth_method 并配置相应的认证凭据")
		}
		fmt.Println("⚠️  警告: 未指定 auth_method，建议明确设置为 'password' 或 'api_token'")
	default:
		return fmt.Errorf("无效的 auth_method: %s，必须是 'password' 或 'api_token'", c.Proxmox.AuthMethod)
	}

	return nil
}
