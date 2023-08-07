package psi

import (
	"fmt"
	"net/url"
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
	return ParsePathEx(s, false)
}

func ParsePathEx(s string, escaped bool) (Path, error) {
	var err error

	rootSepIndex := strings.Index(s, "//")
	isRelative := rootSepIndex == -1
	root := ""
	path := s

	if !isRelative {
		root = s[:rootSepIndex]
		path = s[rootSepIndex+2:]
	}

	components, err := ParsePathElements(path, escaped)

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
func (p Path) Depth() int { return len(p.components) }

//goland:noinspection GoMixedReceiverTypes
func (p Path) IsEmpty() bool { return len(p.components) == 0 && p.root == "" }

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
	return p.Format(false)
}

//goland:noinspection GoMixedReceiverTypes
func (p Path) Format(escaped bool) (res string) {
	for _, component := range p.components {
		res = path.Join(res, url.PathEscape(component.String()))
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

//goland:noinspection GoMixedReceiverTypes
func (p Path) GetCommonAncestor(other Path) (Path, bool) {
	if p.root != other.root {
		return Path{}, false
	}

	components := make([]PathElement, 0, len(p.components))

	for i := 0; i < len(p.components) && i < len(other.components); i++ {
		if p.components[i] != other.components[i] {
			break
		}

		components = append(components, p.components[i])
	}

	return PathFromElements(p.root, p.relative && other.relative, components...), true
}

//goland:noinspection GoMixedReceiverTypes
func (p Path) IsAncestorOf(child Path) bool {
	if p.root != child.root {
		return false
	}

	if len(p.components) >= len(child.components) {
		return false
	}

	for i, component := range p.components {
		if child.components[i] != component {
			return false
		}
	}

	return true
}

//goland:noinspection GoMixedReceiverTypes
func (p Path) ForEachElement(f func(pe PathElement)) {
	for _, component := range p.components {
		f(component)
	}
}

//goland:noinspection GoMixedReceiverTypes
func (p Path) ForEachNode(f func(pe Path)) {
	for i := range p.components {
		f(p.Slice(0, i+1))
	}
}

//goland:noinspection GoMixedReceiverTypes
func (p Path) Slice(first int, last int) Path {
	if last == -1 {
		last = len(p.components)
	}

	if first == 0 && last == len(p.components) {
		return p
	}

	if first < 0 {
		panic("first must be >= 0")
	}

	if last < 0 {
		panic("last must be >= 0")
	}

	if first > last {
		panic("first must be <= last")
	}

	root := ""
	relative := p.relative || first > 0

	if !p.relative && first == 0 {
		root = p.root
	}

	return PathFromElements(root, relative, p.components[first:last]...)
}

func (p Path) Len() int {
	return len(p.components)
}

func (p Path) Name() EdgeKey {
	if len(p.components) == 0 {
		return EdgeKey{}
	}

	return p.components[len(p.components)-1].AsEdgeKey()
}

func (p Path) Equals(p2 Path) bool {
	return p.String() == p2.String()
}

func (p Path) CompareTo(j Path) int {
	if p.root != j.root {
		return strings.Compare(p.root, j.root)
	}

	if p.relative != j.relative {
		if p.relative {
			return -1
		} else {
			return 1
		}
	}

	for i := 0; i < len(p.components) && i < len(j.components); i++ {
		cmp := p.components[i].CompareTo(j.components[i])

		if cmp != 0 {
			return cmp
		}
	}

	return len(p.components) - len(j.components)
}

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

	if p.Kind == EdgeKindChild && p.Name != "" {
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

func (p PathElement) CompareTo(j PathElement) int {
	return strings.Compare(p.String(), j.String())
}

func ParsePathElements(s string, escaped bool) (result []PathElement, err error) {
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

		if escaped {
			part, err = url.PathUnescape(part)

			if err != nil {
				return nil, err
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

var RelativePathToSelf = PathFromElements("", true, PathElement{Name: "."})
var RelativePathToParent = PathFromElements("", true, PathElement{Name: ".."})
