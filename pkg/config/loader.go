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
	return nil
}
