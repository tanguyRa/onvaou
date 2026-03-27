package main

// FileNode represents a file with its types and functions
type FileNode struct {
	Name  string
	Types []TypeInfo
	Funcs []string
}

// TypeInfo holds detailed type information
type TypeInfo struct {
	Name       string
	Properties []PropertyInfo
}

// PropertyInfo represents a field or property of a type
type PropertyInfo struct {
	Name string
	Type string
	Tag  string
}

// Node represents a directory node in the file tree
type Node struct {
	Path     string
	Children map[string]*Node
	Files    map[string]*FileNode
}

// PackageInfo represents a Go package directory
type PackageInfo struct {
	Path       string // Relative path from project root
	Name       string // Package directory name
	HasGoFiles bool   // Whether the directory contains .go files
}
