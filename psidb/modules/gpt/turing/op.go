package turing

type MicroOp int

const (
	MicroOpInvalid MicroOp = iota
	MicroOpPush
	MicroOpPushIn
	MicroOpPopIn
	MicroOpPopIntToOut
	MicroOpPushOut
	MicroOpPopOut
	MicroOpInfer
	MicroOpCall
	MicroOpReturn
	MicroOpAbort
)

type MicroOperation struct {
	Op MicroOp
	V  Value
}
