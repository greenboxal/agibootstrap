package vts

func init() {
	// TODO: Write a type system abstraction compatible with psi.Node
	ts := TypeSystem{
		Node: nil, // replace nil with the appropriate value
	}
	// TODO: Use ts in the initialization logic
}

type TypeSystem struct {
	Node psi.Node
}
