package create

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
)

const scaffoldMarker = "Scaffolded by tools/scaffoldctl"

type managedMethodReplacement struct {
	Start    int
	End      int
	Method   string
	Rendered string
}

func applyManagedMethodReplacements(filename, text string, replacements map[string]string) (string, error) {
	if len(replacements) == 0 {
		return text, nil
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, text, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("parse target file %s: %w", filename, err)
	}

	var items []managedMethodReplacement
	found := map[string]bool{}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || len(fn.Recv.List) != 1 || fn.Name == nil {
			continue
		}
		if !isClientV2Receiver(fn.Recv.List[0].Type) || !hasScaffoldMarker(fn.Doc) {
			continue
		}

		rendered, exists := replacements[fn.Name.Name]
		if !exists {
			continue
		}

		start := fset.Position(fn.Pos()).Offset
		if fn.Doc != nil {
			start = fset.Position(fn.Doc.Pos()).Offset
		}
		end := fset.Position(fn.End()).Offset
		if end < len(text) && text[end] == '\n' {
			end++
		}

		items = append(items, managedMethodReplacement{
			Start:    start,
			End:      end,
			Method:   fn.Name.Name,
			Rendered: strings.TrimRight(rendered, "\n") + "\n",
		})
		found[fn.Name.Name] = true
	}

	for method := range replacements {
		if !found[method] {
			return "", fmt.Errorf("managed replacement target %s not found in %s", method, filename)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Start > items[j].Start
	})

	for _, item := range items {
		text = text[:item.Start] + item.Rendered + text[item.End:]
	}

	return text, nil
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
