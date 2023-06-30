package promptml

import "fmt"

type TokenUnit string

const (
	TokenUnitUndefined TokenUnit = ""
	TokenUnitChar      TokenUnit = "c"
	TokenUnitWord      TokenUnit = "w"
	TokenUnitLine      TokenUnit = "l"
	TokenUnitToken     TokenUnit = "t"
	TokenUnitPercent   TokenUnit = "p"
)

func NewTokenLength(amount float64, unit TokenUnit) TokenLength {
	return TokenLength{
		Amount: amount,
		Unit:   unit,
	}
}

type TokenLength struct {
	Amount float64
	Unit   TokenUnit
}

func (s TokenLength) String() string {
	return fmt.Sprintf("%f%s", s.Amount, s.Unit)
}

func (s TokenLength) GetEffectiveLength(relFn func(float64) int, undefFn func() int) int {
	switch s.Unit {
	case TokenUnitUndefined:
		return undefFn()

	case TokenUnitToken:
		return int(s.Amount)

	case TokenUnitPercent:
		return relFn(s.Amount)
	}

	panic("invalid token unit")
}

type Bounds struct {
	Position int
	Length   TokenLength
}
