package rgjson

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
)

type StructField struct {
	Name string
	Type string
}

func FindStructFields(structName string) (map[string]StructField, error) {
	location, err := FindStructDef(structName)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, location.Path, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", location.Path, err)
	}

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name == nil || typeSpec.Name.Name != structName {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return nil, fmt.Errorf("%s is not a struct", structName)
			}

			fields := make(map[string]StructField, len(structType.Fields.List))
			for _, field := range structType.Fields.List {
				if len(field.Names) == 0 {
					continue
				}

				fieldType := types.ExprString(field.Type)
				for _, name := range field.Names {
					if name == nil {
						continue
					}
					fields[name.Name] = StructField{
						Name: name.Name,
						Type: fieldType,
					}
				}
			}
			return fields, nil
		}
	}

	return nil, fmt.Errorf("struct definition not found: %s", structName)
}
