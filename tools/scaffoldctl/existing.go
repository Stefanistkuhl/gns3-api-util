package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

const scaffoldMarker = "Scaffolded by tools/scaffoldctl"

type ExistingClientMethod struct {
	Name      string
	Line      int
	StartLine int
	EndLine   int
	Managed   bool
}

func findExistingClientMethods(targetFile string) (map[string]ExistingClientMethod, error) {
	data, err := os.ReadFile(targetFile) // #nosec G304
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]ExistingClientMethod{}, nil
		}
		return nil, fmt.Errorf("read target file %s: %w", targetFile, err)
	}

	return findExistingClientMethodsInText(targetFile, string(data))
}

func findExistingClientMethodsInText(filename, text string) (map[string]ExistingClientMethod, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, text, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse client file %s: %w", filename, err)
	}

	out := map[string]ExistingClientMethod{}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || len(fn.Recv.List) != 1 || fn.Name == nil {
			continue
		}

		if !isClientV2Receiver(fn.Recv.List[0].Type) {
			continue
		}

		start := fset.Position(fn.Pos())
		if fn.Doc != nil {
			start = fset.Position(fn.Doc.Pos())
		}
		end := fset.Position(fn.End())

		out[fn.Name.Name] = ExistingClientMethod{
			Name:      fn.Name.Name,
			Line:      start.Line,
			StartLine: start.Line,
			EndLine:   end.Line,
			Managed:   hasScaffoldMarker(fn.Doc),
		}
	}

	return out, nil
}

func isClientV2Receiver(expr ast.Expr) bool {
	star, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}

	ident, ok := star.X.(*ast.Ident)
	return ok && ident.Name == "ClientV2"
}

func hasScaffoldMarker(group *ast.CommentGroup) bool {
	if group == nil {
		return false
	}

	for _, comment := range group.List {
		if comment == nil {
			continue
		}
		if strings.Contains(comment.Text, scaffoldMarker) {
			return true
		}
	}
	return false
}

func isManagedScaffoldFile(path string) (bool, error) {
	data, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("read %s: %w", path, err)
	}

	return strings.Contains(strings.ReplaceAll(string(data), "\r\n", "\n"), scaffoldMarker), nil
}
