package cmd

import (
	"fmt"

	"talos-proxmox-deployer/pkg/config"
	"talos-proxmox-deployer/pkg/deployer"

	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "验证集群状态",
	Long:  `检查集群健康状态和节点状态`,
	RunE:  runVerify,
}

func init() {
	verifyCmd.Flags().StringVarP(&configFile, "config", "c", "cluster-config.yaml", "配置文件路径")
}

func runVerify(cmd *cobra.Command, args []string) error {
	fmt.Println("🔍 验证集群状态")
	fmt.Println("================")
	fmt.Println()

	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	d := deployer.New(cfg)
	return d.Verify()
}
