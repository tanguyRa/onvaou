package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
)

// generateRootTree generates a simple file hierarchy for the root README
func generateRootTree(root *Node, rootFolderName string) string {
	var treeOutput bytes.Buffer
	treeOutput.WriteString("```tree\n")
	fmt.Fprintf(&treeOutput, "%s/\n", rootFolderName)

	// Add root level files first (files only, no types/functions)
	fileNames := make([]string, 0, len(root.Files))
	for fileName := range root.Files {
		fileNames = append(fileNames, fileName)
	}
	sort.Strings(fileNames)

	for i, fileName := range fileNames {
		isLastFile := i == len(fileNames)-1 && len(root.Children) == 0
		fmt.Fprintf(&treeOutput, "%s%s\n", getNodePrefix(isLastFile), fileName)
	}

	// Then print subdirectories (structure only)
	keys := make([]string, 0, len(root.Children))
	for k := range root.Children {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, key := range keys {
		printTreeStructureOnly(&treeOutput, root.Children[key], "", i == len(keys)-1)
	}
	treeOutput.WriteString("```")
	return treeOutput.String()
}

// generatePackageTree generates a detailed tree for a specific package
func generatePackageTree(packageNode *Node, packagePath string) string {
	var treeOutput bytes.Buffer
	treeOutput.WriteString("```tree\n")

	// Package name is the last part of the path
	packageName := filepath.Base(packagePath)
	fmt.Fprintf(&treeOutput, "%s/\n", packageName)

	// Add files with their types and functions
	fileNames := make([]string, 0, len(packageNode.Files))
	for fileName := range packageNode.Files {
		fileNames = append(fileNames, fileName)
	}
	sort.Strings(fileNames)

	// Calculate if we have subdirectories
	hasChildren := len(packageNode.Children) > 0

	for i, fileName := range fileNames {
		isLastFile := i == len(fileNames)-1 && !hasChildren
		file := packageNode.Files[fileName]
		fmt.Fprintf(&treeOutput, "%s%s\n", getNodePrefix(isLastFile), fileName)

		// File content prefix
		contentPrefix := ""
		if !isLastFile || hasChildren {
			contentPrefix = "│   "
		} else {
			contentPrefix = "    "
		}

		// Print types and functions for this file
		for j, t := range file.Types {
			isLastType := j == len(file.Types)-1 && len(file.Funcs) == 0

			var typeStr string
			// Check if it's an underlying type definition
			if len(t.Properties) == 1 && t.Properties[0].Name == "__underlying" {
				typeStr = fmt.Sprintf("type %s %s", t.Name, t.Properties[0].Type)
			} else {
				typeStr = fmt.Sprintf("type %s {", t.Name)
				for k, prop := range t.Properties {
					if k > 0 {
						typeStr += ", "
					}
					if prop.Type == "embedded" {
						typeStr += prop.Name // Already formatted with * prefix
					} else {
						typeStr += fmt.Sprintf("%s: %s", prop.Name, prop.Type)
					}
				}
				typeStr += "}"
			}

			fmt.Fprintf(&treeOutput, "%s%s%s\n", contentPrefix, getNodePrefix(isLastType), typeStr)
		}

		for j, f := range file.Funcs {
			isLastFunc := j == len(file.Funcs)-1
			fmt.Fprintf(&treeOutput, "%s%s%s\n", contentPrefix, getNodePrefix(isLastFunc), f)
		}
	}

	// Print subdirectories if any
	childKeys := make([]string, 0, len(packageNode.Children))
	for k := range packageNode.Children {
		childKeys = append(childKeys, k)
	}
	sort.Strings(childKeys)

	for i, key := range childKeys {
		isLastChild := i == len(childKeys)-1
		printTree(&treeOutput, packageNode.Children[key], "", isLastChild)
	}

	treeOutput.WriteString("```")
	return treeOutput.String()
}

// printTreeStructureOnly prints directory structure without types and functions
func printTreeStructureOnly(w *bytes.Buffer, node *Node, prefix string, isLast bool) {
	// Print current node
	fmt.Fprintf(w, "%s%s%s/\n", prefix, getNodePrefix(isLast), node.Path)

	newPrefix := prefix
	if isLast {
		newPrefix += "    "
	} else {
		newPrefix += "│   "
	}

	// Print files first (names only, no details)
	fileNames := make([]string, 0, len(node.Files))
	for fileName := range node.Files {
		fileNames = append(fileNames, fileName)
	}
	sort.Strings(fileNames)

	// Calculate if we have more items after files
	hasChildren := len(node.Children) > 0

	// Print files
	for i, fileName := range fileNames {
		isLastFile := i == len(fileNames)-1 && !hasChildren
		fmt.Fprintf(w, "%s%s%s\n", newPrefix, getNodePrefix(isLastFile), fileName)
	}

	// Print subdirectories
	childKeys := make([]string, 0, len(node.Children))
	for k := range node.Children {
		childKeys = append(childKeys, k)
	}
	sort.Strings(childKeys)

	for i, key := range childKeys {
		isLastChild := i == len(childKeys)-1
		printTreeStructureOnly(w, node.Children[key], newPrefix, isLastChild)
	}
}

// printTree prints directory structure with full details (types and functions)
func printTree(w *bytes.Buffer, node *Node, prefix string, isLast bool) {
	// Print current node
	fmt.Fprintf(w, "%s%s%s/\n", prefix, getNodePrefix(isLast), node.Path)

	newPrefix := prefix
	if isLast {
		newPrefix += "    "
	} else {
		newPrefix += "│   "
	}

	// Print files first
	fileNames := make([]string, 0, len(node.Files))
	for fileName := range node.Files {
		fileNames = append(fileNames, fileName)
	}
	sort.Strings(fileNames)

	// Calculate if we have more items after files
	hasChildren := len(node.Children) > 0

	// Print files and their contents
	for i, fileName := range fileNames {
		isLastFile := i == len(fileNames)-1 && !hasChildren
		file := node.Files[fileName]
		fmt.Fprintf(w, "%s%s%s\n", newPrefix, getNodePrefix(isLastFile), fileName)

		// File content prefix
		contentPrefix := newPrefix
		if !isLastFile || hasChildren {
			contentPrefix += "│   "
		} else {
			contentPrefix += "    "
		}

		// Print types
		for j, t := range file.Types {
			isLastType := j == len(file.Types)-1 && len(file.Funcs) == 0

			var typeStr string
			// Check if it's an underlying type definition
			if len(t.Properties) == 1 && t.Properties[0].Name == "__underlying" {
				typeStr = fmt.Sprintf("type %s %s", t.Name, t.Properties[0].Type)
			} else {
				typeStr = fmt.Sprintf("type %s {", t.Name)
				for k, prop := range t.Properties {
					if k > 0 {
						typeStr += ", "
					}
					if prop.Type == "embedded" {
						typeStr += prop.Name // Already formatted with * prefix
					} else {
						typeStr += fmt.Sprintf("%s: %s", prop.Name, prop.Type)
					}
				}
				typeStr += "}"
			}

			fmt.Fprintf(w, "%s%s%s\n", contentPrefix, getNodePrefix(isLastType), typeStr)
		}

		// Print functions
		for j, f := range file.Funcs {
			isLastFunc := j == len(file.Funcs)-1
			fmt.Fprintf(w, "%s%s%s\n", contentPrefix, getNodePrefix(isLastFunc), f)
		}
	}

	// Print subdirectories
	childKeys := make([]string, 0, len(node.Children))
	for k := range node.Children {
		childKeys = append(childKeys, k)
	}
	sort.Strings(childKeys)

	for i, key := range childKeys {
		isLastChild := i == len(childKeys)-1
		printTree(w, node.Children[key], newPrefix, isLastChild)
	}
}

// getNodePrefix returns the appropriate tree prefix for a node
func getNodePrefix(isLast bool) string {
	if isLast {
		return "└── "
	}
	return "├── "
}