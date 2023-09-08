package typesystem

import (
	"reflect"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/samber/lo"
	"github.com/stoewer/go-strcase"

	"github.com/greenboxal/aip/aip-sdk/pkg/utils"
)

var pluralizeClient = pluralize.NewClient()

type TypeName struct {
	Package    string
	Name       string
	Plural     string
	Parameters []TypeName
}

func (n TypeName) WithParameters(parameters ...TypeName) TypeName {
	n.Parameters = parameters
	return n
}

func (n TypeName) ToTitle() string {
	return strcase.UpperCamelCase(n.Name)
}

func (n TypeName) ToTitlePlural() string {
	if n.Plural == "" {
		n.Plural = pluralizeClient.Plural(n.Name)
	}

	return strcase.UpperCamelCase(n.Plural)
}

func (n TypeName) FullName() string {
	if n.Package != "" {
		return strings.ReplaceAll(n.Package, "/", ".") + "." + n.Name
	}

	return n.Name
}

func (n TypeName) Args() []string {
	return lo.Map(n.Parameters, func(arg TypeName, _index int) string {
		return arg.String()
	})
}

func (n TypeName) NameWithArgs() string {
	if len(n.Parameters) > 0 {
		a := n.Args()

		args := strings.Join(a, "_QZQZ_")
		args = "_QZQZ_" + args + "_QZQZ_"

		return n.Name + args
	}

	return n.Name
}

func (n TypeName) FullNameWithArgs() string {
	if len(n.Parameters) > 0 {
		a := n.Args()

		args := strings.Join(a, "_QZQZ_")
		args = "_QZQZ_" + args + "_QZQZ_"

		return n.FullName() + args
	}

	return n.FullName()
}

func (n TypeName) GoString() string {
	return n.String()
}

func (n TypeName) String() string {
	args := ""

	if len(n.Parameters) > 0 {
		a := lo.Map(n.Parameters, func(arg TypeName, _index int) string {
			return arg.String()
		})

		args = strings.Join(a, ", ")
		args = "[" + args + "]"
	}

	return n.Name + args
}

func (n TypeName) NormalizedFullNameWithArguments() string {
	args := ""

	if len(n.Parameters) > 0 {
		a := lo.Map(n.Parameters, func(arg TypeName, _index int) string {
			return arg.String()
		})

		args = strings.Join(a, "_QZQZ_")
		args = "_QZQZ_" + args + "_QZQZ_"
	}

	return utils.NormalizeName(n.FullName() + args)
}

func AsTypeName(parsed utils.ParsedTypeName) TypeName {
	return TypeName{
		Name:    parsed.Name,
		Package: parsed.Pkg,
		Parameters: lo.Map(parsed.Args, func(item utils.ParsedTypeName, index int) TypeName {
			return AsTypeName(item)
		}),
	}
}

var packageTypeNameMap = map[string]string{
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs": "vfs",

	"github.com/greenboxal/agibootstrap/psidb/core/api": "psidb",

	"github.com/greenboxal/agibootstrap/psidb/core/":     "psidb.",
	"github.com/greenboxal/agibootstrap/psidb/db/":       "psidb.",
	"github.com/greenboxal/agibootstrap/psidb/services/": "psidb.",
	"github.com/greenboxal/agibootstrap/psidb/apps/":     "",
	"github.com/greenboxal/agibootstrap/psidb/modules/":  "",

	"github.com/greenboxal/agibootstrap/pkg/": "agib.",
}

func GetTypeName(typ reflect.Type) TypeName {
	parsed := utils.GetParsedTypeName(typ)

	if parsed.Pkg == "" {
		parsed.Pkg = "_rt_"
	}

	parsed.Pkg = rewritePackageName(parsed.Pkg)
	parsed.Pkg = strings.ReplaceAll(parsed.Pkg, "/", ".")

	return AsTypeName(parsed)
}

func ParseTypeName(name string) TypeName {
	parsed := utils.ParseTypeName(name)

	if parsed.Pkg == "" {
		parsed.Pkg = "_rt_"
	}

	parsed.Pkg = rewritePackageName(parsed.Pkg)
	parsed.Pkg = strings.ReplaceAll(parsed.Pkg, "/", ".")

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
