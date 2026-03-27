package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// discoverGoPackages scans the project and finds all directories containing Go files
func discoverGoPackages(projectRoot string) ([]PackageInfo, error) {
	var packages []PackageInfo

	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and files
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip the cmd directory to avoid including the docs tool itself
		if info.IsDir() && info.Name() == "cmd" {
			return filepath.SkipDir
		}

		// Skip non-Go files
		if !info.IsDir() && !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		// If it's a Go file, check if its directory is already recorded
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			dirPath := filepath.Dir(path)
			relPath, err := filepath.Rel(projectRoot, dirPath)
			if err != nil {
				return err
			}

			// Check if this package is already in our list
			found := false
			for i := range packages {
				if packages[i].Path == relPath {
					packages[i].HasGoFiles = true
					found = true
					break
				}
			}

			// If not found, add it
			if !found {
				packages = append(packages, PackageInfo{
					Path:       relPath,
					Name:       filepath.Base(dirPath),
					HasGoFiles: true,
				})
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory tree: %v", err)
	}

	// Filter to only include packages with Go files and sort by path
	var goPackages []PackageInfo
	for _, pkg := range packages {
		if pkg.HasGoFiles && pkg.Path != "." { // Exclude root directory
			goPackages = append(goPackages, pkg)
		}
	}

	sort.Slice(goPackages, func(i, j int) bool {
		return goPackages[i].Path < goPackages[j].Path
	})

	return goPackages, nil
}

// buildFileTree walks the directory and builds a complete file tree with Go parsing
func buildFileTree(projectRoot string) (*Node, error) {
	root := &Node{
		Path:     ".",
		Children: make(map[string]*Node),
		Files:    make(map[string]*FileNode),
	}

	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the cmd directory since it contains our documentation tool
		if info.IsDir() && info.Name() == "cmd" {
			return filepath.SkipDir
		}

		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Include multiple file types for structure
		if !strings.HasSuffix(info.Name(), ".go") &&
			!strings.HasSuffix(info.Name(), ".html") &&
			!strings.HasSuffix(info.Name(), ".css") &&
			!strings.HasSuffix(info.Name(), ".js") &&
			!strings.HasSuffix(info.Name(), ".json") &&
			!strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("error getting absolute path for %s: %v", path, err)
		}

		relPath, err := filepath.Rel(projectRoot, absPath)
		if err != nil {
			return fmt.Errorf("error getting relative path: %v", err)
		}

		current := root
		parts := strings.Split(filepath.Dir(relPath), string(os.PathSeparator))

		for _, part := range parts {
			if part == "." {
				continue
			}
			if _, ok := current.Children[part]; !ok {
				current.Children[part] = &Node{
					Path:     part,
					Children: make(map[string]*Node),
					Files:    make(map[string]*FileNode),
				}
			}
			current = current.Children[part]
		}

		// Create file node
		fileName := filepath.Base(path)
		fileNode := &FileNode{
			Name:  fileName,
			Types: make([]TypeInfo, 0),
			Funcs: make([]string, 0),
		}
		current.Files[fileName] = fileNode

		// Only parse Go files for types and functions
		if strings.HasSuffix(fileName, ".go") {
			if err := parseGoFile(path, fileNode); err != nil {
				fmt.Printf("Warning: Failed to parse %s: %v\n", path, err)
				// Continue processing other files instead of failing completely
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}

	return root, nil
}

// buildSimpleFileTree builds a file tree without Go parsing (for root tree after package processing)
func buildSimpleFileTree(projectRoot string) (*Node, error) {
	root := &Node{
		Path:     ".",
		Children: make(map[string]*Node),
		Files:    make(map[string]*FileNode),
	}

	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the cmd directory since it contains our documentation tool
		if info.IsDir() && info.Name() == "cmd" {
			return filepath.SkipDir
		}

		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Include multiple file types for structure
		if !strings.HasSuffix(info.Name(), ".go") &&
			!strings.HasSuffix(info.Name(), ".html") &&
			!strings.HasSuffix(info.Name(), ".css") &&
			!strings.HasSuffix(info.Name(), ".js") &&
			!strings.HasSuffix(info.Name(), ".json") &&
			!strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("error getting absolute path for %s: %v", path, err)
		}

		relPath, err := filepath.Rel(projectRoot, absPath)
		if err != nil {
			return fmt.Errorf("error getting relative path: %v", err)
		}

		current := root
		parts := strings.Split(filepath.Dir(relPath), string(os.PathSeparator))

		for _, part := range parts {
			if part == "." {
				continue
			}
			if _, ok := current.Children[part]; !ok {
				current.Children[part] = &Node{
					Path:     part,
					Children: make(map[string]*Node),
					Files:    make(map[string]*FileNode),
				}
			}
			current = current.Children[part]
		}

		// Create file node (no Go parsing for simple tree)
		fileName := filepath.Base(path)
		fileNode := &FileNode{
			Name:  fileName,
			Types: make([]TypeInfo, 0),
			Funcs: make([]string, 0),
		}
		current.Files[fileName] = fileNode

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}

	return root, nil
}

