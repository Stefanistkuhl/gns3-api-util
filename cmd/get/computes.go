package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetComputesCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "computes",
		Short: "Get avaliable computes",
		Long:  `Get avaliable computes`,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			utils.ExecuteAndPrint(cfg, "getComputes", nil)
		},
	}
	return cmd
}

func NewGetComputeCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "info",
		Short: "Get a compute by name or id",
		Long:  `Get a compute by name or id`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "compute", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getCompute", []string{id})
		},
	}
	return cmd
}

func NewGetComputeDockerImagesCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "compute-docker-images",
		Short: "Get the docker-images of a compute by name or id",
		Long:  `Get the docker-images of a compute by name or id`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "compute", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getComputeDockerImgs", []string{id})
		},
	}
	return cmd
}

func NewGetComputeVirtualboxVMSCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "compute-virtualbox-vms",
		Short: "Get the virutalbox-vms of a compute by name or id",
		Long:  `Get the virutalbox-vms of a compute by name or id`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "compute", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getVirtualboxVms", []string{id})
		},
	}
	return cmd
}

func NewGetComputeVmWareVMSCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "compute-vmware-vms",
		Short: "Get the vmware-vms of a compute by name or id",
		Long:  `Get the vmware-vms of a compute by name or id`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "compute", args[0], nil)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			utils.ExecuteAndPrint(cfg, "getVmwareVms", []string{id})
		},
	}
	return cmd
}
