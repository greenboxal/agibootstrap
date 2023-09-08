package typing

import (
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type Package struct {
	psi.NodeBase

	Name string `json:"name"`
}

var PackageType = psi.DefineNodeType[*Package]()

func NewPackage(name string) *Package {
	pkg := &Package{
		Name: name,
	}

	pkg.Init(pkg, psi.WithNodeType(PackageType))

	return pkg
}

func (t *Package) PsiNodeName() string { return t.Name }
