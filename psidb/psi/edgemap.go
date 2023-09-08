package psi

type EdgeMap interface {
	Get(key EdgeKey) (value Edge, ok bool)
	Set(key EdgeKey, value Edge)
	Remove(key EdgeKey)
}
