package psi

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/samber/lo"
)

type PathElement struct {
	Kind  EdgeKind
	Name  string
	Index int64
}

func (p PathElement) String() string {
	str := ""

	if p.Kind != "" && p.Kind != EdgeKindChild {
		str += fmt.Sprintf(":%s", p.Kind)
	}

	if p.Name != "" {
		str += fmt.Sprintf("#%s", p.Name)
	}

	if p.Index != 0 {
		str += fmt.Sprintf("@%d", p.Index)
	}

	return str
}

func (p PathElement) IsEmpty() bool {
	return p.Kind == "" && p.Name == "" && p.Index == 0
}

type Path struct {
	root       Node
	components []PathElement
}

//goland:noinspection GoMixedReceiverTypes
func (p Path) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", p.String())), nil
}

//goland:noinspection GoMixedReceiverTypes
func (p *Path) UnmarshalJSON(data []byte) error {
	str := string(data)

	if str == "null" {
		return nil
	}

	str = strings.Trim(str, "\"")
	parsed, err := ParsePath(str)

	if err != nil {
		return err
	}

	p.root = parsed.root
	p.components = parsed.components

	return nil
}

func ParsePathComponent(str string) (e PathElement, err error) {
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

func ParsePath(path string) (Path, error) {
	var err error

	strs := strings.Split(path, "/")

	if strs[0] == "" {
		strs = strs[1:]
	}

	components := lo.Map(strs, func(str string, _ int) PathElement {
		c, e := ParsePathComponent(str)

		if e != nil {
			e = multierror.Append(err, e)
		}

		return c
	})

	return PathFromComponents(components...), err
}

func PathFromComponents(components ...PathElement) Path {
	c := make([]PathElement, 0, len(components))
	c = append(c, components...)
	return Path{
		components: c,
	}
}

func (p Path) Parent() Path {
	if len(p.components) == 0 {
		return p
	}

	return PathFromComponents(p.components[:len(p.components)-1]...)
}

func (p Path) Join(other Path) (res Path) {
	var components []PathElement
	components = append(components, p.components...)
	components = append(components, other.components...)
	return PathFromComponents(components...)
}

func (p Path) Child(name PathElement) (res Path) {
	var components []PathElement
	components = append(components, p.components...)
	components = append(components, name)
	return PathFromComponents(components...)
}

func (p Path) Components() []PathElement {
	return p.components
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

func (p Path) WithRoot(root Node) Path {
	return Path{
		root:       root,
		components: p.components,
	}
}

func (p Path) Root() Node {
	return p.root
}

func (p Path) IsEmpty() bool {
	if len(p.components) == 0 {
		return true
	}

	for _, c := range p.components {
		if !c.IsEmpty() {
			return false
		}
	}

	return true
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
	if path.root != nil {
		root = path.root
	}

	if root == nil {
		return nil, ErrNodeNotFound
	}

	result := root

	for i, component := range path.components {
		if component.IsEmpty() {
			if i == 0 {
				component.Name = "/"
			} else {
				panic("empty path component")
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
