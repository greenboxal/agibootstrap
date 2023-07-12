package graphindex

import (
	"fmt"
	"math"
)

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

func (g GraphEmbedding) String() string {
	return fmt.Sprintf("Md=%v Mtd=%v Mr=%v Mt=%v Msema=%v", g.Depth, g.TreeDistance, g.ReferenceDistance, g.Time, g.Semantic)
}

func (g GraphEmbedding) Dimensions() int {
	return len(g.Semantic) + 8
}

func (g GraphEmbedding) ToFloat32Slice(dst []float32) []float32 {
	md1, md2 := rotary(int64(g.Depth), int64(len(g.Semantic)), 10000)
	mtd1, mtd2 := rotary(int64(g.TreeDistance), int64(len(g.Semantic)), 10000)
	mr1, mr2 := rotary(int64(g.ReferenceDistance), int64(len(g.Semantic)), 10000)
	mt1, mt2 := rotary(g.Time, int64(len(g.Semantic)), 10000)

	scale := float32(len(g.Semantic)) + 1

	dst = append(dst, md1/scale, md2/scale, mtd1/scale, mtd2/scale, mr1/scale, mr2/scale, mt1/scale, mt2/scale)
	dst = append(dst, g.Semantic...)

	return dst
}

func rotary(position, d, base int64) (float32, float32) {
	wk := 1.0 / -(math.Log(float64(base)) / float64(d))

	a := float32(math.Sin(wk * float64(position)))
	b := float32(math.Cos(wk * float64(position)))

	return a, b
}
