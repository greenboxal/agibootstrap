package typesystem

type OperatorName string

const (
	OperatorInvalid OperatorName = ""
	OperatorAdd     OperatorName = "+"
	OperatorSub     OperatorName = "-"
	OperatorMul     OperatorName = "*"
	OperatorDiv     OperatorName = "/"
	OperatorRem     OperatorName = "%"
	OperatorAnd     OperatorName = "&&"
	OperatorOr      OperatorName = "||"
	OperatorNot     OperatorName = "!"
	OperatorEq      OperatorName = "=="
	OperatorNeq     OperatorName = "!="
	OperatorNe      OperatorName = "!="
	OperatorLt      OperatorName = "<"
	OperatorGt      OperatorName = ">"
	OperatorLe      OperatorName = "<="
	OperatorGe      OperatorName = ">="
)

type Operator interface {
	Name() OperatorName

	Commutator() OperatorName
	Negator() OperatorName

	ReceiverType() Type
	Parameters() []Type

	Call(args ...Value) (Value, error)
}

type UnaryOperator interface {
	Operator

	Operand() Type
}

type BinaryOperator interface {
	Operator

	Left() Type
	Right() Type
}
