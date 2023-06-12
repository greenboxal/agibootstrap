package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/greenboxal/agibootstrap/pkg/codex"
	"github.com/greenboxal/agibootstrap/pkg/io"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

func main() {
	// Get a list of all Go files in the current directory and subdirectories
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".go" {
			processFile(path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %v: %v\n", ".", err)
	}
}

func processFile(path string) {
	// Read the file
	code, err := io.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading the file %s: %v\n", path, err)
		return
	}

	// Parse the file into an AST
	ast := psi.Parse(path, code)

	if ast.Error() != nil {
		fmt.Printf("Error parsing the file %s: %v\n", path, ast.Error())
		return
	}

	// Process the AST nodes
	updated := codex.ProcessNodes(ast)

	// Convert the AST back to code
	newCode, err := ast.ToCode(updated)
	if err != nil {
		fmt.Printf("Error converting the AST back to code for file %s: %v\n", path, err)
		return
	}

	// Write the new code to a new file
	err = io.WriteFile(path+"_new", newCode)
	if err != nil {
		fmt.Printf("Error writing the new code to a file for %s: %v\n", path, err)
	}
}
