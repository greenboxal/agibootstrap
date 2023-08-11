package typing

import (
	"github.com/greenboxal/agibootstrap/pkg/typesystem"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Element struct {
	psi.NodeBase

	Name string `json:"name"`
}

func (e *Element) PsiNodeName() string { return e.Name }

type Type struct {
	Element `ipld:",inline"`

	typ typesystem.Type
}

var TypeType = psi.DefineNodeType[*Type]()

func NewType(name string) *Type {
	t := &Type{}
	t.Init(t)

	t.Name = name

	return t
}

type Interface struct {
	Type `ipld:",inline"`
}

var InterfaceType = psi.DefineNodeType[*Interface]()

func NewInterface(name string) *Interface {
	iface := &Interface{}
	iface.Init(iface)

	iface.Name = name

	return iface
}

type Method struct {
	Element `ipld:",inline"`
}

var MethodType = psi.DefineNodeType[*Method]()

func NewMethod(name string, req *Type, res *Type) *Interface {
	iface := &Interface{}
	iface.Init(iface)

	iface.Name = name

	return iface
}

type Package struct {
	Element `ipld:",inline"`
}

var PackageType = psi.DefineNodeType[*Package]()

func NewPackage(name string) *Package {
	pkg := &Package{}
	pkg.Init(pkg)

	pkg.Name = name

	return pkg
}

var EdgeKindImplements = psi.DefineEdgeType[*Interface]("implements")
