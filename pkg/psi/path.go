package psi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/samber/lo"
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

func ParsePath(path string) Path {
	strs := strings.Split(path, "/")

	return lo.Map(strs, func(str string, _ int) PathElement {
		kindAndNameIndex := strings.Split(str, ":")
		nameAndIndex := strings.Split(kindAndNameIndex[1], "@")

		kind := EdgeKind(kindAndNameIndex[0])
		name := nameAndIndex[0]
		index, err := strconv.ParseInt(nameAndIndex[1], 10, 64)

		if err != nil {
			panic(err)
		}

		return PathElement{
			Kind:  kind,
			Name:  name,
			Index: int(index),
		}
	})
}

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
