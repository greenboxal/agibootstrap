---
authored-with: gpt-4
---
# Symbol Name Mangling Design Document

## Objective

Define a grammar for mangling symbol names in a dynamic linking context, while accommodating package structure, parametric types, and function signatures.

## Basic Rules

### Allowed Characters

The mangled names may only contain the following characters `[A-Za-z0-9_$]`.

### Prefix

Start all mangled names with a fixed prefix `MGL_`.

## Mangling Types and Packages

1. Replace each dot or slash in package names with an underscore.
    - Example: `foo.bar.Type1` becomes `foo_bar_Type1`
2. Convert uppercase letters in types and packages to lowercase and prepend them with a `$`.
    - Example: `Type1` becomes `$type1`
3. Append the original type at the end, capitalized.
    - Example: `MGL_foo_bar_$type1_Type1`

## Supporting Parametric Types

1. Use `P` as the delimiter to indicate the start of type parameters.
2. List type parameters separated by underscores.
3. Apply the existing mangling rules to each type parameter.

- Example: `Foo<A, B>` becomes `MGL_foo_P$a_$b_Foo`
- Nested example: `Foo<A, Bar<C, D>>` becomes `MGL_foo_P$a_bar_P$c_$d_Bar_Foo`

## Supporting Function Signatures

1. Use `F` as the delimiter to indicate a function.
2. List function parameter types separated by underscores.
3. Use `R` to separate parameter types from the return type.
4. Apply existing mangling rules to each type in the function signature.

- Example: `func(a: A, b: B) -> C` becomes `MGL_foo_F$a_$b_R$c_Func`
- Parametric example: `func(a: Foo<A, B>, b: Bar) -> Baz` becomes `MGL_func_P$foo_P$a_$b_Foo_$bar_R$baz_Func`
- Nested example: `func(a: Foo<A, Bar<B, C>>) -> D` becomes `MGL_pkg_F$foo_P$a_bar_P$b_$c_Bar_R$d_Func`

## EBNF Grammar for Symbol Name Mangling

Here is the Extended Backus-Naur Form (EBNF) for the mangling grammar. This provides a formal representation for all discussed rules.

```
MangledName  = "MGL_", Package, "_", MangledType, ["_", FunctionSignature] ;
Package      = Identifier, {("_", Identifier)} ;
MangledType  = LowercasePrefix, Identifier, ["_P", ParamList] ;

LowercasePrefix = "$", LowercaseIdentifier ;

ParamList    = MangledType, {("_", MangledType)} ;

FunctionSignature = "F", ParamList, "_R", MangledType ;

Identifier   = UpperCaseLetter, {LetterOrDigit} ;
LowercaseIdentifier = LowerCaseLetter, {LetterOrDigit} ;

UpperCaseLetter = "A" | "B" | ... | "Z" ;
LowerCaseLetter = "a" | "b" | ... | "z" ;
LetterOrDigit = UpperCaseLetter | LowerCaseLetter | Digit ;

Digit        = "0" | "1" | ... | "9" ;
```

### Legend:

- `MangledName`: The complete mangled symbol name.
- `Package`: Represents the package path.
- `MangledType`: Represents a mangled type including parametric types.
- `LowercasePrefix`: Prefix for mangled type identifier, using lowercase and prepending "$".
- `ParamList`: List of mangled types used in parametric types or function parameters.
- `FunctionSignature`: Represents the mangled function signature.
- `Identifier`: Represents a normal identifier, starting with an uppercase letter.
- `LowercaseIdentifier`: Represents an identifier starting with a lowercase letter.
- `UpperCaseLetter`: Any uppercase alphabetic character.
- `LowerCaseLetter`: Any lowercase alphabetic character.
- `LetterOrDigit`: Any alphanumeric character.
- `Digit`: Numeric characters 0-9.

This EBNF should encapsulate all rules we've discussed for mangling names.

## Final Thoughts

This design aims to provide a systematic and scalable approach to mangling symbol names, ensuring uniqueness and readability while maintaining a restricted character set.
