package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"talos-proxmox-deployer/pkg/config"
	"talos-proxmox-deployer/pkg/deployer"

	"github.com/spf13/cobra"
)

var forceDestroy bool

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "销毁集群",
	Long:  `停止并删除所有集群节点和配置`,
	RunE:  runDestroy,
}

func init() {
	destroyCmd.Flags().StringVarP(&configFile, "config", "c", "cluster-config.yaml", "配置文件路径")
	destroyCmd.Flags().BoolVarP(&forceDestroy, "force", "f", false, "强制销毁，不询问确认")
}

func runDestroy(cmd *cobra.Command, args []string) error {
	fmt.Println("⚠️  销毁集群")
	fmt.Println("============")
	fmt.Println()

	if !forceDestroy {
		fmt.Print("确定要销毁整个集群吗？这不可恢复！(yes/no): ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer != "yes" {
			fmt.Println("操作已取消")
			return nil
		}
	}

	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	d := deployer.New(cfg)
	if err := d.Destroy(); err != nil {
		return fmt.Errorf("销毁集群失败: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ 集群已销毁")
	return nil
}
