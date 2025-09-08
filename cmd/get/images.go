package get

import (
	"bytes"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/fuzzy"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/tidwall/gjson"
)

func NewGetImagesCmd() *cobra.Command {
	var (
		imageType string
	)
	cmd := &cobra.Command{
		Use:     utils.ListAllCmdName,
		Short:   "Get the available images on the Server",
		Long:    `Get the available images on the Server`,
		Example: "gns3util -s https://controller:3080 image ls",
		Run: func(cmd *cobra.Command, args []string) {
			if imageType != "" && imageType != "qemu" && imageType != "ios" && imageType != "iou" {
				fmt.Println("The image type can only be qemu, ios, or iou")
				return
			}
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			utils.ExecuteAndPrint(cfg, "getImages", []string{imageType})
		},
	}
	cmd.Flags().StringVarP(&imageType, "image-type", "t", "", "What type of image to get (qemu/ios/iou)")
	return cmd
}

func NewGetImageCmd() *cobra.Command {
	var useFuzzy bool
	var multi bool
	var imageType string
	var cmd = &cobra.Command{
		Use:     utils.ListSingleElementCmdName + " [image-path]",
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
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
			if err != nil {
				fmt.Printf("failed to get global options: %v", err)
			}
			if useFuzzy {
				rawData, _, err := utils.CallClient(cfg, "getImages", []string{imageType}, nil)
				if err != nil {
					fmt.Printf("Error getting images: %v\n", err)
					return
				}

				result := gjson.ParseBytes(rawData)
				if !result.IsArray() {
					fmt.Println("Expected array response")
					return
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
					fmt.Println("No images found")
					return
				}

				results := fuzzy.NewFuzzyFinder(filenames, multi)

				var selected []gjson.Result
				for _, result := range results {
				outer:
					for _, data := range apiData {
						if filename := data.Get("filename"); filename.Exists() && filename.String() == result {
							selected = append(selected, data)
							break outer
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
				if cfg.Raw {
					utils.PrintJson(toPrint)
				} else {
					utils.PrintKV(toPrint)
				}
			} else {
				path := args[0]
				utils.ExecuteAndPrint(cfg, "getImage", []string{path})
			}
		},
	}
	cmd.Flags().BoolVarP(&useFuzzy, "fuzzy", "f", false, "Use fuzzy search to find an image")
	cmd.Flags().BoolVarP(&multi, "multi", "m", false, "Get multiple images")
	cmd.Flags().StringVarP(&imageType, "image-type", "t", "", "What type of image to get (qemu/ios/iou)")
	return cmd
}
