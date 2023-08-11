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
		return n.Package + "." + n.Name
	}

	return n.Name
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

		args = strings.Join(a, "__")
		args = "___" + args + "___"
	}

	return utils.NormalizeName(n.FullName() + args)
}

func getTypeName(typ reflect.Type) TypeName {
	var asTypeName func(parsed utils.ParsedTypeName) TypeName

	asTypeName = func(parsed utils.ParsedTypeName) TypeName {
		return TypeName{
			Name:    parsed.Name,
			Package: parsed.Pkg,
			Parameters: lo.Map(parsed.Args, func(item utils.ParsedTypeName, index int) TypeName {
				return asTypeName(item)
			}),
		}
	}

	parsed := utils.GetParsedTypeName(typ)

	return asTypeName(parsed)
}
