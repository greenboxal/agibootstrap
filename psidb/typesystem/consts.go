package typesystem

type PrimitiveKind int

const (
	PrimitiveKindInvalid PrimitiveKind = iota
	PrimitiveKindBoolean
	PrimitiveKindBytes
	PrimitiveKindString
	PrimitiveKindInt
	PrimitiveKindUnsignedInt
	PrimitiveKindFloat
	PrimitiveKindList
	PrimitiveKindMap
	PrimitiveKindStruct
	PrimitiveKindInterface
	PrimitiveKindLink
	PrimitiveKindFunction
	PrimitiveKindVector
)

func (p PrimitiveKind) MarshalJSON() ([]byte, error) {
	return []byte(`"` + p.String() + `"`), nil
}

func (p *PrimitiveKind) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case `"invalid"`:
		*p = PrimitiveKindInvalid
	case `"boolean"`:
		*p = PrimitiveKindBoolean
	case `"bytes"`:
		*p = PrimitiveKindBytes
	case `"string"`:
		*p = PrimitiveKindString
	case `"int"`:
		*p = PrimitiveKindInt
	case `"uint"`:
		*p = PrimitiveKindUnsignedInt
	case `"float"`:
		*p = PrimitiveKindFloat
	case `"list"`:
		*p = PrimitiveKindList
	case `"map"`:
		*p = PrimitiveKindMap
	case `"struct"`:
		*p = PrimitiveKindStruct
	case `"interface"`:
		*p = PrimitiveKindInterface
	case `"link"`:
		*p = PrimitiveKindLink
	case `"function"`:
		*p = PrimitiveKindFunction
	case `"vector"`:
		*p = PrimitiveKindFunction
	}

	panic("unknown primitive kind")
}

func (p PrimitiveKind) String() string {
	switch p {
	case PrimitiveKindInvalid:
		return "invalid"
	case PrimitiveKindBoolean:
		return "boolean"
	case PrimitiveKindBytes:
		return "bytes"
	case PrimitiveKindString:
		return "string"
	case PrimitiveKindInt:
		return "int"
	case PrimitiveKindUnsignedInt:
		return "uint"
	case PrimitiveKindFloat:
		return "float"
	case PrimitiveKindList:
		return "list"
	case PrimitiveKindMap:
		return "map"
	case PrimitiveKindStruct:
		return "struct"
	case PrimitiveKindInterface:
		return "interface"
	case PrimitiveKindLink:
		return "link"
	case PrimitiveKindFunction:
		return "function"
	case PrimitiveKindVector:
		return "vector"
	}

	return "unknown"
}
