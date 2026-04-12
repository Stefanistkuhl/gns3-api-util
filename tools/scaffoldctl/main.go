// this will never work on windows because i dont care enough
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/0xveya/gns3util/tools/scaffoldctl/create"
	"github.com/0xveya/gns3util/tools/scaffoldctl/diff"
	"github.com/0xveya/gns3util/tools/scaffoldctl/rgjson"
	"github.com/0xveya/gns3util/tools/scaffoldctl/spec"
)

type Config struct {
	SpecPath     string
	SpecDir      string
	TemplatesDir string
	DryRun       bool
	Write        bool
	Force        bool
	Verbose      bool
	ListSections bool
	ListPackages bool
	ListClient   bool
	ListCLI      bool
	ListExisting bool
	ListPlan     bool
	Check        bool
	NoPager      bool
	SkipExisting bool
	Mode         string
	Only         []string
}

type UsefulFileLocations struct{}

const (
	defaultSpecDir      = ".codegen/scaffoldctl/packages"
	defaultTemplatesDir = "tools/scaffoldctl/templates"
)

var clientSections = []string{
	"transport",
	"auth",
	"cluster",
	"users",
	"roles",
	"jobs",
	"buckets",
	"files",
	"backups",
	"vm_images",
	"bucket_permissions",
	"file_permissions",
	"public_tokens",
	"transfers",
}

func init() {
	if runtime.GOOS == "windows" {
		fmt.Println("this doenst work on windows since i hardcoded forward slashes please run this tool via wsl or use a real os to code")
		os.Exit(67)
	}
}

func main() {
	log.SetFlags(0)

	cfg, showHelp, err := parseFlags(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	if showHelp {
		printUsage(os.Stdout)
		return
	}

	if cfg.ListSections {
		printSections()
		return
	}

	if err := run(&cfg); err != nil {
		log.Fatal(err)
	}
}

func parseFlags(args []string) (Config, bool, error) {
	var cfg Config
	fs := flag.NewFlagSet("scaffoldctl", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	fs.StringVar(&cfg.SpecPath, "spec", "", "Path to the YAML scaffold spec.")
	fs.StringVar(&cfg.SpecDir, "spec-dir", defaultSpecDir, "Directory containing scaffold specs for auto-discovery.")
	fs.StringVar(&cfg.TemplatesDir, "templates-dir", defaultTemplatesDir, "Directory containing Go templates.")
	fs.BoolVar(&cfg.DryRun, "dry-run", false, "Render and validate without writing files.")
	fs.BoolVar(&cfg.Write, "write", false, "Write generated changes to target files and gofmt changed Go files.")
	fs.BoolVar(&cfg.Force, "force", false, "Allow overwriting scaffolded editable files.")
	fs.BoolVar(&cfg.Verbose, "v", false, "Enable verbose logging.")
	fs.BoolVar(&cfg.ListSections, "list-sections", false, "Print supported ClientV2 section keys and exit.")
	fs.BoolVar(&cfg.ListPackages, "list-packages", false, "Print discovered scaffold packages and exit.")
	fs.BoolVar(&cfg.ListClient, "list-client", false, "Print configured client methods and exit.")
	fs.BoolVar(&cfg.ListCLI, "list-cli", false, "Print configured CLI commands and exit.")
	fs.BoolVar(&cfg.ListExisting, "list-existing", false, "Print existing ClientV2 methods in target files and exit.")
	fs.BoolVar(&cfg.ListPlan, "list-plan", false, "Print the client and CLI generation plan.")
	fs.BoolVar(&cfg.Check, "check", false, "Validate specs, referenced models, and scaffold ownership without rendering a diff.")
	fs.BoolVar(&cfg.NoPager, "no-pager", false, "Use plain diff output instead of piping through delta.")
	fs.BoolVar(&cfg.SkipExisting, "skip-existing", false, "Skip configured client methods that already exist in the target file.")
	fs.StringVar(&cfg.Mode, "mode", "all", "Generation mode: all, client, cli.")

	var onlyRaw string
	fs.StringVar(&onlyRaw, "only", "", "Comma-separated IDs to generate, for example: list_roles,get_role.")

	fs.Usage = func() { printFlagSetUsage(fs, fs.Output()) }

	if len(args) == 0 {
		return Config{}, true, nil
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(2)
		}
		if suggestion := suggestFlagName(err.Error(), fs); suggestion != "" {
			return Config{}, false, fmt.Errorf("%s\n\ndid you mean -%s?", strings.TrimSpace(err.Error()), suggestion)
		}
		return Config{}, false, err
	}
	if cfg.DryRun && cfg.Write {
		return Config{}, false, fmt.Errorf("-dry-run and -write cannot be used together")
	}

	cfg.Mode = strings.TrimSpace(strings.ToLower(cfg.Mode))
	switch cfg.Mode {
	case "all", "client", "cli":
	default:
		return Config{}, false, fmt.Errorf("invalid -mode %q: must be one of all, client, cli", cfg.Mode)
	}

	if onlyRaw != "" {
		for part := range strings.SplitSeq(onlyRaw, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			cfg.Only = append(cfg.Only, part)
		}
	}

	if cfg.ListSections {
		return cfg, false, nil
	}

	cfg.SpecPath = strings.TrimSpace(cfg.SpecPath)
	cfg.SpecDir = strings.TrimSpace(cfg.SpecDir)
	cfg.TemplatesDir = strings.TrimSpace(cfg.TemplatesDir)

	if cfg.SpecPath == "" && cfg.SpecDir == "" {
		cfg.SpecDir = defaultSpecDir
	}

	if cfg.TemplatesDir == "" {
		return Config{}, false, fmt.Errorf("-templates-dir must not be empty")
	}

	return cfg, false, nil
}

func run(cfg *Config) error {
	if cfg.SpecPath != "" {
		if _, err := os.Stat(cfg.SpecPath); err != nil {
			return fmt.Errorf("failed to stat spec %q: %w", cfg.SpecPath, err)
		}
	} else {
		if _, err := os.Stat(cfg.SpecDir); err != nil {
			return fmt.Errorf("failed to stat spec dir %q: %w", cfg.SpecDir, err)
		}
	}

	if _, err := os.Stat(cfg.TemplatesDir); err != nil {
		return fmt.Errorf("failed to stat templates dir %q: %w", cfg.TemplatesDir, err)
	}

	if cfg.Verbose {
		if cfg.SpecPath != "" {
			logValue("spec", cfg.SpecPath)
		} else {
			logValue("spec-dir", cfg.SpecDir)
		}
		logValue("templates-dir", cfg.TemplatesDir)
		logValue("mode", cfg.Mode)
		logBool("dry-run", cfg.DryRun)
		logBool("write", cfg.Write)
		logBool("force", cfg.Force)
		logBool("skip-existing", cfg.SkipExisting)
		if len(cfg.Only) > 0 {
			logValue("only", strings.Join(cfg.Only, ", "))
		}
	}

	if cfg.ListPackages || cfg.ListClient || cfg.ListCLI || cfg.ListExisting || cfg.ListPlan || cfg.Check || cfg.DryRun || cfg.Write {
		specs, err := loadConfiguredSpecs(cfg)
		if err != nil {
			return err
		}

		only := onlySet(cfg.Only)

		if cfg.ListPackages {
			printPackages(specs)
			return nil
		}
		if cfg.ListClient {
			printClientMethods(specs, only)
			return nil
		}
		if cfg.ListCLI {
			printCLICommands(specs, only)
			return nil
		}
		if cfg.ListExisting {
			return printExistingMethods(specs)
		}
		if cfg.ListPlan {
			return printGenerationPlan(specs, only, cfg.SkipExisting)
		}
		if cfg.Check {
			if checkErr := checkSpecs(specs, only, cfg.SkipExisting || cfg.Mode == "cli"); checkErr != nil {
				return checkErr
			}
			if err := printCheckSummary(specs, only, cfg.SkipExisting); err != nil {
				return err
			}
			fmt.Println("scaffoldctl: check OK")
			return nil
		}

		if checkErr := checkSpecs(specs, only, cfg.SkipExisting || cfg.Mode == "cli"); checkErr != nil {
			return checkErr
		}

		var changedGoFiles []string
		if cfg.Mode == "all" || cfg.Mode == "client" {
			builds, err := buildClientFiles(specs, only, cfg.TemplatesDir, cfg.SkipExisting)
			if err != nil {
				return err
			}

			for _, build := range builds {
				generated, err := create.BuildFinalFile(build)
				if err != nil {
					return fmt.Errorf("build final file %s: %w", build.TargetFile, err)
				}
				generated, err = formatGeneratedSource(build.TargetFile, generated)
				if err != nil {
					return err
				}

				if err := diff.DiffFileWithOptions(build.TargetFile, generated, diff.Options{NoPager: cfg.NoPager}); err != nil {
					return fmt.Errorf("diff %s: %w", build.TargetFile, err)
				}

				if cfg.Write {
					if err := writeGeneratedFile(build.TargetFile, generated); err != nil {
						return fmt.Errorf("write %s: %w", build.TargetFile, err)
					}
					fmt.Printf("wrote %s\n", build.TargetFile)
				}
			}
			changedGoFiles = append(changedGoFiles, buildTargetFiles(builds)...)
		}

		if cfg.Mode == "all" || cfg.Mode == "cli" {
			cliFiles, err := buildCLIFiles(specs, only, cfg.TemplatesDir)
			if err != nil {
				return err
			}
			cliParentFiles, err := buildCLIParentFiles(specs, only)
			if err != nil {
				return err
			}

			for targetFile, generated := range cliFiles {
				generated, err = formatGeneratedSource(targetFile, generated)
				if err != nil {
					return err
				}
				if diffErr := diff.DiffFileWithOptions(targetFile, generated, diff.Options{NoPager: cfg.NoPager}); diffErr != nil {
					return fmt.Errorf("diff %s: %w", targetFile, diffErr)
				}

				if cfg.Write {
					if genErr := writeGeneratedNewFile(targetFile, generated, cfg.Force); genErr != nil {
						return genErr
					}
					fmt.Printf("wrote %s\n", targetFile)
				}
			}
			changedGoFiles = append(changedGoFiles, sortedMapKeys(cliFiles)...)

			for targetFile, generated := range cliParentFiles {
				generated, err = formatGeneratedSource(targetFile, generated)
				if err != nil {
					return err
				}
				if err := diff.DiffFileWithOptions(targetFile, generated, diff.Options{NoPager: cfg.NoPager}); err != nil {
					return fmt.Errorf("diff %s: %w", targetFile, err)
				}

				if cfg.Write {
					if err := writeGeneratedFile(targetFile, generated); err != nil {
						return fmt.Errorf("write %s: %w", targetFile, err)
					}
					fmt.Printf("updated %s\n", targetFile)
				}
			}
			changedGoFiles = append(changedGoFiles, sortedMapKeys(cliParentFiles)...)
		}

		if cfg.Write {
			if err := formatGoFiles(uniqueStrings(changedGoFiles)); err != nil {
				return err
			}
		}

	}
	fmt.Println("scaffoldctl: bootstrap OK")
	return nil
}

func printSections() {
	fmt.Println("Supported ClientV2 section keys:")
	for _, section := range clientSections {
		fmt.Printf("  - %s\n", section)
	}
}

func logValue(label, value string) {
	_, _ = fmt.Fprintf(os.Stderr, "%s: %q\n", label, value)
}

func logBool(label string, value bool) {
	_, _ = fmt.Fprintf(os.Stderr, "%s: %t\n", label, value)
}

func printUsage(w *os.File) {
	fs := flag.NewFlagSet("scaffoldctl", flag.ContinueOnError)

	var cfg Config
	var onlyRaw string

	fs.StringVar(&cfg.SpecPath, "spec", "", "Path to the YAML scaffold spec.")
	fs.StringVar(&cfg.SpecDir, "spec-dir", defaultSpecDir, "Directory containing scaffold specs for auto-discovery.")
	fs.StringVar(&cfg.TemplatesDir, "templates-dir", defaultTemplatesDir, "Directory containing Go templates.")
	fs.BoolVar(&cfg.DryRun, "dry-run", false, "Render and validate without writing files.")
	fs.BoolVar(&cfg.Write, "write", false, "Write generated changes to target files and gofmt changed Go files.")
	fs.BoolVar(&cfg.Force, "force", false, "Allow overwriting scaffolded editable files.")
	fs.BoolVar(&cfg.Verbose, "v", false, "Enable verbose logging.")
	fs.BoolVar(&cfg.ListSections, "list-sections", false, "Print supported ClientV2 section keys and exit.")
	fs.BoolVar(&cfg.ListPackages, "list-packages", false, "Print discovered scaffold packages and exit.")
	fs.BoolVar(&cfg.ListClient, "list-client", false, "Print configured client methods and exit.")
	fs.BoolVar(&cfg.ListCLI, "list-cli", false, "Print configured CLI commands and exit.")
	fs.BoolVar(&cfg.ListExisting, "list-existing", false, "Print existing ClientV2 methods in target files and exit.")
	fs.BoolVar(&cfg.ListPlan, "list-plan", false, "Print the client and CLI generation plan.")
	fs.BoolVar(&cfg.Check, "check", false, "Validate specs, referenced models, and scaffold ownership without rendering a diff.")
	fs.BoolVar(&cfg.NoPager, "no-pager", false, "Use plain diff output instead of piping through delta.")
	fs.BoolVar(&cfg.SkipExisting, "skip-existing", false, "Skip configured client methods that already exist in the target file.")
	fs.StringVar(&cfg.Mode, "mode", "all", "Generation mode: all, client, cli.")
	fs.StringVar(&onlyRaw, "only", "", "Comma-separated IDs to generate, for example: list_roles,get_role.")
	printFlagSetUsage(fs, w)
}

func printFlagSetUsage(fs *flag.FlagSet, w interface{ Write([]byte) (int, error) }) {
	fmt.Fprintf(w, "Usage: %s [flags]\n\n", fs.Name())
	_, _ = fmt.Fprintln(w, "scaffoldctl bootstraps editable client and CLI scaffolding from YAML specs.")
	_, _ = fmt.Fprintln(w)
	fmt.Fprintf(w, "Default spec discovery: %s%s*.yaml%s\n", defaultSpecDir, string(os.PathSeparator), "")
	fmt.Fprintf(w, "Default templates dir: %s\n", defaultTemplatesDir)
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Flags:")
	fs.SetOutput(w)
	fs.PrintDefaults()
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Examples:")
	fmt.Fprintf(w, "  %s -dry-run\n", fs.Name())
	fmt.Fprintf(w, "  %s -dry-run -no-pager\n", fs.Name())
	fmt.Fprintf(w, "  %s -write -skip-existing\n", fs.Name())
	fmt.Fprintf(w, "  %s -check\n", fs.Name())
	fmt.Fprintf(w, "  %s -list-plan -skip-existing\n", fs.Name())
	fmt.Fprintf(w, "  %s -list-packages\n", fs.Name())
	fmt.Fprintf(w, "  %s -list-client\n", fs.Name())
	fmt.Fprintf(w, "  %s -list-cli\n", fs.Name())
	fmt.Fprintf(w, "  %s -spec-dir %s -dry-run -v\n", fs.Name(), defaultSpecDir)
	fmt.Fprintf(w, "  %s -spec %s%srbac.yaml -only list_roles,get_role\n", fs.Name(), defaultSpecDir, string(os.PathSeparator))
	fmt.Fprintf(w, "  %s -list-sections\n", fs.Name())
}

func loadConfiguredSpecs(cfg *Config) ([]*spec.SpecFile, error) {
	var (
		paths []string
		err   error
	)
	if cfg.SpecPath != "" {
		paths = []string{cfg.SpecPath}
	} else {
		paths, err = DiscoverYamlFiles(cfg.SpecDir)
	}
	if err != nil {
		return nil, err
	}

	specs, err := LoadSpecs(paths)
	if err != nil {
		return nil, fmt.Errorf("failed to load scaffold specs: %w", err)
	}
	return specs, nil
}

func onlySet(items []string) map[string]struct{} {
	if len(items) == 0 {
		return nil
	}

	out := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.ToLower(strings.TrimSpace(item))
		if item != "" {
			out[item] = struct{}{}
		}
	}
	return out
}

func includesOnly(only map[string]struct{}, ids ...string) bool {
	if len(only) == 0 {
		return true
	}
	for _, id := range ids {
		if _, ok := only[strings.ToLower(strings.TrimSpace(id))]; ok {
			return true
		}
	}
	return false
}

func sortedClientMethods(methods []spec.ClientMethodSpec, only map[string]struct{}) []spec.ClientMethodSpec {
	out := make([]spec.ClientMethodSpec, 0, len(methods))
	for i := range methods {
		method := methods[i]
		if includesOnly(only, method.ID, method.Method) {
			out = append(out, method)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i].Method) < strings.ToLower(out[j].Method)
	})
	return out
}

func sortedCLICommands(commands []spec.CLICommandSpec, only map[string]struct{}) []spec.CLICommandSpec {
	out := make([]spec.CLICommandSpec, 0, len(commands))
	for i := range commands {
		command := commands[i]
		if includesOnly(only, command.ID, command.ClientMethod, command.Use) {
			out = append(out, command)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i].Use) < strings.ToLower(out[j].Use)
	})
	return out
}

func printPackages(specs []*spec.SpecFile) {
	for _, specFile := range specs {
		fmt.Printf("%s\tpackage=%s\tservice=%s\tclient_methods=%d\tcli_commands=%d\n",
			specFile.Path,
			specFile.Package.ID,
			specFile.Package.Service,
			len(specFile.ClientMethods),
			len(specFile.CLICommands),
		)
	}
}

func printClientMethods(specs []*spec.SpecFile, only map[string]struct{}) {
	for _, specFile := range specs {
		methods := sortedClientMethods(specFile.ClientMethods, only)
		for i := range methods {
			method := &methods[i]
			section := method.Section
			if section == "" {
				section = specFile.Client.DefaultSection
			}
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", specFile.Package.ID, section, method.HTTPMethod, method.Method, method.ID)
		}
	}
}

func printCLICommands(specs []*spec.SpecFile, only map[string]struct{}) {
	for _, specFile := range specs {
		commands := sortedCLICommands(specFile.CLICommands, only)
		for i := range commands {
			command := &commands[i]
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", specFile.Package.ID, command.GroupID, command.Use, command.ClientMethod, command.ID)
		}
	}
}

func printGenerationPlan(specs []*spec.SpecFile, only map[string]struct{}, skipExisting bool) error {
	existingByTarget, err := existingMethodsByTarget(specs)
	if err != nil {
		return err
	}

	for _, specFile := range specs {
		methods := sortedClientMethods(specFile.ClientMethods, only)
		for i := range methods {
			method := &methods[i]
			section := method.Section
			if section == "" {
				section = specFile.Client.DefaultSection
			}

			status := "generate"
			location := ""
			if existing, exists := existingByTarget[specFile.Client.TargetFile][method.Method]; exists {
				location = fmt.Sprintf("%s:%d", specFile.Client.TargetFile, existing.Line)
				switch {
				case skipExisting:
					status = "skip-existing"
				case existing.Managed:
					status = "update-managed"
				default:
					status = "skip-human"
				}
			}

			fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				specFile.Package.ID,
				status,
				section,
				method.HTTPMethod,
				method.Method,
				method.ID,
				location,
			)
		}

		commands := sortedCLICommands(specFile.CLICommands, only)
		if len(commands) == 0 {
			continue
		}

		status := "generate-cli-file"
		location := specFile.CLI.TargetFile
		if _, err := os.Stat(specFile.CLI.TargetFile); err == nil {
			managed, managedErr := isManagedScaffoldFile(specFile.CLI.TargetFile)
			if managedErr != nil {
				return managedErr
			}
			if managed {
				status = "update-cli-file"
			} else {
				status = "skip-human-cli-file"
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("stat CLI target %s: %w", specFile.CLI.TargetFile, err)
		}
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			specFile.Package.ID,
			status,
			"cli",
			specFile.CLI.Use,
			specFile.CLI.CommandFactory,
			"cli_file",
			location,
		)

		if specFile.CLI.ParentFile != "" {
			fmt.Printf("%s\tgenerate-parent-registration\tcli\t%s\t%s\t%s\t%s\n",
				specFile.Package.ID,
				specFile.CLI.ParentFactory,
				specFile.CLI.CommandFactory,
				specFile.CLI.GroupID,
				specFile.CLI.ParentFile,
			)
		}

		for i := range commands {
			command := &commands[i]
			fmt.Printf("%s\tgenerate-cli-command\t%s\t%s\t%s\t%s\t%s\n",
				specFile.Package.ID,
				command.GroupID,
				command.Use,
				command.Factory,
				command.ClientMethod,
				command.ID,
			)
		}
	}
	return nil
}

func printExistingMethods(specs []*spec.SpecFile) error {
	for _, targetFile := range uniqueTargetFiles(specs) {
		methods, err := findExistingClientMethods(targetFile)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", targetFile)
		names := make([]string, 0, len(methods))
		for name := range methods {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			owner := "human"
			if methods[name].Managed {
				owner = "managed"
			}
			fmt.Printf("  %s:%d\t%s\t%s\n", targetFile, methods[name].Line, owner, name)
		}
	}
	return nil
}

func printCheckSummary(specs []*spec.SpecFile, only map[string]struct{}, skipExisting bool) error {
	existingByTarget, err := existingMethodsByTarget(specs)
	if err != nil {
		return err
	}

	var generateCount, updateCount, humanCount, skippedCount int
	for _, specFile := range specs {
		methods := sortedClientMethods(specFile.ClientMethods, only)
		for i := range methods {
			method := methods[i]
			existing, exists := existingByTarget[specFile.Client.TargetFile][method.Method]
			switch {
			case !exists:
				generateCount++
			case skipExisting:
				skippedCount++
			case existing.Managed:
				updateCount++
			default:
				humanCount++
			}
		}
	}

	var cliGenerateCount, cliUpdateCount, cliHumanCount int
	for _, specFile := range specs {
		commands := sortedCLICommands(specFile.CLICommands, only)
		if len(commands) == 0 || strings.TrimSpace(specFile.CLI.TargetFile) == "" {
			continue
		}

		managed, managedErr := isManagedScaffoldFile(specFile.CLI.TargetFile)
		if managedErr != nil {
			return managedErr
		}
		if _, err := os.Stat(specFile.CLI.TargetFile); true {
			switch {
			case os.IsNotExist(err):
				cliGenerateCount++
			case err != nil:
				return fmt.Errorf("stat CLI target %s: %w", specFile.CLI.TargetFile, err)
			case managed:
				cliUpdateCount++
			default:
				cliHumanCount++
			}
		}
	}

	fmt.Printf("client_methods\tgenerate=%d\tupdate-managed=%d\tskip-human=%d\tskip-existing=%d\n", generateCount, updateCount, humanCount, skippedCount)
	fmt.Printf("cli_files\tgenerate=%d\tupdate-managed=%d\tskip-human=%d\n", cliGenerateCount, cliUpdateCount, cliHumanCount)
	return nil
}

func checkSpecs(specs []*spec.SpecFile, only map[string]struct{}, skipExisting bool) error {
	_ = skipExisting
	diagnostics := &Diagnostics{}
	for _, specFile := range specs {
		methods := sortedClientMethods(specFile.ClientMethods, only)
		for i := range methods {
			method := &methods[i]
			if method.Request != nil && method.Request.Type != "" {
				typeName := baseTypeName(method.Request.Type)
				if _, err := rgjson.FindStructDef(typeName); err != nil {
					diagnostics.Add(specFile.Path, "client_methods."+method.ID+".request.type", err.Error())
				}
			}
		}
	}
	if diagnostics.HasErrors() {
		return diagnostics
	}
	return nil
}

func buildClientFiles(
	specs []*spec.SpecFile,
	only map[string]struct{},
	templatesDir string,
	skipExisting bool,
) (map[string]*create.FileBuild, error) {
	existingByTarget, err := existingMethodsByTarget(specs)
	if err != nil {
		return nil, err
	}

	builds := map[string]*create.FileBuild{}
	for _, specFile := range specs {
		targetFile := specFile.Client.TargetFile
		if targetFile == "" {
			return nil, fmt.Errorf("spec %s: client.target_file is empty", specFile.Path)
		}

		build, ok := builds[targetFile]
		if !ok {
			build = create.NewFileBuild(targetFile)
			builds[targetFile] = build
		}

		methods := sortedClientMethods(specFile.ClientMethods, only)
		for i := range methods {
			method := &methods[i]
			existing, exists := existingByTarget[targetFile][method.Method]
			if exists {
				if skipExisting {
					continue
				}
				if !existing.Managed {
					continue
				}
			}

			if method.Request != nil && method.Request.Type != "" {
				typeName := baseTypeName(method.Request.Type)
				if _, err := rgjson.FindStructDef(typeName); err != nil {
					return nil, fmt.Errorf("spec %s method %q: %w", specFile.Path, method.ID, err)
				}
			}

			rendered, err := create.CreateClientMethod(
				method,
				filepath.Join(templatesDir, "client_method.go.tmpl"),
			)
			if err != nil {
				return nil, fmt.Errorf("spec %s method %q: %w", specFile.Path, method.ID, err)
			}

			section := method.Section
			if section == "" {
				section = specFile.Client.DefaultSection
			}
			if section == "" {
				return nil, fmt.Errorf("spec %s method %q: no section set and no default_section configured", specFile.Path, method.ID)
			}

			if exists && existing.Managed {
				build.Replace(method.Method, rendered)
				continue
			}

			build.AppendSorted(section, method.Method, rendered)
		}
	}

	return builds, nil
}

func buildCLIFiles(
	specs []*spec.SpecFile,
	only map[string]struct{},
	templatesDir string,
) (map[string]string, error) {
	files := map[string]string{}
	for _, specFile := range specs {
		commands := sortedCLICommands(specFile.CLICommands, only)
		if len(commands) == 0 {
			continue
		}

		targetFile := strings.TrimSpace(specFile.CLI.TargetFile)
		if targetFile == "" {
			return nil, fmt.Errorf("spec %s: cli.target_file is empty", specFile.Path)
		}

		rendered, err := create.CreateCLIFile(
			specFile,
			commands,
			filepath.Join(templatesDir, "cli_file.go.tmpl"),
		)
		if err != nil {
			return nil, fmt.Errorf("spec %s cli: %w", specFile.Path, err)
		}

		if _, exists := files[targetFile]; exists {
			return nil, fmt.Errorf("multiple specs generate CLI target %s", targetFile)
		}
		files[targetFile] = rendered
	}
	return files, nil
}

func buildCLIParentFiles(
	specs []*spec.SpecFile,
	only map[string]struct{},
) (map[string]string, error) {
	registrationsByParentFile := map[string][]create.CLIParentRegistration{}
	for _, specFile := range specs {
		commands := sortedCLICommands(specFile.CLICommands, only)
		if len(commands) == 0 {
			continue
		}

		parentFile := strings.TrimSpace(specFile.CLI.ParentFile)
		if parentFile == "" {
			continue
		}
		if strings.TrimSpace(specFile.CLI.ParentFactory) == "" {
			return nil, fmt.Errorf("spec %s: cli.parent_factory is required when cli.parent_file is set", specFile.Path)
		}
		if strings.TrimSpace(specFile.CLI.CommandFactory) == "" {
			return nil, fmt.Errorf("spec %s: cli.command_factory is required when cli.parent_file is set", specFile.Path)
		}

		registrationsByParentFile[parentFile] = append(registrationsByParentFile[parentFile], create.CLIParentRegistration{
			ParentFactory:  specFile.CLI.ParentFactory,
			CommandFactory: specFile.CLI.CommandFactory,
			GroupID:        specFile.CLI.GroupID,
		})
	}

	files := map[string]string{}
	for parentFile, registrations := range registrationsByParentFile {
		generated, err := create.BuildCLIParentFile(parentFile, registrations)
		if err != nil {
			return nil, err
		}
		files[parentFile] = generated
	}
	return files, nil
}

func existingMethodsByTarget(specs []*spec.SpecFile) (map[string]map[string]ExistingClientMethod, error) {
	out := map[string]map[string]ExistingClientMethod{}
	for _, targetFile := range uniqueTargetFiles(specs) {
		existing, err := findExistingClientMethods(targetFile)
		if err != nil {
			return nil, err
		}
		out[targetFile] = existing
	}
	return out, nil
}

func buildTargetFiles(builds map[string]*create.FileBuild) []string {
	files := make([]string, 0, len(builds))
	for targetFile := range builds {
		files = append(files, targetFile)
	}
	sort.Strings(files)
	return files
}

func writeGeneratedFile(path, content string) error {
	mode := os.FileMode(0o600)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	return os.WriteFile(path, []byte(content), mode) // #nosec G306
}

func writeGeneratedNewFile(path, content string, force bool) error {
	if existing, err := os.ReadFile(path); err == nil { // #nosec G304
		if string(existing) == content {
			return nil
		}
		managed, managedErr := isManagedScaffoldFile(path)
		if managedErr != nil {
			return managedErr
		}
		if !managed && !force {
			return fmt.Errorf("refusing to overwrite human-owned CLI scaffold %s; restore the scaffold marker or pass -force", path)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read %s: %w", path, err)
	}

	return writeGeneratedFile(path, content)
}

func formatGoFiles(files []string) error {
	var goFiles []string
	for _, file := range files {
		if filepath.Ext(file) == ".go" {
			goFiles = append(goFiles, file)
		}
	}
	if len(goFiles) == 0 {
		return nil
	}

	args := append([]string{"-w"}, goFiles...)
	cmd := exec.CommandContext(context.Background(), "gofmt", args...) // #nosec G204
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("gofmt %s: %w: %s", strings.Join(goFiles, ", "), err, strings.TrimSpace(string(output)))
	}
	for _, file := range goFiles {
		fmt.Printf("formatted %s\n", file)
	}
	return nil
}

func formatGeneratedSource(path, content string) (string, error) {
	if filepath.Ext(path) != ".go" {
		return content, nil
	}

	formatted, err := format.Source([]byte(content))
	if err != nil {
		return "", fmt.Errorf("format generated source %s: %w", path, err)
	}
	return string(formatted), nil
}

func uniqueTargetFiles(specs []*spec.SpecFile) []string {
	seen := map[string]struct{}{}
	for _, specFile := range specs {
		if specFile.Client.TargetFile == "" {
			continue
		}
		seen[specFile.Client.TargetFile] = struct{}{}
	}

	files := make([]string, 0, len(seen))
	for file := range seen {
		files = append(files, file)
	}
	sort.Strings(files)
	return files
}

func sortedMapKeys[V any](items map[string]V) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func uniqueStrings(items []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if _, exists := seen[item]; exists {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}

func suggestFlagName(parseErr string, fs *flag.FlagSet) string {
	const prefix = "flag provided but not defined: -"
	_, after, ok := strings.Cut(parseErr, prefix)
	if !ok {
		return ""
	}

	unknown := strings.TrimSpace(after)
	unknown = strings.TrimLeft(unknown, "-")
	if unknown == "" {
		return ""
	}

	bestName := ""
	bestDistance := 1 << 30
	fs.VisitAll(func(f *flag.Flag) {
		distance := levenshtein(unknown, f.Name)
		if distance < bestDistance {
			bestDistance = distance
			bestName = f.Name
		}
	})

	threshold := 2
	if len(unknown) > 8 {
		threshold = 3
	}
	if bestDistance > threshold {
		return ""
	}
	return bestName
}

func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	if a == "" {
		return len(b)
	}
	if b == "" {
		return len(a)
	}

	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			curr[j] = min3(
				prev[j]+1,
				curr[j-1]+1,
				prev[j-1]+cost,
			)
		}
		prev, curr = curr, prev
	}

	return prev[len(b)]
}

func min3(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= c {
		return b
	}
	return c
}

func RunFmt() error {
	cmd := exec.CommandContext(context.Background(), "mise", "run", "fmt")
	_, err := cmd.CombinedOutput()
	return err
}

func baseTypeName(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "*")
	s = strings.TrimPrefix(s, "models.")
	return s
}
