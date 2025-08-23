package get

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
)

func NewGetComputesCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     utils.ListAllCmdName,
		Short:   "Get available computes",
		Long:    `Get available computes`,
		Example: "gns3util -s https://controller:3080 compute ls",
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
		Use:     utils.ListSingleElementCmdName + " [compute-name/id]",
		Short:   "Get a compute by name or id",
		Long:    `Get a compute by name or id`,
		Example: "gns3util -s https://controller:3080 compute info my-compute",
		Args:    cobra.ExactArgs(1),
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
		Use:     "docker-images [compute-name/id]",
		Short:   "Get the docker-images of a compute by name or id",
		Long:    `Get the docker-images of a compute by name or id`,
		Example: "gns3util -s https://controller:3080 compute docker-images my-compute",
		Args:    cobra.ExactArgs(1),
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
		Use:     "virtualbox-vms [compute-name/id]",
		Short:   "Get the virtualbox-vms of a compute by name or id",
		Long:    `Get the virtualbox-vms of a compute by name or id`,
		Example: "gns3util -s https://controller:3080 compute virtualbox-vms my-compute",
		Args:    cobra.ExactArgs(1),
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
		Use:     "vmware-vms [compute-name/id]",
		Short:   "Get the vmware-vms of a compute by name or id",
		Long:    `Get the vmware-vms of a compute by name or id`,
		Example: "gns3util -s https://controller:3080 compute vmware-vms my-compute",
		Args:    cobra.ExactArgs(1),
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
