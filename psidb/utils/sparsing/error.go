package sparsing

import (
	"fmt"

	"github.com/pkg/errors"
)

type ParsingError struct {
	Err                 error
	Position            Position
	RecoverablePosition Position
}

func (pe *ParsingError) Error() string  { return pe.Err.Error() }
func (pe *ParsingError) Unwrap() error  { return pe.Err }
func (pe *ParsingError) String() string { return fmt.Sprintf("ParsingError: %v", pe.Err) }

var ErrInvalidToken = errors.New("invalid token")
var ErrNoTokenHandler = errors.New("no token handler")
var ErrNoLexerHandler = errors.New("no lexer handler")
var ErrNoNodeHandler = errors.New("no node handler")
var ErrEndOfStream = errors.New("end of stream")

type RollbackError struct {
	Err error
}

func (e *RollbackError) Error() string { return e.Err.Error() }
func (e *RollbackError) Unwrap() error { return e.Err }
