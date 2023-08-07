package indexing

import "github.com/greenboxal/agibootstrap/pkg/platform/db/psids"

var dsKeyIndexItem = psids.KeyTemplate[IndexedItem]("index/%s/%d")
var dsKeyIndexItemPrefix = psids.KeyTemplate[IndexedItem]("index/%s")
var dsKeyLastID = psids.KeyTemplate[int64]("index/_last")

var dsKeyInvertedIndex = psids.KeyTemplate[uint64]("inverted/%s")
