package main

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/alecthomas/repr"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Boolean bool

func (b *Boolean) Capture(values []string) error {
	*b = values[0] == "TRUE"
	return nil
}

type Select struct {
	Path   *Expression `"SELECT" @@`
	Search *Expression `( "SEARCH" @@ )?`
	Limit  *Expression `( "LIMIT" @@ )?`
}

type Expression struct {
	Or []*OrCondition `@@ ( "OR" @@ )*`
}

type OrCondition struct {
	And []*Condition `@@ ( "AND" @@ )*`
}

type Condition struct {
	Operand *ConditionOperand `  @@`
	Not     *Condition        `| "NOT" @@`
	Exists  *Select           `| "EXISTS" "(" @@ ")"`
}

type ConditionOperand struct {
	Operand      *Operand      `@@`
	ConditionRHS *ConditionRHS `@@?`
}

type ConditionRHS struct {
	Compare *Compare `  @@`
	Is      *Is      `| "IS" @@`
	Between *Between `| "BETWEEN" @@`
	In      *In      `| "IN" "(" @@ ")"`
	Like    *Like    `| "LIKE" @@`
}

type Compare struct {
	Operator string         `@( "<>" | "<=" | ">=" | "=" | "<" | ">" | "!=" )`
	Operand  *Operand       `(  @@`
	Select   *CompareSelect ` | @@ )`
}

type CompareSelect struct {
	All    bool    `(  @"ALL"`
	Any    bool    ` | @"ANY"`
	Some   bool    ` | @"SOME" )`
	Select *Select `"(" @@ ")"`
}

type Like struct {
	Not     bool     `[ @"NOT" ]`
	Operand *Operand `@@`
}

type Is struct {
	Not          bool     `[ @"NOT" ]`
	Null         bool     `( @"NULL"`
	DistinctFrom *Operand `  | "DISTINCT" "FROM" @@ )`
}

type Between struct {
	Start *Operand `@@`
	End   *Operand `"AND" @@`
}

type In struct {
	Select      *Select       `  @@`
	Expressions []*Expression `| @@ ( "," @@ )*`
}

type Operand struct {
	Summand []*Summand `@@ ( "|" "|" @@ )*`
}

type Summand struct {
	LHS *Factor `@@`
	Op  string  `[ @("+" | "-")`
	RHS *Factor `  @@ ]`
}

type Factor struct {
	LHS *Term  `@@`
	Op  string `( @("*" | "/" | "%")`
	RHS *Term  `  @@ )?`
}

type Term struct {
	Select        *Select     `  @@`
	Value         *Value      `| @@`
	SymbolRef     *SymbolRef  `| @@`
	SubExpression *Expression `| "(" @@ ")"`
}

type SymbolRef struct {
	Symbol     string        `@Ident @( "." Ident )*`
	Parameters []*Expression `( "(" @@ ( "," @@ )* ")" )?`
}

type Value struct {
	Wildcard bool     `(  @"*"`
	Number   *float64 ` | @Number`
	String   *string  ` | @String`
	Boolean  *Boolean ` | @("TRUE" | "FALSE")`
	Null     bool     ` | @"NULL"`
	Array    *Array   ` | @@`
	Object   *Object  ` | @@`
	Path     *Path    ` | @@ )`
}

type Array struct {
	Expressions []*Expression `"(" @@ ( "," @@ )* ")"`
}

type PathElementQualifier struct {
	Key   *string `( @(Ident | String)`
	Type  *string `| (":" @(Ident | String))`
	Name  *string `| ("#" @(Ident | String))`
	Index *int64  `| ("@" @Number) )`
}

func (q *PathElementQualifier) ToPathElement(pe *psi.PathElement) {
	if q.Key != nil {
		pe.Name = *q.Key
	}

	if q.Type != nil {
		pe.Kind = psi.EdgeKind(*q.Type)
	}

	if q.Name != nil {
		pe.Name = *q.Name
	}

	if q.Index != nil {
		pe.Index = *q.Index
	}

	return
}

type PathElement struct {
	Qualifiers []PathElementQualifier `@@*`
}

func (pe *PathElement) ToPathElement() psi.PathElement {
	result := psi.PathElement{}

	for _, q := range pe.Qualifiers {
		q.ToPathElement(&result)
	}

	return result
}

type Path struct {
	Root     *string       `"[" (@Ident "//")?`
	Elements []PathElement `( @@ ( "/" @@ )* )? "]"`
}

func (pe *Path) ToPath() psi.Path {
	var elements []psi.PathElement

	for _, e := range pe.Elements {
		elements = append(elements, e.ToPathElement())
	}

	if pe.Root != nil {
		return psi.PathFromElements(*pe.Root, false, elements...)
	} else {
		return psi.PathFromElements("", false, elements...)
	}
}

type ObjectField struct {
	Key   string      `@(Ident | String | Number)`
	Value *Expression `":" @@`
}

type Object struct {
	Type   *string        `"{" ("[" @(Ident | String) "]")?`
	Fields []*ObjectField `(@@ ( "," @@ )*)? "}"`
}

var (
	sqlLexer = lexer.MustSimple([]lexer.SimpleRule{
		{`Keyword`, `(?i)\b(SELECT|FROM|TOP|DISTINCT|ALL|WHERE|GROUP|BY|HAVING|UNION|MINUS|EXCEPT|INTERSECT|ORDER|LIMIT|OFFSET|TRUE|FALSE|NULL|IS|NOT|ANY|SOME|BETWEEN|AND|OR|LIKE|AS|IN)\b`},
		{`Ident`, `[a-zA-Z_][a-zA-Z0-9_]*`},
		{`Number`, `[-+]?\d*\.?\d+([eE][-+]?\d+)?`},
		{`String`, `'[^']*'|"[^"]*"`},
		{`Operators`, `<>|!=|<=|>=|[-+*/%,.()=<>{}[\]@:#]`},
		{"PathRootSeparator", `//`},
		{"PathSeparator", `/`},
		{"whitespace", `\s+`},
	})
	parser = participle.MustBuild[Select](
		participle.Lexer(sqlLexer),
		participle.Unquote("String"),
		participle.CaseInsensitive("Keyword"),
		participle.UseLookahead(participle.MaxLookahead),
		// participle.Elide("Comment"),
	)
)

func main() {
	result, err := parser.ParseString("",
		`SELECT "ROOT//XYZ/FOO/BAR" SEARCH ({ [Text] foo: "foo" }) LIMIT 5`,
	)

	if err != nil {
		panic(err)
	}

	repr.Println(result, repr.Indent("  "), repr.OmitEmpty(true))
}
