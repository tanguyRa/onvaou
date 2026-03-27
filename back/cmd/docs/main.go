package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("Starting enhanced documentation generator...")

	// Step 1: Find the API root directory
	apiRoot, err := findApiRoot()
	if err != nil {
		fmt.Printf("Error finding api root: %v\n", err)
		os.Exit(1)
	}

	if err := os.Chdir(apiRoot); err != nil {
		fmt.Printf("Error changing to api root directory: %v\n", err)
		os.Exit(1)
	}

	absProjectPath, err := filepath.Abs(".")
	if err != nil {
		fmt.Printf("Error getting absolute path: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Discover all Go packages
	packages, err := discoverGoPackages(absProjectPath)
	if err != nil {
		fmt.Printf("Error discovering packages: %v\n", err)
		os.Exit(1)
	}

	for _, pkg := range packages {
		fmt.Printf("  - %s\n", pkg.Path)
	}

	// Step 3: Build complete file tree for project structure
	root, err := buildFileTree(absProjectPath)
	if err != nil {
		fmt.Printf("Error building file tree: %v\n", err)
		os.Exit(1)
	}

	// Step 4: Generate/update package-specific README files
	for _, pkg := range packages {
		packageAbsPath := filepath.Join(absProjectPath, pkg.Path)

		// Ensure the package has a README.md file (with package name as title)
		packageName := filepath.Base(pkg.Path)
		if err := ensurePackageReadme(packageAbsPath, packageName); err != nil {
			fmt.Printf("Warning: Failed to ensure README for %s: %v\n", pkg.Path, err)
			continue
		}

		// Find the package node in our tree
		packageNode := findPackageNode(root, pkg.Path)
		if packageNode == nil {
			fmt.Printf("Warning: Could not find node for package %s\n", pkg.Path)
			continue
		}

		// Generate package-specific tree content
		packageTreeContent := generatePackageTree(packageNode, pkg.Path)

		// Update the package README
		if err := updatePackageReadme(packageAbsPath, packageTreeContent); err != nil {
			fmt.Printf("Warning: Failed to update README for %s: %v\n", pkg.Path, err)
			continue
		}
	}

	// Step 5: Rebuild the tree after creating new README files
	root, err = buildSimpleFileTree(absProjectPath)
	if err != nil {
		fmt.Printf("Error rebuilding file tree: %v\n", err)
		os.Exit(1)
	}

	// Step 6: Generate and update root README.md (structure only, using correct folder name)
	rootFolderName := filepath.Base(absProjectPath)
	rootTreeContent := generateRootTree(root, rootFolderName)

	if err := updateRootReadme(absProjectPath, rootTreeContent); err != nil {
		fmt.Printf("Error updating root README: %v\n", err)
		os.Exit(1)
	}

	// Step 7: Summary
	fmt.Printf("Documentation generation completed successfully!\n")
	fmt.Printf("Updated root README.md and %d package README files.\n", len(packages))
}
