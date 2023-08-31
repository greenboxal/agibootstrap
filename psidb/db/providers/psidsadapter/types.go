package psidsadapter

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/db/psids"
	`github.com/greenboxal/agibootstrap/psidb/core/api`
)

var dsKeyBitmap = psids.Key[BitmapSnapshot]("_graph/bitmap")
var dsKeyNodeData = psids.KeyTemplate[*coreapi.SerializedNode]("node-state/%d")
var dsKeyNodeEdge = psids.KeyTemplate[*coreapi.SerializedEdge]("node-edge/%d!/%s")
var dsKeyEdgePrefix = psids.KeyTemplate[*coreapi.SerializedEdge]("node-edge/%d!")

var dsKeyNodeDataMvcc = psids.KeyTemplate[*coreapi.SerializedNode]("node-state/%d\000%d")
var dsKeyNodeEdgeMvcc = psids.KeyTemplate[*coreapi.SerializedEdge]("node-edge/%d!/%s\000%d")
