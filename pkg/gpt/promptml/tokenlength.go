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

func (s TokenLength) Add(other TokenLength) TokenLength {
	if s.Unit != other.Unit {
		panic("cannot add token lengths with different units")
	}

	return NewTokenLength(s.Amount+other.Amount, s.Unit)
}

func (s TokenLength) Sub(other TokenLength) TokenLength {
	if s.Unit != other.Unit {
		panic("cannot subtract token lengths with different units")
	}

	return NewTokenLength(s.Amount-other.Amount, s.Unit)
}

func (s TokenLength) MulScalar(scalar float64) TokenLength {
	return NewTokenLength(s.Amount*scalar, s.Unit)
}

func (s TokenLength) MulInt(scalar int) TokenLength {
	return s.Mul(NewTokenLength(float64(scalar), TokenUnitToken))
}

func (s TokenLength) Mul(other TokenLength) TokenLength {
	if s.Unit != other.Unit {
		if s.Unit == TokenUnitPercent {
			return NewTokenLength(s.Amount*other.Amount, other.Unit)
		} else if other.Unit == TokenUnitPercent {
			return other.Mul(s)
		}

		if s.Unit == TokenUnitUndefined {
			return other
		} else if other.Unit == TokenUnitUndefined {
			return s
		}

		panic("cannot multiply token lengths with different units")
	}

	return NewTokenLength(s.Amount*other.Amount, s.Unit)
}

func (s TokenLength) Div(other TokenLength) TokenLength {
	if s.Unit != other.Unit {
		panic("cannot divide token lengths with different units")
	}

	return NewTokenLength(s.Amount/other.Amount, s.Unit)
}

func (s TokenLength) Mod(other TokenLength) TokenLength {
	if s.Unit != other.Unit {
		panic("cannot modulo token lengths with different units")
	}

	return NewTokenLength(float64(int(s.Amount)%int(other.Amount)), s.Unit)
}

func (s TokenLength) IsZero() bool        { return s.Amount == 0 }
func (s TokenLength) IsNegative() bool    { return s.Amount < 0 }
func (s TokenLength) IsPositive() bool    { return s.Amount > 0 }
func (s TokenLength) IsNonNegative() bool { return s.Amount >= 0 }
func (s TokenLength) IsNonPositive() bool { return s.Amount <= 0 }
func (s TokenLength) IsUnbounded() bool   { return s.Unit == TokenUnitUndefined }
func (s TokenLength) IsReal() bool        { return s.Unit == TokenUnitToken }

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

func (s TokenLength) TokenCount() int {
	switch s.Unit {
	case TokenUnitToken:
		return int(s.Amount)

	}

	return 0
}

type Bounds struct {
	Position int
	Length   TokenLength
}
