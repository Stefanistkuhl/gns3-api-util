package class

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/config"
	"github.com/stefanistkuhl/gns3util/pkg/utils/class"
	"github.com/stefanistkuhl/gns3util/pkg/utils/colorUtils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/server"
)

func NewClassCreateCmd() *cobra.Command {
	var createClassCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a class with students and groups",
		Long: `Create a class with students and groups. This command can either:
- Create a class from a JSON file
- Launch an interactive web interface for class creation

The class structure includes:
- A main class group
- Student groups within the class
- Students assigned to both the class group and their respective student groups`,
		Example: `
  # Create class from JSON file
  gns3util -s https://controller:3080 class create --file class.json

  # Launch interactive class creation
  gns3util -s https://controller:3080 class create --interactive

  # Create class with specific name
  gns3util -s https://controller:3080 class create --file class.json --name "CS101"
		`,
		RunE: runCreateClass,
	}

	createClassCmd.Flags().String("file", "", "JSON file containing class data")
	createClassCmd.Flags().Bool("interactive", false, "Launch interactive web interface for class creation")
	createClassCmd.Flags().String("name", "", "Override class name from file")
	createClassCmd.Flags().Int("port", 8080, "Port for interactive web interface")
	createClassCmd.Flags().String("host", "localhost", "Host for interactive web interface")

	return createClassCmd
}

func runCreateClass(cmd *cobra.Command, args []string) error {
	cfg, err := config.GetGlobalOptionsFromContext(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get global options: %w", err)
	}

	filePath, _ := cmd.Flags().GetString("file")
	interactive, _ := cmd.Flags().GetBool("interactive")
	className, _ := cmd.Flags().GetString("name")
	port, _ := cmd.Flags().GetInt("port")
	host, _ := cmd.Flags().GetString("host")

	if filePath == "" && !interactive {
		return fmt.Errorf("either --file or --interactive must be specified")
	}

	if filePath != "" && interactive {
		return fmt.Errorf("cannot specify both --file and --interactive")
	}

	var classData schemas.Class

	if interactive {
		var err error
		classData, err = server.StartInteractiveServer(host, port)
		if err != nil {
			return fmt.Errorf("failed to start interactive server: %w", err)
		}
	} else {
		var err error
		classData, err = class.LoadClassFromFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to load class from file: %w", err)
		}
	}

	if className != "" {
		classData.Name = className
	}

	success, err := class.CreateClass(cfg, classData)
	if err != nil {
		return fmt.Errorf("failed to create class: %w", err)
	}

	if success {
		fmt.Printf("%v Created class %v\n",
			colorUtils.Success("Success:"),
			colorUtils.Bold(classData.Name))
	} else {
		fmt.Printf("%v Class creation failed\n", colorUtils.Error("Error:"))
		return fmt.Errorf("class creation failed")
	}

	return nil
}
