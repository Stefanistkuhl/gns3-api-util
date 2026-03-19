package get

import (
	"bytes"
	"fmt"

	"github.com/0xveya/gns3util/internal/cli/cli_pkg/config"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/fuzzy"
	"github.com/0xveya/gns3util/internal/cli/cli_pkg/utils"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

func NewGetImagesCmd() *cobra.Command {
	var imageType string
	cmd := &cobra.Command{
		Use:     utils.ListAllCmdName,
		Aliases: []string{"list", "l"},
		Short:   "Get the available images on the Server",
		Long:    `Get the available images on the Server`,
		Example: "gns3util -s https://controller:3080 image ls",
		RunE: func(cmd *cobra.Command, args []string) error {
			if imageType != "" && imageType != "qemu" && imageType != "ios" && imageType != "iou" {
				return fmt.Errorf("the image type can only be qemu, ios, or iou")
			}
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			utils.ExecuteAndPrint(cfg, "getImages", []string{imageType})
			return nil
		},
	}
	cmd.Flags().StringVarP(&imageType, "image-type", "t", "", "What type of image to get (qemu/ios/iou)")
	return cmd
}

func NewGetImageCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	var imageType string
	cmd := &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [image-path]",
		Aliases: []string{"get", "i"},
		Short:   "Get an image by path",
		Long:    `Get an image by path`,
		Example: "gns3util -s https://controller:3080 image info /path/to/image",
		Args: func(cmd *cobra.Command, args []string) error {
			if useFuzzy {
				if len(args) > 1 {
					return fmt.Errorf("at most 1 positional arg allowed when --fuzzy is set")
				}
				return nil
			}
			if len(args) != 1 {
				return fmt.Errorf("requires 1 arg [image-path] when --fuzzy is not set")
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if multi && !useFuzzy {
				return fmt.Errorf("the --multi (-m) flag can only be used together with --fuzzy (-f)")
			}
			if imageType != "" && imageType != "qemu" && imageType != "ios" && imageType != "iou" {
				return fmt.Errorf("the image type can only be qemu, ios, or iou")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to get global options: %w", err)
			}
			if useFuzzy {
				rawData, _, err := utils.CallClient(cfg, "getImages", []string{imageType}, nil)
				if err != nil {
					return fmt.Errorf("error getting images: %w", err)
				}

				result := gjson.ParseBytes(rawData)
				if !result.IsArray() {
					return fmt.Errorf("expected array response")
				}

				var filenames []string
				var apiData []gjson.Result

				result.ForEach(func(_, value gjson.Result) bool {
					apiData = append(apiData, value)
					if filename := value.Get("filename"); filename.Exists() {
						filenames = append(filenames, filename.String())
					}
					return true
				})

				if len(filenames) == 0 {
					return fmt.Errorf("no images found")
				}

				results := fuzzy.NewFuzzyFinder(filenames, multi)

				var selected []gjson.Result
				for _, result := range results {
					for _, data := range apiData {
						if filename := data.Get("filename"); filename.Exists() && filename.String() == result {
							selected = append(selected, data)
							break
						}
					}
				}

				var buf bytes.Buffer
				buf.WriteByte('[')
				for i, s := range selected {
					if i > 0 {
						buf.WriteByte(',')
					}
					buf.WriteString(s.Raw)
				}
				buf.WriteByte(']')

				toPrint := buf.Bytes()
				utils.PrintOutput(toPrint, cfg)
			} else {
				path := args[0]
				utils.ExecuteAndPrint(cfg, "getImage", []string{path})
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find an image")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple images")
	cmd.Flags().StringVarP(&imageType, "image-type", "t", "", "What type of image to get (qemu/ios/iou)")
	return cmd
}
