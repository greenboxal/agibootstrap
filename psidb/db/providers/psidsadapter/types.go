package psidsadapter

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/db/psids"
	"github.com/greenboxal/agibootstrap/psidb/db/graphfs"
)

var dsKeyBitmap = psids.Key[BitmapSnapshot]("_graph/bitmap")
var dsKeyNodeData = psids.KeyTemplate[*graphfs.SerializedNode]("node-state/%d")
var dsKeyNodeEdge = psids.KeyTemplate[*graphfs.SerializedEdge]("node-edge/%d!/%s")
var dsKeyEdgePrefix = psids.KeyTemplate[*graphfs.SerializedEdge]("node-edge/%d!")

var dsKeyNodeDataMvcc = psids.KeyTemplate[*graphfs.SerializedNode]("node-state/%d\000%d")
var dsKeyNodeEdgeMvcc = psids.KeyTemplate[*graphfs.SerializedEdge]("node-edge/%d!/%s\000%d")
