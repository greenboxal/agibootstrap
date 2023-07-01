package graphfx

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
)

type GraphBase[K comparable, N Node, E WeightedEdge] struct {
	helper GraphExpressionHelper[K, N, E]
}

func (o *GraphBase[K, N, E]) FireListeners(ev GraphChangeEvent[K, N, E]) {
	if o.helper != nil {
		o.helper.OnGraphChanged(ev)
	}
}

func (o *GraphBase[K, N, E]) AddListener(listener obsfx.InvalidationListener) {
	if o.helper == nil {
		o.helper = &singleInvalidationGraphExpressionHelper[K, N, E]{listener: listener}
	} else {
		o.helper = o.helper.AddListener(listener)
	}
}

func (o *GraphBase[K, N, E]) RemoveListener(listener obsfx.InvalidationListener) {
	if o.helper != nil {
		o.helper = o.helper.RemoveListener(listener)
	}
}

func (o *GraphBase[K, N, E]) AddGraphListener(listener GraphListener[K, N, E]) {
	if o.helper == nil {
		o.helper = &singleGraphListenerExpressionHelper[K, N, E]{listener: listener}
	} else {
		o.helper = o.helper.AddGraphListener(listener)
	}
}

func (o *GraphBase[K, N, E]) RemoveGraphListener(listener GraphListener[K, N, E]) {
	if o.helper != nil {
		o.helper = o.helper.RemoveGraphListener(listener)
	}
}

//type mutableGraph[K comparable, N Value, E WeightedEdge] struct {
//	idMap collectionsfx.MutableMap[K, N]
//	nodes collectionsfx.MutableMap[int64, N]
//	from  collectionsfx.MutableMap[int64, collectionsfx.ObservableMap[int64, E]]
//	to    collectionsfx.MutableMap[int64, collectionsfx.ObservableMap[int64, E]]
//
//	self, absent float64
//
//	bitmap *allocator.AllocationBitmap
//	used   *roaring.Bitmap
//
//	helper GraphExpressionHelper[K, N, E]
//}
//
//func NewMutableWeightedDirected[K comparable, N Value, E WeightedEdge](self, absent float64) MutableWeightedDirectedGraph[K, N, E] {
//	mg := &mutableGraph[K, N, E]{
//		self:   self,
//		absent: absent,
//		bitmap: allocator.NewAllocationMap(0x7FFFFFFF, ""),
//	}
//
//	return mg
//}
//
//func (o *mutableGraph[K, N, E]) NewNode() N {
//	id, ok, err := o.bitmap.AllocateNext()
//
//	if err != nil {
//		panic(err)
//	}
//
//	if !ok {
//		panic("id allocation failed")
//	}
//
//	return simple.Value(int64(id))
//}
//
//func (o *mutableGraph[K, N, E]) AddNode(node N) {
//
//	o.nodes.Set(node.ID(), node)
//}
//
//func (o *mutableGraph[K, N, E]) Value(v int64) N {
//	n, _ := o.nodes.Get(v)
//
//	return n
//}
//
//func (o *mutableGraph[K, N, E]) Nodes() graph.Nodes {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (o *mutableGraph[K, N, E]) From(id int64) graph.Nodes {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (o *mutableGraph[K, N, E]) To(id int64) graph.Nodes {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (o *mutableGraph[K, N, E]) NewEdge(from, to N) E {
//	panic("implement me")
//}
//
//func (o *mutableGraph[K, N, E]) SetEdge(e E) {
//	panic("implement me")
//}
//
//func (o *mutableGraph[K, N, E]) Edge(uid, vid int64) E {
//	return o.WeightedEdge(uid, vid)
//}
//
//func (o *mutableGraph[K, N, E]) NewWeightedEdge(from, to N, weight float64) E {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (o *mutableGraph[K, N, E]) SetWeightedEdge(e E) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (o *mutableGraph[K, N, E]) WeightedEdge(uid, vid int64) (res E) {
//	edges, _ := o.from.Get(uid)
//
//	if edges == nil {
//		return res
//	}
//
//	edge, _ := edges.Get(vid)
//
//	return edge
//}
//
//func (o *mutableGraph[K, N, E]) Weight(xid, yid int64) (w float64, ok bool) {
//	edge := o.WeightedEdge(xid, yid)
//
//	if edge == nil {
//		return o.absent, false
//	}
//
//	return edge.Weight(), true
//}
//
//func (o *mutableGraph[K, N, E]) HasEdgeBetween(xid, yid int64) bool {
//	return o.Edge(xid, yid) != nil || o.Edge(yid, xid) != nil
//}
//
//func (o *mutableGraph[K, N, E]) HasEdgeFromTo(uid, vid int64) bool {
//	return o.Edge(uid, vid) != nil
//}
//
