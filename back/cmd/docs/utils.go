package main

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"strings"
)

// exprToString converts an AST expression to its string representation
func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + exprToString(e.X)
	case *ast.ArrayType:
		if e.Len == nil {
			return "[]" + exprToString(e.Elt)
		}
		return "[" + exprToString(e.Len) + "]" + exprToString(e.Elt)
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	case *ast.MapType:
		return "map[" + exprToString(e.Key) + "]" + exprToString(e.Value)
	case *ast.ChanType:
		switch e.Dir {
		case ast.SEND:
			return "chan<- " + exprToString(e.Value)
		case ast.RECV:
			return "<-chan " + exprToString(e.Value)
		default:
			return "chan " + exprToString(e.Value)
		}
	case *ast.FuncType:
		return "func()"
	case *ast.BasicLit:
		return e.Value
	case *ast.Ellipsis:
		return "..." + exprToString(e.Elt)
	case *ast.IndexExpr:
		// Handle generic type parameters
		return exprToString(e.X) + "[" + exprToString(e.Index) + "]"
	case *ast.IndexListExpr:
		// Handle multiple generic type parameters
		base := exprToString(e.X) + "["
		for i, index := range e.Indices {
			if i > 0 {
				base += ", "
			}
			base += exprToString(index)
		}
		return base + "]"
	default:
		return fmt.Sprintf("%T", expr)
	}
}

// findApiRoot attempts to find the API root directory
func findApiRoot() (string, error) {
	// Try to find app directory from current working directory (runs in Docker container)
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting working directory: %v", err)
	}

	// Check if we're already in the app directory (Docker container context)
	if filepath.Base(cwd) == "app" {
		return cwd, nil
	}

	// Check if we're in a subdirectory of app
	if strings.Contains(cwd, "app") {
		// Navigate up until we find the app directory
		for dir := cwd; dir != "/"; dir = filepath.Dir(dir) {
			if filepath.Base(dir) == "app" {
				return dir, nil
			}
		}
	}

	// Check if app is a subdirectory of current directory
	appPath := filepath.Join(cwd, "app")
	if _, err := os.Stat(appPath); err == nil {
		return appPath, nil
	}

	return "", fmt.Errorf("could not find app directory")
}

// ensurePackageReadme creates a README.md file with default content if it doesn't exist
func ensurePackageReadme(packagePath string, packageName string) error {
	readmePath := filepath.Join(packagePath, "README.md")

	// Check if README.md already exists
	if _, err := os.Stat(readmePath); err == nil {
		return nil // File already exists
	}

	// Create default content with package title
	defaultContent := fmt.Sprintf("# %s\n\n```tree\n```\n", packageName)

	// Write the file
	err := os.WriteFile(readmePath, []byte(defaultContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create README.md at %s: %v", readmePath, err)
	}

	fmt.Printf("Created README.md at %s\n", readmePath)
	return nil
}

// updatePackageReadme updates the tree section in a package's README.md file
func updatePackageReadme(packagePath string, treeContent string) error {
	readmePath := filepath.Join(packagePath, "README.md")

	input, err := os.ReadFile(readmePath)
	if err != nil {
		if os.IsNotExist(err) {
			packageName := filepath.Base(packagePath)
			if createErr := ensurePackageReadme(packagePath, packageName); createErr != nil {
				return createErr
			}
			input, err = os.ReadFile(readmePath)
		}
		if err != nil {
			return fmt.Errorf("error reading %s: %v", readmePath, err)
		}
	}

	content := string(input)
	startMarker := "```tree"
	endMarker := "```"

	startIdx := strings.Index(content, startMarker)
	if startIdx == -1 {
		return fmt.Errorf("no tree section found in %s", readmePath)
	}

	afterStart := content[startIdx+len(startMarker):]
	endIdx := strings.Index(afterStart, endMarker)
	if endIdx == -1 {
		return fmt.Errorf("no closing tree marker found in %s", readmePath)
	}

	endIdx = startIdx + len(startMarker) + endIdx + len(endMarker)
	newContent := content[:startIdx] + treeContent + content[endIdx:]

	err = os.WriteFile(readmePath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing to %s: %v", readmePath, err)
	}

	return nil
}

// updateRootReadme updates the tree section in the root README.md file
func updateRootReadme(rootPath string, treeContent string) error {
	readmePath := filepath.Join(rootPath, "README.md")

	input, err := os.ReadFile(readmePath)
	if err != nil {
		if os.IsNotExist(err) {
			defaultContent := "```tree\n```\n"
			if writeErr := os.WriteFile(readmePath, []byte(defaultContent), 0644); writeErr != nil {
				return fmt.Errorf("failed to create README.md at %s: %v", readmePath, writeErr)
			}
			input = []byte(defaultContent)
		} else {
			return fmt.Errorf("error reading %s: %v", readmePath, err)
		}
	}

	content := string(input)
	startMarker := "```tree"
	endMarker := "```"

	startIdx := strings.Index(content, startMarker)
	if startIdx == -1 {
		return fmt.Errorf("no tree section found in %s", readmePath)
	}

	afterStart := content[startIdx+len(startMarker):]
	endIdx := strings.Index(afterStart, endMarker)
	if endIdx == -1 {
		return fmt.Errorf("no closing tree marker found in %s", readmePath)
	}

	endIdx = startIdx + len(startMarker) + endIdx + len(endMarker)
	newContent := content[:startIdx] + treeContent + content[endIdx:]

	err = os.WriteFile(readmePath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing to %s: %v", readmePath, err)
	}

	return nil
}

// findPackageNode traverses the tree to find a specific package node
func findPackageNode(root *Node, packagePath string) *Node {
	packageNode := root
	parts := strings.Split(packagePath, string(os.PathSeparator))

	for _, part := range parts {
		if part == "." || part == "" {
			continue
		}
		if child, exists := packageNode.Children[part]; exists {
			packageNode = child
		} else {
			return nil
		}
	}

	return packageNode
}
