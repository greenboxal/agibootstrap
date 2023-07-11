package graphindex

import "math"

type Embedding interface {
	Dimensions() int
	ToFloat32Slice(dst []float32) []float32
}

type GraphEmbedding struct {
	Depth             int       `json:"depth,omitempty"`
	TreeDistance      int       `json:"treeDistance,omitempty"`
	ReferenceDistance int       `json:"referenceDistance,omitempty"`
	Time              int64     `json:"time,omitempty"`
	Semantic          []float32 `json:"semantic,omitempty"`
}

func (g GraphEmbedding) Dimensions() int {
	return len(g.Semantic) + 8
}

func (g GraphEmbedding) ToFloat32Slice(dst []float32) []float32 {
	md1, md2 := rotary(int64(g.Depth), 10, 2)
	mtd1, mtd2 := rotary(int64(g.TreeDistance), 10, 2)
	mr1, mr2 := rotary(int64(g.ReferenceDistance), 10, 2)
	mt1, mt2 := rotary(g.Time, 10, 2)

	dst = append(dst, md1, md2, mtd1, mtd2, mr1, mr2, mt1, mt2)
	dst = append(dst, g.Semantic...)

	return dst
}

func rotary(position, d, base int64) (float32, float32) {
	wk := 1.0 / -(math.Log(float64(base)) / float64(d))

	a := float32(math.Sin(wk * float64(position)))
	b := float32(math.Cos(wk * float64(position)))

	return a, b
}
