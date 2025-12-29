package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"talos-proxmox-deployer/pkg/config"
	"talos-proxmox-deployer/pkg/deployer"
)

var manageCmd = &cobra.Command{
	Use:   "manage",
	Short: "管理集群",
	Long:  `启动、停止、重启集群节点`,
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "启动集群节点",
	RunE:  runStart,
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "停止集群节点",
	RunE:  runStop,
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "重启集群节点",
	RunE:  runRestart,
}

func init() {
	manageCmd.AddCommand(startCmd)
	manageCmd.AddCommand(stopCmd)
	manageCmd.AddCommand(restartCmd)

	startCmd.Flags().StringVarP(&configFile, "config", "c", "cluster-config.yaml", "配置文件路径")
	stopCmd.Flags().StringVarP(&configFile, "config", "c", "cluster-config.yaml", "配置文件路径")
	restartCmd.Flags().StringVarP(&configFile, "config", "c", "cluster-config.yaml", "配置文件路径")
}

func runStart(cmd *cobra.Command, args []string) error {
	fmt.Println("▶️  启动集群节点")
	cfg, err := config.Load(configFile)
	if err != nil {
		return err
	}
	d := deployer.New(cfg)
	return d.StartNodes()
}

func runStop(cmd *cobra.Command, args []string) error {
	fmt.Println("⏸️  停止集群节点")
	cfg, err := config.Load(configFile)
	if err != nil {
		return err
	}
	d := deployer.New(cfg)
	return d.StopNodes()
}

func runRestart(cmd *cobra.Command, args []string) error {
	fmt.Println("🔄 重启集群节点")
	cfg, err := config.Load(configFile)
	if err != nil {
		return err
	}
	d := deployer.New(cfg)
	if err := d.StopNodes(); err != nil {
		return err
	}
	return d.StartNodes()
}
