package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "talos-deployer",
	Short: "Talos Linux + Proxmox VE Kubernetes 集群部署工具",
	Long:  `一个用于在 Proxmox VE 上自动化部署 Talos Linux Kubernetes 集群的工具`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(destroyCmd)
	rootCmd.AddCommand(manageCmd)
}
