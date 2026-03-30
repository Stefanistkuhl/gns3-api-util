package objstorecmd

import "github.com/spf13/cobra"

func NewObjCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "object-store",
		Aliases: []string{"obj", "object"},
		Short:   "Commands to interact with the gns3util cluster object storage",
		Long:    "Commands to interact with the gns3util cluster object storage",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Help()
			return nil
		},
	}

	objectGroup := &cobra.Group{ID: "object", Title: "Object & File Operations:"}
	bucketGroup := &cobra.Group{ID: "bucket", Title: "Bucket Operations:"}
	accessGroup := &cobra.Group{ID: "access", Title: "Access Management:"}

	cmd.AddGroup(objectGroup, bucketGroup, accessGroup)

	uploadCmd := NewUploadCmd()
	uploadCmd.GroupID = "object"

	uploadToBucketCmd := NewUploadToBucketCmd()
	uploadToBucketCmd.GroupID = "object"

	downloadCmd := NewDownloadCmd()
	downloadCmd.GroupID = "object"

	deleteFileCmd := NewDeleteFileCmd()
	deleteFileCmd.GroupID = "object"

	createBucketCmd := NewCreateBucketCmd()
	createBucketCmd.GroupID = "bucket"

	listBucketsCmd := NewListBucketsCmd()
	listBucketsCmd.GroupID = "bucket"

	listBucketFilesCmd := NewListBucketFilesCmd()
	listBucketFilesCmd.GroupID = "bucket"

	deleteBucketCmd := NewDeleteBucketCmd()
	deleteBucketCmd.GroupID = "bucket"

	genTokenCmd := NewGeneratePublicTokenCmd()
	genTokenCmd.GroupID = "access"

	bucketPermsCmd := NewBucketPermissionsCmd()
	bucketPermsCmd.GroupID = "access"

	filePermsCmd := NewFilePermissionsCmd()
	filePermsCmd.GroupID = "access"

	cmd.AddCommand(
		uploadCmd,
		uploadToBucketCmd,
		downloadCmd,
		deleteFileCmd,
		createBucketCmd,
		listBucketsCmd,
		listBucketFilesCmd,
		deleteBucketCmd,
		genTokenCmd,
		bucketPermsCmd,
		filePermsCmd,
	)

	return cmd
}
