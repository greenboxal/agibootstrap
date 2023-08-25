package stdlib

import "github.com/greenboxal/agibootstrap/pkg/psi"

type Error struct {
	psi.NodeBase

	Message string `json:"message"`
}

var ErrorType = psi.DefineNodeType[*Error]()

func NewErrorWithMessage(message string) *Error {
	t := &Error{
		Message: message,
	}

	t.Init(t, psi.WithNodeType(ErrorType))

	return t
}

func WrapError(err error) *Error {
	return NewErrorWithMessage(err.Error())
}
