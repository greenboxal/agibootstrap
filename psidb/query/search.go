package query

import (
	"math"

	"golang.org/x/exp/slices"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path/dynamic"
	"gonum.org/v1/gonum/graph/simple"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type Hit struct {
	psi.Node

	Score float32
}

type ScoreFunc func(ctx QueryContext, node psi.Node) float32
type CostFunc func(ctx QueryContext, node, goal psi.Node) float32

type Search struct {
	Filter FilterFunc `json:"filter"`
	Score  ScoreFunc  `json:"score"`
	Cost   CostFunc   `json:"cost"`
	Limit  int        `json:"limit"`

	Goal psi.Node `json:"goal"`
}

func (s Search) Run(ctx QueryContext, in Iterator) (Iterator, error) {
	capHint := s.Limit

	if capHint == -1 {
		capHint = in.Len()
	}

	if capHint == -1 {
		capHint = 2
	}

	ch := make(chan IteratorItem)

	go func() {
		defer close(ch)

		result := make([]Hit, 0, capHint)

		g := NewGraphWrapper(ctx.Graph())
		goal := g.Add(s.Goal)

		for in.Next() {
			select {
			case <-ctx.Context().Done():
				return
			default:
			}

			start := g.Add(in.Value())

			dstar := dynamic.NewDStarLite(start, goal, g, func(x, y graph.Node) float64 {
				return float64(s.Cost(ctx, x.(*nodeWrapper).Node, y.(*nodeWrapper).Node))
			}, simple.NewWeightedDirectedGraph(0, math.Inf(1)))

			for dstar.Step() {
				select {
				case <-ctx.Context().Done():
					return
				default:
				}

				node := dstar.Here().(*nodeWrapper).Node

				if s.Filter != nil && !s.Filter(ctx, node) {
					continue
				}

				hit := Hit{
					Node:  node,
					Score: s.Score(ctx, node),
				}

				idx, _ := slices.BinarySearchFunc(result, hit, func(i, j Hit) int {
					if i.Score > j.Score {
						return 1
					}

					if i.Score < j.Score {
						return -1
					}

					return 0
				})

				if idx >= cap(result) {
					continue
				}

				result = slices.Insert(result, idx, hit)

				if s.Limit != -1 && len(result) > s.Limit {
					result = result[:s.Limit]
				}
			}
		}

		if in.Err() != nil {
			ch <- IteratorItem{Err: in.Err()}
			return
		}

		for _, n := range result {
			ch <- IteratorItem{Value: n.Node}
		}
	}()

	return NewChanIterator(ch), nil
}
