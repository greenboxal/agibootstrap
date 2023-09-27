package codeanalysis

import (
	"net/url"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type Symbol struct {
	psi.NodeBase

	Name string `json:"name"`
}

func NewSymbol(name string) *Symbol {
	sym := &Symbol{
		Name: name,
	}

	sym.Init(sym)

	return sym
}

func (s *Symbol) String() string      { return "codeanalysis.Symbol(" + s.Name + ")" }
func (s *Symbol) PsiNodeName() string { return s.Name }

func (s *Symbol) GetDeclaration() psi.Node {
	return psi.GetEdgeOrNil[psi.Node](s, DeclarationEdge.Singleton())
}
func (s *Symbol) SetDeclaration(node psi.Node) { s.SetEdge(DeclarationEdge.Singleton(), node) }

func (s *Symbol) GetReferences() []psi.Node {
	return psi.GetEdges(s, ReferenceBackEdge)
}

func (s *Symbol) AddReference(n psi.Node) {
	n.SetEdge(ReferenceEdge.Named(url.PathEscape(n.CanonicalPath().String())), s)
	s.SetEdge(ReferenceBackEdge.Named(url.PathEscape(n.CanonicalPath().String())), n)
}

var SymbolType = psi.DefineNodeType[*Symbol]()
var SymbolEdge = psi.DefineEdgeType[*Symbol]("ca.Symbol")
var DeclarationEdge = psi.DefineEdgeType[psi.Node]("ca.Declaration")
var ReferenceEdge = psi.DefineEdgeType[*Symbol]("ca.Reference")
var ReferenceBackEdge = psi.DefineEdgeType[psi.Node]("ca.BackRef")
