package psi

import (
	"fmt"
)

type PathElement struct {
	Kind  EdgeKind
	Name  string
	Index int
}

func (p PathElement) String() string {
	return fmt.Sprintf("%s:%s@%d", p.Kind, p.Name, p.Index)
}

type Path []PathElement

func (p Path) Parent() Path {
	if len(p) == 0 {
		return p
	}

	return p[:len(p)-1]
}

func (p Path) Join(other Path) (res Path) {
	res = append(res, p...)
	res = append(res, other...)
	return
}

func (p Path) Child(name PathElement) (res Path) {
	res = append(res, p...)
	res = append(res, name)
	return
}

func (p Path) Components() []PathElement {
	return p
}

func (p Path) String() (res string) {
	for _, component := range p {
		res += "/" + component.String()
	}

	if res == "" {
		res = "/"
	}

	return
}
