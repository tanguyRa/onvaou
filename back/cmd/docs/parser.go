package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
)

// Collector holds the state during AST traversal (Separation of Concerns)
type Collector struct {
	Fset     *token.FileSet
	FileNode *FileNode
}

func parseGoFile(filePath string, fileNode *FileNode) error {
	fset := token.NewFileSet()
	// Parse the file including comments if needed
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	collector := &Collector{
		Fset:     fset,
		FileNode: fileNode,
	}

	// Use ast.Walk with a custom visitor for cleaner routing
	ast.Walk(collector, file)
	return nil
}

// Visit implements the ast.Visitor interface
func (c *Collector) Visit(n ast.Node) ast.Visitor {
	switch x := n.(type) {
	case *ast.TypeSpec:
		c.handleType(x)
	case *ast.FuncDecl:
		c.handleFunc(x)
	}
	return c
}

func (c *Collector) handleType(x *ast.TypeSpec) {
	typeInfo := TypeInfo{
		Name:       x.Name.Name,
		Properties: []PropertyInfo{},
	}

	switch t := x.Type.(type) {
	case *ast.InterfaceType:
		// Interfaces: type Provider interface { Chat(...) }
		for _, field := range t.Methods.List {
			// By formatting field.Type (the FuncType) and field.Names,
			// we skip the field.Doc and field.Comment nodes.
			methodName := ""
			if len(field.Names) > 0 {
				methodName = field.Names[0].Name
			}

			methodSig := nodeToSingleLineString(c.Fset, field.Type)
			// Clean "func" from "func(ctx...) error"
			cleanSig := strings.TrimPrefix(methodSig, "func")

			typeInfo.Properties = append(typeInfo.Properties, PropertyInfo{
				Name: methodName,
				Type: cleanSig,
			})
		}

	case *ast.StructType:
		// Structs: type User struct { Name string }
		for _, field := range t.Fields.List {
			typeStr := nodeToSingleLineString(c.Fset, field.Type)

			// If no names, it's an embedded struct
			if len(field.Names) == 0 {
				typeInfo.Properties = append(typeInfo.Properties, PropertyInfo{
					Name: "embedded",
					Type: typeStr,
				})
			} else {
				for _, name := range field.Names {
					typeInfo.Properties = append(typeInfo.Properties, PropertyInfo{
						Name: name.Name,
						Type: typeStr,
					})
				}
			}
		}

	default:
		// Basic types: type MyInt int
		typeInfo.Properties = append(typeInfo.Properties, PropertyInfo{
			Name: "__underlying",
			Type: nodeToSingleLineString(c.Fset, x.Type),
		})
	}

	c.FileNode.Types = append(c.FileNode.Types, typeInfo)
}

func (c *Collector) handleFunc(funcNode *ast.FuncDecl) {
	var sig strings.Builder

	sig.WriteString("func ")

	if funcNode.Recv != nil && len(funcNode.Recv.List) > 0 {
		// Wrap the type in parens manually because format.Node
		// on a lone type node won't include them.
		typeName := nodeToSingleLineString(c.Fset, funcNode.Recv.List[0].Type)
		sig.WriteString(fmt.Sprintf("(%s) ", typeName))
	}

	sig.WriteString(funcNode.Name.Name)

	// Format funcNode.Type (the FuncType) and strip the leading "func"
	fullSig := nodeToSingleLineString(c.Fset, funcNode.Type)
	sig.WriteString(strings.TrimPrefix(fullSig, "func"))

	c.FileNode.Funcs = append(c.FileNode.Funcs, sig.String())
}

func nodeToSingleLineString(fset *token.FileSet, node interface{}) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return ""
	}
	// Efficiently collapse all whitespace/newlines into single spaces
	return strings.Join(strings.Fields(buf.String()), " ")
}
