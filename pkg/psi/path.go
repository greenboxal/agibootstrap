package psi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// PathElement is a string of the form:
//
//	:<kind>#<name>@<index>
//
// Where:
//
//	<kind> is the edge kind (optional, defaults to "child")
//	<name> is the name of the node (optional)
//	<index> is the index of the node (optional)
//
// If no special characters are present, the component is assumed to be a name.
//
// Examples:
//
//	:child#foo@0
//	:child#foo
//	:child@0
//	:child
//	#foo
//	@0
//	foo
type PathElement struct {
	Kind  EdgeKind
	Name  string
	Index int64
}

func (p PathElement) String() string {
	str := ""

	if p.Kind == EdgeKindChild && p.Name != "" && p.Index == 0 {
		return p.Name
	}

	if p.Kind != "" && p.Kind != EdgeKindChild {
		str += fmt.Sprintf(":%s", p.Kind)
	}

	if p.Name != "" {
		str += fmt.Sprintf("#%s", p.Name)
	}

	if p.Index != 0 {
		str += fmt.Sprintf("@%d", p.Index)
	}

	if str == "" {
		str = fmt.Sprintf("@%d", p.Index)
	}

	return str
}

func (p PathElement) IsEmpty() bool {
	return p.Kind == "" && p.Name == "" && p.Index == 0
}

func ParsePathElements(s string) (result []PathElement, err error) {
	s = strings.TrimSpace(s)
	s = strings.TrimRight(s, "/")

	if s == "" {
		return nil, nil
	}

	sp := strings.Split(s, "/")
	result = make([]PathElement, 0, len(sp))

	for i, part := range sp {
		part = strings.TrimSpace(part)

		if part == "" {
			if i == 0 {
				continue
			} else {
				return nil, errors.New("empty Path component")
			}
		}

		c, e := ParsePathElement(part)

		if e != nil {
			err = multierror.Append(err, e)
		}

		result = append(result, c)
	}

	return
}

// ParsePathElement parses a single Path component.
func ParsePathElement(str string) (e PathElement, err error) {
	state := '#'
	acc := ""

	e.Kind = EdgeKindChild

	endState := func() {
		if acc == "" {
			return
		}

		switch state {
		case '@':
			e.Index, err = strconv.ParseInt(acc, 10, 64)

			if err != nil {
				return
			}
		case '#':
			e.Name += acc
		case ':':
			e.Kind = EdgeKind(acc)
		}

		acc = ""
	}

	for _, ch := range str {
		switch ch {
		case '@':
			fallthrough
		case '#':
			fallthrough
		case ':':
			endState()

			state = ch

			continue

		default:
			acc += string(ch)
		}
	}

	endState()

	return
}

func MustParsePath(path string) Path {
	p, err := ParsePath(path)

	if err != nil {
		panic(err)
	}

	return p
}

func ParsePath(s string) (Path, error) {
	var err error

	components, err := ParsePathElements(s)

	if err != nil {
		return Path{}, err
	}

	return PathFromElements(components...), nil
}

func PathFromElements(components ...PathElement) Path {
	c := make([]PathElement, 0, len(components))
	c = append(c, components...)
	return Path{components: c}
}

type Path struct {
	components []PathElement
}

func (p Path) PrimitiveKind() typesystem.PrimitiveKind { return typesystem.PrimitiveKindString }
func (p Path) Components() []PathElement               { return p.components }

func (p Path) Parent() Path {
	if len(p.components) == 0 {
		return p
	}

	return PathFromElements(p.components[:len(p.components)-1]...)
}

func (p Path) Join(other Path) Path {
	var components []PathElement
	components = append(components, p.components...)
	components = append(components, other.components...)
	return PathFromElements(components...)
}

func (p Path) Child(name PathElement) Path {
	var components []PathElement
	components = append(components, p.components...)
	components = append(components, name)
	return PathFromElements(components...)
}

func (p Path) String() (res string) {
	for _, component := range p.components {
		res += "/" + component.String()
	}

	if res == "" {
		res = "/"
	}

	return
}

func (p Path) IsEmpty() bool {
	return len(p.components) == 0
}

func (p Path) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", p.String())), nil
}

func (p *Path) UnmarshalJSON(data []byte) error {
	str := string(data)

	if str == "null" {
		return nil
	}

	str = str[1 : len(str)-1]
	components, err := ParsePathElements(str)

	if err != nil {
		return err
	}

	p.components = components

	return nil
}

func (p Path) MarshalText() (text []byte, err error) {
	return []byte(p.String()), nil
}

func (p *Path) UnmarshalText(text []byte) error {
	components, err := ParsePathElements(string(text))

	if err != nil {
		return err
	}

	p.components = components

	return nil
}

func (p Path) MarshalBinary() (data []byte, err error) {
	return p.MarshalText()
}

func (p *Path) UnmarshalBinary(data []byte) error {
	return p.UnmarshalText(data)
}

func ResolveChild(parent Path, child PathElement) Path {
	return parent.Child(child)
}

func ResolveEdge[T Node](parent Node, key TypedEdgeKey[T]) (def T) {
	e := parent.GetEdge(key)

	if e == nil {
		return
	}

	return e.To().(T)
}

func Resolve(root Node, path string) (Node, error) {
	p, err := ParsePath(path)

	if err != nil {
		return nil, err
	}

	return ResolvePath(root, p)
}

func ResolvePath(root Node, path Path) (Node, error) {
	if root == nil {
		return nil, ErrNodeNotFound
	}

	result := root

	for i, component := range path.components {
		if component.IsEmpty() {
			if i == 0 {
				component.Name = "/"
			} else {
				panic("empty Path component")
			}
		}

		if component.Kind == EdgeKindChild && component.Index == 0 {
			if component.Name == "/" {
				result = root
			} else if component.Name == "." {
				continue
			} else if component.Name == ".." {
				cn := result.Parent()

				if cn != nil {
					result = cn
				}

				continue
			}
		}

		cn := result.ResolveChild(component)

		if cn == nil {
			break
		}

		result = cn
	}

	if result == nil {
		return nil, ErrNodeNotFound
	}

	return result, nil
}
