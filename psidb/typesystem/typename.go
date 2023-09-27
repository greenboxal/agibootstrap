package typesystem

import (
	"path"
	"reflect"
	"strings"

	"github.com/samber/lo"
	"github.com/stoewer/go-strcase"

	"github.com/greenboxal/aip/aip-sdk/pkg/utils"
)

type TypeClass int

const (
	TypeClassUnknown TypeClass = iota
	TypeClassObject
	TypeClassFunction
	TypeClassArray
	TypeClassMap
	TypeClassInterface
	TypeClassPointer
	TypeClassReference
)

func (tc TypeClass) Prefix() string {
	switch tc {
	case TypeClassObject:
		return "O"
	case TypeClassFunction:
		return "F"
	case TypeClassInterface:
		return "I"
	case TypeClassArray:
		return "A"
	case TypeClassMap:
		return "M"
	case TypeClassPointer:
		return "P"
	case TypeClassReference:
		return "R"
	default:
		return ""
	}
}

type TypeName struct {
	Class   TypeClass
	Package string
	Name    string

	InParameters  []TypeName
	OutParameters []TypeName
}

func ParseMangledName(mangledName string) (tn TypeName, rest string) {
	rest = mangledName

	parenIndex := strings.IndexByte(rest, '(')

	if parenIndex == -1 {
		parenIndex = len(rest)
	}

	fullName := strings.Split(rest[:parenIndex], "/")
	rest = rest[parenIndex:]

	tn.Package = strings.Join(fullName[:len(fullName)-1], "/")
	tn.Name = fullName[len(fullName)-1]

	if len(rest) == 0 {
		return
	}

	if rest[0] != '(' {
		return
	}

	rest = rest[1:]
	tn.InParameters = []TypeName{}

	for len(rest) > 0 {
		var nested TypeName

		if rest[0] == ')' {
			if tn.OutParameters == nil {
				rest = rest[1:]
				tn.OutParameters = []TypeName{}
			} else {
				break
			}
		} else if rest[0] == ';' {
			return
		}

		if len(rest) == 0 {
			break
		}

		nested, rest = ParseMangledName(rest)

		if tn.OutParameters == nil {
			tn.InParameters = append(tn.InParameters, nested)
		} else {
			tn.OutParameters = append(tn.OutParameters, nested)
		}

		if len(rest) > 0 && rest[0] == ';' {
			rest = rest[1:]
		}
	}

	if tn.OutParameters == nil {
		if len(rest) > 0 {
			if rest[0] != ')' {
				panic("invalid mangled name")
			}

			rest = rest[1:]
		}
	}

	return
}

func (n TypeName) MangledName() string {
	base := n.Class.Prefix() + path.Join(n.Package, n.Name)
	args := ""

	if len(n.InParameters) > 0 || len(n.OutParameters) > 0 {
		in := lo.Map(n.InParameters, func(arg TypeName, _index int) string {
			return arg.MangledName()
		})

		out := lo.Map(n.OutParameters, func(arg TypeName, _index int) string {
			return arg.MangledName()
		})

		args = "(" + strings.Join(in, ";") + ")" + strings.Join(out, ";")
	}

	return base + args
}

func (n TypeName) WithInParameters(parameters ...TypeName) TypeName {
	n.InParameters = parameters
	return n
}

func (n TypeName) WithOutParameters(parameters ...TypeName) TypeName {
	n.OutParameters = parameters
	return n
}

func (n TypeName) ToTitle() string {
	return strcase.UpperCamelCase(n.Name)
}

func (n TypeName) FullName() string {
	return path.Join(n.Package, n.Name)
}

func (n TypeName) Args() []string {
	return lo.Map(n.InParameters, func(arg TypeName, _index int) string {
		return arg.String()
	})
}

func (n TypeName) NameWithArgs() string {
	args := ""

	if len(n.InParameters) > 0 || len(n.OutParameters) > 0 {
		in := lo.Map(n.InParameters, func(arg TypeName, _index int) string {
			return arg.MangledName()
		})

		out := lo.Map(n.OutParameters, func(arg TypeName, _index int) string {
			return arg.MangledName()
		})

		args = "(" + strings.Join(in, ";") + ")" + strings.Join(out, ";")
	}

	return n.Name + args
}

func (n TypeName) FullNameWithArgs() string {
	return n.MangledName()
}

func (n TypeName) GoString() string {
	return n.String()
}

func (n TypeName) String() string {
	args := ""

	if len(n.InParameters) > 0 {
		a := lo.Map(n.InParameters, func(arg TypeName, _index int) string {
			return arg.String()
		})

		args = strings.Join(a, ", ")
		args = "[" + args + "]"
	}

	return n.Name + args
}

func (n TypeName) NormalizedFullNameWithArguments() string {
	return n.FullNameWithArgs()
}

func AsTypeName(parsed utils.ParsedTypeName) TypeName {
	for len(parsed.Pkg) > 0 && parsed.Pkg[0] == '*' {
		return TypeName{
			Name: "ptr",
			InParameters: []TypeName{AsTypeName(utils.ParsedTypeName{
				Name: parsed.Name,
				Pkg:  parsed.Pkg[1:],
			})},
		}
	}

	parsed.Pkg = rewritePackageName(parsed.Pkg)

	tn := TypeName{
		Name:    parsed.Name,
		Package: parsed.Pkg,
	}

	if len(parsed.Args) > 0 {
		tn.InParameters = lo.Map(parsed.Args, func(item utils.ParsedTypeName, index int) TypeName {
			return AsTypeName(item)
		})
	}

	return tn
}

var packageTypeNameMap = map[string]string{
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs": "vfs",

	"github.com/greenboxal/agibootstrap/psidb/core/api": "psidb",

	"github.com/greenboxal/agibootstrap/psidb/core/":     "psidb/",
	"github.com/greenboxal/agibootstrap/psidb/db/":       "psidb/",
	"github.com/greenboxal/agibootstrap/psidb/services/": "psidb/",
	"github.com/greenboxal/agibootstrap/psidb/apps/":     "",
	"github.com/greenboxal/agibootstrap/psidb/modules/":  "",

	"github.com/greenboxal/agibootstrap/pkg/": "agib.",
}

func GetTypeName(typ reflect.Type) TypeName {
	parsed := utils.GetParsedTypeName(typ)

	tn := AsTypeName(parsed)

	if typ.Kind() == reflect.Func {
		if typ.NumIn() > 0 {
			tn.InParameters = make([]TypeName, typ.NumIn())

			for i := 0; i < typ.NumIn(); i++ {
				tn.InParameters[i] = GetTypeName(typ.In(i))
			}
		}

		if typ.NumOut() > 0 {
			tn.OutParameters = make([]TypeName, typ.NumOut())

			for i := 0; i < typ.NumOut(); i++ {
				tn.OutParameters[i] = GetTypeName(typ.Out(i))
			}
		}
	} else if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array || typ.Kind() == reflect.Pointer {
		tn.InParameters = []TypeName{GetTypeName(typ.Elem())}
	} else if typ.Kind() == reflect.Map {
		tn.InParameters = []TypeName{GetTypeName(typ.Key()), GetTypeName(typ.Elem())}
	}

	return tn
}

func ParseTypeName(name string) TypeName {
	parsed := utils.ParseTypeName(name)

	parsed.Pkg = rewritePackageName(parsed.Pkg)

	return AsTypeName(parsed)
}

func rewritePackageName(pkg string) string {
	longest := ""

	for k := range packageTypeNameMap {
		if strings.HasPrefix(pkg, k) && len(k) > len(longest) {
			longest = k
		}
	}

	if longest == "" {
		return pkg
	}

	return packageTypeNameMap[longest] + strings.TrimPrefix(pkg, longest)
}
