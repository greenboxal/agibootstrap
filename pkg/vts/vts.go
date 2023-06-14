package vts

import "github.com/greenboxal/agibootstrap/pkg/psi"

func init() {
	// TODO: Use the TypeSystem in the initialization logic
	ts := TypeSystem{
		Node: psi.NewFileNode("main.go"), // replace "main.go" with the appropriate file path
		// TODO: Use ts in the initialization logic
	}
}

type TypeSystem struct {
	Node psi.Node
}
