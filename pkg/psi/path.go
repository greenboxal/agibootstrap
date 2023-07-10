package psi

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

func MustParsePath(path string) Path {
	p, err := ParsePath(path)

	if err != nil {
		panic(err)
	}

	return p
}

func ParsePath(s string) (Path, error) {
	var err error

	rootSepIndex := strings.Index(s, "//")
	isRelative := rootSepIndex == -1
	root := ""
	path := s

	if !isRelative {
		root = s[:rootSepIndex]
		path = s[rootSepIndex+2:]
	}

	components, err := ParsePathElements(path)

	if err != nil {
		return Path{}, err
	}

	return PathFromElements(root, isRelative, components...), nil
}

func PathFromElements(root string, isRelative bool, components ...PathElement) Path {
	c := make([]PathElement, 0, len(components))
	c = append(c, components...)
	return Path{root: root, relative: isRelative, components: c}
}

type Path struct {
	relative   bool
	root       string
	components []PathElement
}

//goland:noinspection GoMixedReceiverTypes
func (p Path) IsEmpty() bool { return len(p.components) == 0 && p.relative }

//goland:noinspection GoMixedReceiverTypes
func (p Path) PrimitiveKind() typesystem.PrimitiveKind { return typesystem.PrimitiveKindString }

//goland:noinspection GoMixedReceiverTypes
func (p Path) Components() []PathElement { return p.components }

//goland:noinspection GoMixedReceiverTypes
func (p Path) Root() string { return p.root }

//goland:noinspection GoMixedReceiverTypes
func (p Path) WithRoot(root string) Path {
	p.root = root
	return p
}

//goland:noinspection GoMixedReceiverTypes
func (p Path) IsRelative() bool { return p.relative }

//goland:noinspection GoMixedReceiverTypes
func (p Path) Parent() Path {
	if len(p.components) == 0 {
		return p
	}

	return PathFromElements(p.root, p.relative, p.components[:len(p.components)-1]...)
}

//goland:noinspection GoMixedReceiverTypes
func (p Path) RelativeTo(other Path) (Path, error) {
	if p.root != other.root {
		return Path{}, fmt.Errorf("cannot make path %s relative to %s", p, other)
	}

	if !p.IsChildOf(other) {
		return Path{}, fmt.Errorf("cannot make path %s relative to %s", p, other)
	}

	return PathFromElements(p.root, true, p.components[len(other.components):]...), nil
}

//goland:noinspection GoMixedReceiverTypes
func (p Path) IsChildOf(other Path) bool {
	if p.root != other.root {
		return false
	}

	if len(p.components) <= len(other.components) {
		return false
	}

	for i, component := range other.components {
		if p.components[i] != component {
			return false
		}
	}

	return true
}

//goland:noinspection GoMixedReceiverTypes
func (p Path) Last() PathElement {
	if len(p.components) == 0 {
		return PathElement{}
	}

	return p.components[len(p.components)-1]
}

//goland:noinspection GoMixedReceiverTypes
func (p Path) Join(other Path) Path {
	if !p.relative && !other.relative {
		panic("cannot join two absolute paths")
	}

	root := p.root

	if p.relative {
		root = other.root
	}

	var components []PathElement
	components = append(components, p.components...)
	components = append(components, other.components...)
	return PathFromElements(root, p.relative && other.relative, components...)
}

//goland:noinspection GoMixedReceiverTypes
func (p Path) Child(name PathElement) Path {
	var components []PathElement
	components = append(components, p.components...)
	components = append(components, name)
	return PathFromElements(p.root, p.relative, components...)
}

//goland:noinspection GoMixedReceiverTypes
func (p Path) String() (res string) {
	for _, component := range p.components {
		res = path.Join(res, component.String())
	}

	if !p.relative {
		res = p.root + "//" + res
	}

	return
}

//goland:noinspection GoMixedReceiverTypes
func (p Path) MarshalJSON() ([]byte, error) { return []byte(fmt.Sprintf("\"%s\"", p.String())), nil }

//goland:noinspection GoMixedReceiverTypes
func (p *Path) UnmarshalJSON(data []byte) error {
	str := string(data)

	if str == "null" {
		return nil
	}

	return p.UnmarshalText(data[1 : len(data)-1])
}

//goland:noinspection GoMixedReceiverTypes
func (p Path) MarshalText() (text []byte, err error) { return []byte(p.String()), nil }

//goland:noinspection GoMixedReceiverTypes
func (p *Path) UnmarshalText(text []byte) error {
	parsed, err := ParsePath(string(text))

	if err != nil {
		return err
	}

	*p = parsed

	return nil
}

//goland:noinspection GoMixedReceiverTypes
func (p Path) MarshalBinary() (data []byte, err error) { return p.MarshalText() }

//goland:noinspection GoMixedReceiverTypes
func (p *Path) UnmarshalBinary(data []byte) error { return p.UnmarshalText(data) }

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

func (p PathElement) AsEdgeKey() EdgeKey {
	return EdgeKey{
		Kind:  p.Kind,
		Name:  p.Name,
		Index: p.Index,
	}
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
