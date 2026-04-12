package create

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"sort"
	"strings"
)

type CLIParentRegistration struct {
	ParentFactory  string
	CommandFactory string
	GroupID        string
}

func BuildCLIParentFile(targetFile string, registrations []CLIParentRegistration) (string, error) {
	src, err := os.ReadFile(targetFile) // #nosec G304
	if err != nil {
		return "", fmt.Errorf("read CLI parent file %s: %w", targetFile, err)
	}

	text := strings.ReplaceAll(string(src), "\r\n", "\n")
	pending := missingParentRegistrations(text, registrations)
	if len(pending) == 0 {
		return text, nil
	}

	parentFactory := strings.TrimSpace(pending[0].ParentFactory)
	if parentFactory == "" {
		return "", fmt.Errorf("CLI parent file %s: parent factory is required", targetFile)
	}
	for _, reg := range pending[1:] {
		if strings.TrimSpace(reg.ParentFactory) != parentFactory {
			return "", fmt.Errorf("CLI parent file %s: mixed parent factories %q and %q", targetFile, parentFactory, reg.ParentFactory)
		}
	}

	insertOffset, err := parentReturnOffset(targetFile, text, parentFactory)
	if err != nil {
		return "", err
	}

	block := renderParentRegistrationBlock(pending)
	return text[:insertOffset] + block + text[insertOffset:], nil
}

func missingParentRegistrations(text string, registrations []CLIParentRegistration) []CLIParentRegistration {
	pending := make([]CLIParentRegistration, 0, len(registrations))
	seenFactories := map[string]struct{}{}
	for _, reg := range registrations {
		reg.ParentFactory = strings.TrimSpace(reg.ParentFactory)
		reg.CommandFactory = strings.TrimSpace(reg.CommandFactory)
		reg.GroupID = strings.TrimSpace(reg.GroupID)
		if reg.ParentFactory == "" || reg.CommandFactory == "" {
			continue
		}
		if _, exists := seenFactories[reg.CommandFactory]; exists {
			continue
		}
		seenFactories[reg.CommandFactory] = struct{}{}
		if factoryCallExists(text, reg.CommandFactory) {
			continue
		}
		pending = append(pending, reg)
	}

	sort.SliceStable(pending, func(i, j int) bool {
		return strings.ToLower(pending[i].CommandFactory) < strings.ToLower(pending[j].CommandFactory)
	})
	return pending
}

func factoryCallExists(text, factory string) bool {
	pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(factory) + `\s*\(`)
	return pattern.FindStringIndex(text) != nil
}

func parentReturnOffset(filename, text, parentFactory string) (int, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, text, 0)
	if err != nil {
		return 0, fmt.Errorf("parse CLI parent file %s: %w", filename, err)
	}

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name == nil || fn.Name.Name != parentFactory || fn.Body == nil {
			continue
		}

		for i := len(fn.Body.List) - 1; i >= 0; i-- {
			ret, ok := fn.Body.List[i].(*ast.ReturnStmt)
			if !ok {
				continue
			}
			if len(ret.Results) != 1 {
				continue
			}
			if ident, ok := ret.Results[0].(*ast.Ident); ok && ident.Name == "cmd" {
				offset := fset.Position(ret.Pos()).Offset
				if lineStart := strings.LastIndex(text[:offset], "\n"); lineStart >= 0 {
					return lineStart + 1, nil
				}
				return 0, nil
			}
		}
		return 0, fmt.Errorf("CLI parent factory %s in %s has no `return cmd` statement", parentFactory, filename)
	}

	return 0, fmt.Errorf("CLI parent factory %s not found in %s", parentFactory, filename)
}

func renderParentRegistrationBlock(registrations []CLIParentRegistration) string {
	var b strings.Builder
	b.WriteString("\t// Added by scaffoldctl CLI parent registration.\n")
	for _, reg := range registrations {
		varName := parentCommandVarName(reg.CommandFactory)
		fmt.Fprintf(&b, "\t%s := %s()\n", varName, reg.CommandFactory)
		if reg.GroupID != "" {
			fmt.Fprintf(&b, "\t%s.GroupID = %q\n", varName, reg.GroupID)
		}
		b.WriteString("\n")
	}
	b.WriteString("\tcmd.AddCommand(\n")
	for _, reg := range registrations {
		fmt.Fprintf(&b, "\t\t%s,\n", parentCommandVarName(reg.CommandFactory))
	}
	b.WriteString("\t)\n\n")
	return b.String()
}

func parentCommandVarName(factory string) string {
	factory = strings.TrimPrefix(strings.TrimSpace(factory), "New")
	factory = strings.TrimSuffix(factory, "Command")
	factory = strings.TrimSuffix(factory, "Cmd")
	return lowerCamel(factory) + "Cmd"
}
