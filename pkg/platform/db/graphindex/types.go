package graphindex

import "github.com/greenboxal/agibootstrap/pkg/platform/db/psids"

var dsKeyIndexItem = psids.KeyTemplate[IndexedItem]("index/%s/%d")
var dsKeyIndexItemPrefix = psids.KeyTemplate[IndexedItem]("index/%s")
