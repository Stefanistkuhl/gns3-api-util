package get

import (
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
)

func NewGetComputesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.ListAllCmdName,
		Aliases: []string{"list", "l"},
		Short:   "Get available computes",
		Long:    `Get available computes`,
		Example: "gns3util -s https://controller:3080 compute ls",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			utils.ExecuteAndPrint(cfg, "getComputes", nil)
			return nil
		},
	}
	return cmd
}

func NewGetComputeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [compute-name/id]",
		Aliases: []string{"get", "i"},
		Short:   "Get a compute by name or id",
		Long:    `Get a compute by name or id`,
		Example: "gns3util -s https://controller:3080 compute info my-compute",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "compute", args[0], nil)
				if err != nil {
					return err
				}
			}
			utils.ExecuteAndPrint(cfg, "getCompute", []string{id})
			return nil
		},
	}
	return cmd
}

func NewGetComputeDockerImagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "docker-images [compute-name/id]",
		Aliases: []string{"docker", "images", "img"},
		Short:   "Get the docker-images of a compute by name or id",
		Long:    `Get the docker-images of a compute by name or id`,
		Example: "gns3util -s https://controller:3080 compute docker-images my-compute",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "compute", args[0], nil)
				if err != nil {
					return err
				}
			}
			utils.ExecuteAndPrint(cfg, "getComputeDockerImgs", []string{id})
			return nil
		},
	}
	return cmd
}

func NewGetComputeVirtualboxVMSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "virtualbox-vms [compute-name/id]",
		Aliases: []string{"vbox", "vbox-vms"},
		Short:   "Get the virtualbox-vms of a compute by name or id",
		Long:    `Get the virtualbox-vms of a compute by name or id`,
		Example: "gns3util -s https://controller:3080 compute virtualbox-vms my-compute",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "compute", args[0], nil)
				if err != nil {
					return err
				}
			}
			utils.ExecuteAndPrint(cfg, "getVirtualboxVms", []string{id})
			return nil
		},
	}
	return cmd
}

func NewGetComputeVmWareVMSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "vmware-vms [compute-name/id]",
		Aliases: []string{"vmw", "vmware"},
		Short:   "Get the vmware-vms of a compute by name or id",
		Long:    `Get the vmware-vms of a compute by name or id`,
		Example: "gns3util -s https://controller:3080 compute vmware-vms my-compute",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			if !utils.IsValidUUIDv4(args[0]) {
				id, err = utils.ResolveID(cfg, "compute", args[0], nil)
				if err != nil {
					return err
				}
			}
			utils.ExecuteAndPrint(cfg, "getVmwareVms", []string{id})
			return nil
		},
	}
	return cmd
}
