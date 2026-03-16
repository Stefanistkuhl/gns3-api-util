package objstorecmd

import "github.com/spf13/cobra"

func NewObjCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "object-store",
		Aliases: []string{"obj", "object"},
		Short:   "Commands to interact with the gns3util cluster object storrage",
		Long:    "Commands to interact with the gns3util cluster object storrage",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return nil
		},
	}
	cmd.AddCommand(NewUploadCmd())
	cmd.AddCommand(NewDeleteFileCmd())
	cmd.AddCommand(NewDownloadCmd())
	cmd.AddCommand(NewListBucketsCmd())
	cmd.AddCommand(NewListBucketFilesCmd())
	cmd.AddCommand(NewGeneratePublicTokenCmd())
	cmd.AddCommand(NewUploadToBucketCmd())
	cmd.AddCommand(NewDeleteBucketCmd())
	return cmd
}
