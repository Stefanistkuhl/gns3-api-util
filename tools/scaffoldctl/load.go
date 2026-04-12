package main

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/0xveya/gns3util/tools/scaffoldctl/spec"
	"gopkg.in/yaml.v3"
)

func LoadSpecs(paths []string) ([]*spec.SpecFile, error) {
	sort.Strings(paths)

	specs := make([]*spec.SpecFile, 0, len(paths))
	diagnostics := &Diagnostics{}

	for _, path := range paths {
		specFile, err := loadSpec(path)
		if err != nil {
			var typeErr *yaml.TypeError
			if errors.As(err, &typeErr) {
				for _, msg := range typeErr.Errors {
					diagnostics.Add(path, "", msg)
				}
				continue
			}
			diagnostics.Add(path, "", err.Error())
			continue
		}

		validateSpec(specFile, diagnostics)
		specs = append(specs, specFile)
	}

	if diagnostics.HasErrors() {
		return nil, diagnostics
	}

	return specs, nil
}

func loadSpec(path string) (*spec.SpecFile, error) {
	data, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var specFile spec.SpecFile
	dec := yaml.NewDecoder(strings.NewReader(string(data)))
	dec.KnownFields(true)
	if err := dec.Decode(&specFile); err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %w", err)
	}

	specFile.Path = path
	return &specFile, nil
}

func validateSpec(specFile *spec.SpecFile, diagnostics *Diagnostics) {
	if specFile.Version != 1 {
		diagnostics.Add(specFile.Path, "version", "must be 1")
	}
	if strings.TrimSpace(specFile.Package.ID) == "" {
		diagnostics.Add(specFile.Path, "package.id", "is required")
	}
	if strings.TrimSpace(specFile.Package.Service) == "" {
		diagnostics.Add(specFile.Path, "package.service", "is required")
	}
	if strings.TrimSpace(specFile.Client.TargetFile) == "" {
		diagnostics.Add(specFile.Path, "client.target_file", "is required")
	}
	if strings.TrimSpace(specFile.Client.DefaultSection) == "" {
		diagnostics.Add(specFile.Path, "client.default_section", "is required")
	}
	if !isKnownSection(specFile.Client.DefaultSection) {
		diagnostics.Add(specFile.Path, "client.default_section", fmt.Sprintf("unknown section %q", specFile.Client.DefaultSection))
	}
	if len(specFile.ClientMethods) == 0 {
		diagnostics.Add(specFile.Path, "client_methods", "must contain at least one method")
	}

	seen := map[string]struct{}{}
	for i := range specFile.ClientMethods {
		method := &specFile.ClientMethods[i]
		prefix := fmt.Sprintf("client_methods[%d]", i)
		if strings.TrimSpace(method.ID) == "" {
			diagnostics.Add(specFile.Path, prefix+".id", "is required")
		} else {
			if _, exists := seen[method.ID]; exists {
				diagnostics.Add(specFile.Path, prefix+".id", fmt.Sprintf("duplicate method id %q", method.ID))
			}
			seen[method.ID] = struct{}{}
		}

		if strings.TrimSpace(method.Method) == "" {
			diagnostics.Add(specFile.Path, prefix+".method", "is required")
		}
		if strings.TrimSpace(method.HTTPMethod) == "" {
			diagnostics.Add(specFile.Path, prefix+".http_method", "is required")
		}
		if strings.TrimSpace(method.URLExpr) == "" {
			diagnostics.Add(specFile.Path, prefix+".url_expr", "is required")
		}

		section := method.Section
		if section == "" {
			section = specFile.Client.DefaultSection
		}
		if !isKnownSection(section) {
			diagnostics.Add(specFile.Path, prefix+".section", fmt.Sprintf("unknown section %q", section))
		}

		if method.Request != nil && method.Request.Body {
			if strings.TrimSpace(method.Request.ArgName) == "" {
				diagnostics.Add(specFile.Path, prefix+".request.arg_name", "is required when request.body=true")
			}
			if strings.TrimSpace(method.Request.Type) == "" {
				diagnostics.Add(specFile.Path, prefix+".request.type", "is required when request.body=true")
			}
			for j, arg := range method.Args {
				if arg.Name == method.Request.ArgName {
					diagnostics.Add(specFile.Path, fmt.Sprintf("%s.args[%d].name", prefix, j), fmt.Sprintf("duplicates request.arg_name %q; body args are generated from request", method.Request.ArgName))
				}
			}
		}

		if !method.Response.NoResponse {
			if strings.TrimSpace(method.Response.PointerType) == "" {
				diagnostics.Add(specFile.Path, prefix+".response.pointer_type", "is required")
			}
			if strings.TrimSpace(method.Response.ValueType) == "" {
				diagnostics.Add(specFile.Path, prefix+".response.value_type", "is required")
			}
			if strings.TrimSpace(method.Response.DecodeVar) == "" {
				diagnostics.Add(specFile.Path, prefix+".response.decode_var", "is required")
			}
			if strings.TrimSpace(method.Response.DecodeError) == "" {
				diagnostics.Add(specFile.Path, prefix+".response.decode_error", "is required")
			}
		}
	}

	cliGroups := map[string]struct{}{}
	if len(specFile.CLICommands) > 0 {
		if strings.TrimSpace(specFile.CLI.TargetFile) == "" {
			diagnostics.Add(specFile.Path, "cli.target_file", "is required when cli_commands are configured")
		}
		if strings.TrimSpace(specFile.CLI.CommandFactory) == "" {
			diagnostics.Add(specFile.Path, "cli.command_factory", "is required when cli_commands are configured")
		}
		if strings.TrimSpace(specFile.CLI.Use) == "" {
			diagnostics.Add(specFile.Path, "cli.use", "is required when cli_commands are configured")
		}
		if strings.TrimSpace(specFile.CLI.Short) == "" {
			diagnostics.Add(specFile.Path, "cli.short", "is required when cli_commands are configured")
		}
		if specFile.CLI.AuthMode != "" && !isKnownAuthMode(specFile.CLI.AuthMode) {
			diagnostics.Add(specFile.Path, "cli.auth_mode", fmt.Sprintf("unknown auth mode %q", specFile.CLI.AuthMode))
		}
		for i, group := range specFile.CLI.Groups {
			prefix := fmt.Sprintf("cli.groups[%d]", i)
			if strings.TrimSpace(group.ID) == "" {
				diagnostics.Add(specFile.Path, prefix+".id", "is required")
			} else {
				if _, exists := cliGroups[group.ID]; exists {
					diagnostics.Add(specFile.Path, prefix+".id", fmt.Sprintf("duplicate group id %q", group.ID))
				}
				cliGroups[group.ID] = struct{}{}
			}
			if strings.TrimSpace(group.Title) == "" {
				diagnostics.Add(specFile.Path, prefix+".title", "is required")
			}
		}

		cliRows := map[string]struct{}{}
		for i, row := range specFile.CLI.Rows {
			prefix := fmt.Sprintf("cli.rows[%d]", i)
			if strings.TrimSpace(row.Type) == "" {
				diagnostics.Add(specFile.Path, prefix+".type", "is required")
			} else {
				if _, exists := cliRows[row.Type]; exists {
					diagnostics.Add(specFile.Path, prefix+".type", fmt.Sprintf("duplicate row type %q", row.Type))
				}
				cliRows[row.Type] = struct{}{}
			}
			if len(row.Fields) == 0 {
				diagnostics.Add(specFile.Path, prefix+".fields", "must contain at least one field")
			}
			seenFields := map[string]struct{}{}
			for j, field := range row.Fields {
				fieldPrefix := fmt.Sprintf("%s.fields[%d]", prefix, j)
				if strings.TrimSpace(field.Name) == "" {
					diagnostics.Add(specFile.Path, fieldPrefix+".name", "is required")
				} else {
					if _, exists := seenFields[field.Name]; exists {
						diagnostics.Add(specFile.Path, fieldPrefix+".name", fmt.Sprintf("duplicate field name %q", field.Name))
					}
					seenFields[field.Name] = struct{}{}
				}
			}
		}
	}

	seenCLI := map[string]struct{}{}
	clientMethods := map[string]struct{}{}
	for i := range specFile.ClientMethods {
		clientMethods[specFile.ClientMethods[i].Method] = struct{}{}
	}
	for i := range specFile.CLICommands {
		command := &specFile.CLICommands[i]
		prefix := fmt.Sprintf("cli_commands[%d]", i)
		if strings.TrimSpace(command.ID) == "" {
			diagnostics.Add(specFile.Path, prefix+".id", "is required")
		} else {
			if _, exists := seenCLI[command.ID]; exists {
				diagnostics.Add(specFile.Path, prefix+".id", fmt.Sprintf("duplicate command id %q", command.ID))
			}
			seenCLI[command.ID] = struct{}{}
		}
		if strings.TrimSpace(command.Use) == "" {
			diagnostics.Add(specFile.Path, prefix+".use", "is required")
		}
		if strings.TrimSpace(command.Short) == "" {
			diagnostics.Add(specFile.Path, prefix+".short", "is required")
		}
		if command.AuthMode != "" && !isKnownAuthMode(command.AuthMode) {
			diagnostics.Add(specFile.Path, prefix+".auth_mode", fmt.Sprintf("unknown auth mode %q", command.AuthMode))
		}
		if command.GroupID != "" && len(cliGroups) > 0 {
			if _, exists := cliGroups[command.GroupID]; !exists {
				diagnostics.Add(specFile.Path, prefix+".group_id", fmt.Sprintf("unknown CLI group %q", command.GroupID))
			}
		}
		if command.ClientMethod != "" {
			if _, exists := clientMethods[command.ClientMethod]; !exists {
				diagnostics.Add(specFile.Path, prefix+".client_method", fmt.Sprintf("unknown client method %q", command.ClientMethod))
			}
		}
		for j, flag := range command.Flags {
			flagPrefix := fmt.Sprintf("%s.flags[%d]", prefix, j)
			if strings.TrimSpace(flag.Name) == "" {
				diagnostics.Add(specFile.Path, flagPrefix+".name", "is required")
			}
			if strings.TrimSpace(flag.Type) == "" {
				diagnostics.Add(specFile.Path, flagPrefix+".type", "is required")
			}
			if strings.TrimSpace(flag.Usage) == "" {
				diagnostics.Add(specFile.Path, flagPrefix+".usage", "is required")
			}
		}
	}
}

func isKnownSection(section string) bool {
	return slices.Contains(clientSections, section)
}

func isKnownAuthMode(authMode string) bool {
	return slices.Contains([]string{"cluster-only", "flexible", "none"}, authMode)
}
