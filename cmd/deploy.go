package cmd

import (
	"fmt"

	"talos-proxmox-deployer/pkg/config"
	"talos-proxmox-deployer/pkg/deployer"

	"github.com/spf13/cobra"
)

var (
	configFile    string
	skipPrepare   bool
	skipTemplate  bool
	skipConfig    bool
	skipBootstrap bool
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "部署 Talos 集群",
	Long:  `根据配置文件部署 Talos Kubernetes 集群到 Proxmox VE`,
	RunE:  runDeploy,
}

func init() {
	deployCmd.Flags().StringVarP(&configFile, "config", "c", "cluster-config.yaml", "配置文件路径")
	deployCmd.Flags().BoolVarP(&skipPrepare, "skip-prepare", "s", false, "跳过镜像准备")
	deployCmd.Flags().BoolVarP(&skipTemplate, "skip-template", "t", false, "跳过模板创建")
	deployCmd.Flags().BoolVar(&skipConfig, "skip-config", false, "跳过配置生成")
	deployCmd.Flags().BoolVar(&skipBootstrap, "skip-bootstrap", false, "跳过集群引导")
}

func runDeploy(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 开始部署 Talos 集群")
	fmt.Println("========================")
	fmt.Println()

	// 加载配置
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	fmt.Printf("集群名称: %s\n", cfg.ClusterName)
	fmt.Printf("Talos 版本: %s\n", cfg.TalosVersion)
	fmt.Printf("控制平面节点: %d\n", len(cfg.Nodes.ControlPlanes))
	fmt.Printf("工作节点: %d\n", len(cfg.Nodes.Workers))
	fmt.Println()

	// 创建部署器
	d := deployer.New(cfg)

	// 执行部署步骤
	if !skipPrepare {
		if err := d.PrepareImage(); err != nil {
			return fmt.Errorf("准备镜像失败: %w", err)
		}
	}

	if !skipTemplate {
		if err := d.CreateTemplate(); err != nil {
			return fmt.Errorf("创建模板失败: %w", err)
		}
	}

	if err := d.CreateNodes(); err != nil {
		return fmt.Errorf("创建节点失败: %w", err)
	}

	if !skipConfig {
		if err := d.GenerateConfig(); err != nil {
			return fmt.Errorf("生成配置失败: %w", err)
		}
	}

	if err := d.ApplyConfig(); err != nil {
		return fmt.Errorf("应用配置失败: %w", err)
	}

	if !skipBootstrap {
		if err := d.Bootstrap(); err != nil {
			return fmt.Errorf("引导集群失败: %w", err)
		}
	}

	fmt.Println()
	fmt.Println("✅ 部署完成!")
	fmt.Println()
	fmt.Println("下一步:")
	fmt.Printf("  export KUBECONFIG=$(pwd)/%s-config/kubeconfig\n", cfg.ClusterName)
	fmt.Println("  kubectl get nodes")

	return nil
}
