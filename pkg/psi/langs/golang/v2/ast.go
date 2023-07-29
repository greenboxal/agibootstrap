package golang

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Expr interface {
	Node
}

func NewFromExpr(fset *token.FileSet, node ast.Expr) Expr {
	return GoAstToPsi(fset, node).(Expr)
}

type Stmt interface {
	Node
}

func NewFromStmt(fset *token.FileSet, node ast.Stmt) Stmt {
	return GoAstToPsi(fset, node).(Stmt)
}

type Decl interface {
	Node
}

func NewFromDecl(fset *token.FileSet, node ast.Decl) Decl {
	return GoAstToPsi(fset, node).(Decl)
}

type Comment struct {
	NodeBase
	Text string `json:"Text"`
}

var CommentType = psi.DefineNodeType[*Comment]()

func NewFromComment(fset *token.FileSet, node *ast.Comment) *Comment {
	n := &Comment{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(CommentType))

	return n
}

func (n *Comment) CopyFromGoAst(fset *token.FileSet, src *ast.Comment) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	n.Text = src.Text
}

func (n *Comment) ToGoAst() ast.Node { return n.ToGoComment(nil) }

func (n *Comment) ToGoComment(dst *ast.Comment) *ast.Comment {
	if dst == nil {
		dst = &ast.Comment{}
	}
	dst.Text = n.Text

	return dst
}

type CommentGroup struct {
	NodeBase
}

var CommentGroupType = psi.DefineNodeType[*CommentGroup]()

var EdgeKindCommentGroupList = psi.DefineEdgeType[*Comment]("GoCommentGroupList")

func (n *CommentGroup) GetList() []*Comment      { return psi.GetEdges(n, EdgeKindCommentGroupList) }
func (n *CommentGroup) SetList(nodes []*Comment) { psi.UpdateEdges(n, EdgeKindCommentGroupList, nodes) }
func NewFromCommentGroup(fset *token.FileSet, node *ast.CommentGroup) *CommentGroup {
	n := &CommentGroup{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(CommentGroupType))

	return n
}

func (n *CommentGroup) CopyFromGoAst(fset *token.FileSet, src *ast.CommentGroup) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	for i, v := range src.List {
		tmpList := NewFromComment(fset, v)
		tmpList.SetParent(n)
		psi.UpdateEdge(n, EdgeKindCommentGroupList.Indexed(int64(i)), tmpList)
	}

}

func (n *CommentGroup) ToGoAst() ast.Node { return n.ToGoCommentGroup(nil) }

func (n *CommentGroup) ToGoCommentGroup(dst *ast.CommentGroup) *ast.CommentGroup {
	if dst == nil {
		dst = &ast.CommentGroup{}
	}
	tmpList := psi.GetEdges(n, EdgeKindCommentGroupList)
	dst.List = make([]*ast.Comment, len(tmpList))
	for i, v := range tmpList {
		dst.List[i] = v.ToGoAst().(*ast.Comment)
	}

	return dst
}

type Field struct {
	NodeBase
}

var FieldType = psi.DefineNodeType[*Field]()

var EdgeKindFieldDoc = psi.DefineEdgeType[*CommentGroup]("GoFieldDoc")
var EdgeKindFieldNames = psi.DefineEdgeType[*Ident]("GoFieldNames")
var EdgeKindFieldType = psi.DefineEdgeType[Expr]("GoFieldType")
var EdgeKindFieldTag = psi.DefineEdgeType[*BasicLit]("GoFieldTag")
var EdgeKindFieldComment = psi.DefineEdgeType[*CommentGroup]("GoFieldComment")

func (n *Field) GetDoc() *CommentGroup {
	return psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindFieldDoc.Singleton())
}
func (n *Field) SetDoc(node *CommentGroup) { psi.UpdateEdge(n, EdgeKindFieldDoc.Singleton(), node) }
func (n *Field) GetNames() []*Ident        { return psi.GetEdges(n, EdgeKindFieldNames) }
func (n *Field) SetNames(nodes []*Ident)   { psi.UpdateEdges(n, EdgeKindFieldNames, nodes) }
func (n *Field) GetType() Expr             { return psi.GetEdgeOrNil[Expr](n, EdgeKindFieldType.Singleton()) }
func (n *Field) SetType(node Expr)         { psi.UpdateEdge(n, EdgeKindFieldType.Singleton(), node) }
func (n *Field) GetTag() *BasicLit {
	return psi.GetEdgeOrNil[*BasicLit](n, EdgeKindFieldTag.Singleton())
}
func (n *Field) SetTag(node *BasicLit) { psi.UpdateEdge(n, EdgeKindFieldTag.Singleton(), node) }
func (n *Field) GetComment() *CommentGroup {
	return psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindFieldComment.Singleton())
}
func (n *Field) SetComment(node *CommentGroup) {
	psi.UpdateEdge(n, EdgeKindFieldComment.Singleton(), node)
}
func NewFromField(fset *token.FileSet, node *ast.Field) *Field {
	n := &Field{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(FieldType))

	return n
}

func (n *Field) CopyFromGoAst(fset *token.FileSet, src *ast.Field) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Doc != nil {
		tmpDoc := NewFromCommentGroup(fset, src.Doc)
		tmpDoc.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFieldDoc.Singleton(), tmpDoc)
	}

	for i, v := range src.Names {
		tmpNames := NewFromIdent(fset, v)
		tmpNames.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFieldNames.Indexed(int64(i)), tmpNames)
	}

	if src.Type != nil {
		tmpType := NewFromExpr(fset, src.Type)
		tmpType.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFieldType.Singleton(), tmpType)
	}

	if src.Tag != nil {
		tmpTag := NewFromBasicLit(fset, src.Tag)
		tmpTag.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFieldTag.Singleton(), tmpTag)
	}

	if src.Comment != nil {
		tmpComment := NewFromCommentGroup(fset, src.Comment)
		tmpComment.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFieldComment.Singleton(), tmpComment)
	}

}

func (n *Field) ToGoAst() ast.Node { return n.ToGoField(nil) }

func (n *Field) ToGoField(dst *ast.Field) *ast.Field {
	if dst == nil {
		dst = &ast.Field{}
	}
	tmpDoc := psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindFieldDoc.Singleton())
	if tmpDoc != nil {
		dst.Doc = tmpDoc.ToGoAst().(*ast.CommentGroup)
	}

	tmpNames := psi.GetEdges(n, EdgeKindFieldNames)
	dst.Names = make([]*ast.Ident, len(tmpNames))
	for i, v := range tmpNames {
		dst.Names[i] = v.ToGoAst().(*ast.Ident)
	}

	tmpType := psi.GetEdgeOrNil[Expr](n, EdgeKindFieldType.Singleton())
	if tmpType != nil {
		dst.Type = tmpType.ToGoAst().(ast.Expr)
	}

	tmpTag := psi.GetEdgeOrNil[*BasicLit](n, EdgeKindFieldTag.Singleton())
	if tmpTag != nil {
		dst.Tag = tmpTag.ToGoAst().(*ast.BasicLit)
	}

	tmpComment := psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindFieldComment.Singleton())
	if tmpComment != nil {
		dst.Comment = tmpComment.ToGoAst().(*ast.CommentGroup)
	}

	return dst
}

type FieldList struct {
	NodeBase
}

var FieldListType = psi.DefineNodeType[*FieldList]()

var EdgeKindFieldListList = psi.DefineEdgeType[*Field]("GoFieldListList")

func (n *FieldList) GetList() []*Field      { return psi.GetEdges(n, EdgeKindFieldListList) }
func (n *FieldList) SetList(nodes []*Field) { psi.UpdateEdges(n, EdgeKindFieldListList, nodes) }
func NewFromFieldList(fset *token.FileSet, node *ast.FieldList) *FieldList {
	n := &FieldList{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(FieldListType))

	return n
}

func (n *FieldList) CopyFromGoAst(fset *token.FileSet, src *ast.FieldList) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	for i, v := range src.List {
		tmpList := NewFromField(fset, v)
		tmpList.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFieldListList.Indexed(int64(i)), tmpList)
	}

}

func (n *FieldList) ToGoAst() ast.Node { return n.ToGoFieldList(nil) }

func (n *FieldList) ToGoFieldList(dst *ast.FieldList) *ast.FieldList {
	if dst == nil {
		dst = &ast.FieldList{}
	}
	tmpList := psi.GetEdges(n, EdgeKindFieldListList)
	dst.List = make([]*ast.Field, len(tmpList))
	for i, v := range tmpList {
		dst.List[i] = v.ToGoAst().(*ast.Field)
	}

	return dst
}

type BadExpr struct {
	NodeBase
}

var BadExprType = psi.DefineNodeType[*BadExpr]()

func NewFromBadExpr(fset *token.FileSet, node *ast.BadExpr) *BadExpr {
	n := &BadExpr{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(BadExprType))

	return n
}

func (n *BadExpr) CopyFromGoAst(fset *token.FileSet, src *ast.BadExpr) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
}

func (n *BadExpr) ToGoAst() ast.Node { return n.ToGoBadExpr(nil) }

func (n *BadExpr) ToGoBadExpr(dst *ast.BadExpr) *ast.BadExpr {
	if dst == nil {
		dst = &ast.BadExpr{}
	}

	return dst
}

type Ident struct {
	NodeBase
	Name string `json:"Name"`
}

var IdentType = psi.DefineNodeType[*Ident]()

func NewFromIdent(fset *token.FileSet, node *ast.Ident) *Ident {
	n := &Ident{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(IdentType))

	return n
}

func (n *Ident) CopyFromGoAst(fset *token.FileSet, src *ast.Ident) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	n.Name = src.Name
}

func (n *Ident) ToGoAst() ast.Node { return n.ToGoIdent(nil) }

func (n *Ident) ToGoIdent(dst *ast.Ident) *ast.Ident {
	if dst == nil {
		dst = &ast.Ident{}
	}
	dst.Name = n.Name

	return dst
}

type Ellipsis struct {
	NodeBase
}

var EllipsisType = psi.DefineNodeType[*Ellipsis]()

var EdgeKindEllipsisElt = psi.DefineEdgeType[Expr]("GoEllipsisElt")

func (n *Ellipsis) GetElt() Expr     { return psi.GetEdgeOrNil[Expr](n, EdgeKindEllipsisElt.Singleton()) }
func (n *Ellipsis) SetElt(node Expr) { psi.UpdateEdge(n, EdgeKindEllipsisElt.Singleton(), node) }
func NewFromEllipsis(fset *token.FileSet, node *ast.Ellipsis) *Ellipsis {
	n := &Ellipsis{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(EllipsisType))

	return n
}

func (n *Ellipsis) CopyFromGoAst(fset *token.FileSet, src *ast.Ellipsis) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Elt != nil {
		tmpElt := NewFromExpr(fset, src.Elt)
		tmpElt.SetParent(n)
		psi.UpdateEdge(n, EdgeKindEllipsisElt.Singleton(), tmpElt)
	}

}

func (n *Ellipsis) ToGoAst() ast.Node { return n.ToGoEllipsis(nil) }

func (n *Ellipsis) ToGoEllipsis(dst *ast.Ellipsis) *ast.Ellipsis {
	if dst == nil {
		dst = &ast.Ellipsis{}
	}
	tmpElt := psi.GetEdgeOrNil[Expr](n, EdgeKindEllipsisElt.Singleton())
	if tmpElt != nil {
		dst.Elt = tmpElt.ToGoAst().(ast.Expr)
	}

	return dst
}

type BasicLit struct {
	NodeBase
	Kind  token.Token `json:"Kind"`
	Value string      `json:"Value"`
}

var BasicLitType = psi.DefineNodeType[*BasicLit]()

func NewFromBasicLit(fset *token.FileSet, node *ast.BasicLit) *BasicLit {
	n := &BasicLit{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(BasicLitType))

	return n
}

func (n *BasicLit) CopyFromGoAst(fset *token.FileSet, src *ast.BasicLit) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	n.Kind = src.Kind
	n.Value = src.Value
}

func (n *BasicLit) ToGoAst() ast.Node { return n.ToGoBasicLit(nil) }

func (n *BasicLit) ToGoBasicLit(dst *ast.BasicLit) *ast.BasicLit {
	if dst == nil {
		dst = &ast.BasicLit{}
	}
	dst.Kind = n.Kind
	dst.Value = n.Value

	return dst
}

type FuncLit struct {
	NodeBase
}

var FuncLitType = psi.DefineNodeType[*FuncLit]()

var EdgeKindFuncLitType = psi.DefineEdgeType[*FuncType]("GoFuncLitType")
var EdgeKindFuncLitBody = psi.DefineEdgeType[*BlockStmt]("GoFuncLitBody")

func (n *FuncLit) GetType() *FuncType {
	return psi.GetEdgeOrNil[*FuncType](n, EdgeKindFuncLitType.Singleton())
}
func (n *FuncLit) SetType(node *FuncType) { psi.UpdateEdge(n, EdgeKindFuncLitType.Singleton(), node) }
func (n *FuncLit) GetBody() *BlockStmt {
	return psi.GetEdgeOrNil[*BlockStmt](n, EdgeKindFuncLitBody.Singleton())
}
func (n *FuncLit) SetBody(node *BlockStmt) { psi.UpdateEdge(n, EdgeKindFuncLitBody.Singleton(), node) }
func NewFromFuncLit(fset *token.FileSet, node *ast.FuncLit) *FuncLit {
	n := &FuncLit{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(FuncLitType))

	return n
}

func (n *FuncLit) CopyFromGoAst(fset *token.FileSet, src *ast.FuncLit) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Type != nil {
		tmpType := NewFromFuncType(fset, src.Type)
		tmpType.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFuncLitType.Singleton(), tmpType)
	}

	if src.Body != nil {
		tmpBody := NewFromBlockStmt(fset, src.Body)
		tmpBody.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFuncLitBody.Singleton(), tmpBody)
	}

}

func (n *FuncLit) ToGoAst() ast.Node { return n.ToGoFuncLit(nil) }

func (n *FuncLit) ToGoFuncLit(dst *ast.FuncLit) *ast.FuncLit {
	if dst == nil {
		dst = &ast.FuncLit{}
	}
	tmpType := psi.GetEdgeOrNil[*FuncType](n, EdgeKindFuncLitType.Singleton())
	if tmpType != nil {
		dst.Type = tmpType.ToGoAst().(*ast.FuncType)
	}

	tmpBody := psi.GetEdgeOrNil[*BlockStmt](n, EdgeKindFuncLitBody.Singleton())
	if tmpBody != nil {
		dst.Body = tmpBody.ToGoAst().(*ast.BlockStmt)
	}

	return dst
}

type CompositeLit struct {
	NodeBase
	Incomplete bool `json:"Incomplete"`
}

var CompositeLitType = psi.DefineNodeType[*CompositeLit]()

var EdgeKindCompositeLitType = psi.DefineEdgeType[Expr]("GoCompositeLitType")
var EdgeKindCompositeLitElts = psi.DefineEdgeType[Expr]("GoCompositeLitElts")

func (n *CompositeLit) GetType() Expr {
	return psi.GetEdgeOrNil[Expr](n, EdgeKindCompositeLitType.Singleton())
}
func (n *CompositeLit) SetType(node Expr) {
	psi.UpdateEdge(n, EdgeKindCompositeLitType.Singleton(), node)
}
func (n *CompositeLit) GetElts() []Expr      { return psi.GetEdges(n, EdgeKindCompositeLitElts) }
func (n *CompositeLit) SetElts(nodes []Expr) { psi.UpdateEdges(n, EdgeKindCompositeLitElts, nodes) }
func NewFromCompositeLit(fset *token.FileSet, node *ast.CompositeLit) *CompositeLit {
	n := &CompositeLit{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(CompositeLitType))

	return n
}

func (n *CompositeLit) CopyFromGoAst(fset *token.FileSet, src *ast.CompositeLit) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	n.Incomplete = src.Incomplete
	if src.Type != nil {
		tmpType := NewFromExpr(fset, src.Type)
		tmpType.SetParent(n)
		psi.UpdateEdge(n, EdgeKindCompositeLitType.Singleton(), tmpType)
	}

	for i, v := range src.Elts {
		tmpElts := NewFromExpr(fset, v)
		tmpElts.SetParent(n)
		psi.UpdateEdge(n, EdgeKindCompositeLitElts.Indexed(int64(i)), tmpElts)
	}

}

func (n *CompositeLit) ToGoAst() ast.Node { return n.ToGoCompositeLit(nil) }

func (n *CompositeLit) ToGoCompositeLit(dst *ast.CompositeLit) *ast.CompositeLit {
	if dst == nil {
		dst = &ast.CompositeLit{}
	}
	dst.Incomplete = n.Incomplete
	tmpType := psi.GetEdgeOrNil[Expr](n, EdgeKindCompositeLitType.Singleton())
	if tmpType != nil {
		dst.Type = tmpType.ToGoAst().(ast.Expr)
	}

	tmpElts := psi.GetEdges(n, EdgeKindCompositeLitElts)
	dst.Elts = make([]ast.Expr, len(tmpElts))
	for i, v := range tmpElts {
		dst.Elts[i] = v.ToGoAst().(ast.Expr)
	}

	return dst
}

type ParenExpr struct {
	NodeBase
}

var ParenExprType = psi.DefineNodeType[*ParenExpr]()

var EdgeKindParenExprX = psi.DefineEdgeType[Expr]("GoParenExprX")

func (n *ParenExpr) GetX() Expr     { return psi.GetEdgeOrNil[Expr](n, EdgeKindParenExprX.Singleton()) }
func (n *ParenExpr) SetX(node Expr) { psi.UpdateEdge(n, EdgeKindParenExprX.Singleton(), node) }
func NewFromParenExpr(fset *token.FileSet, node *ast.ParenExpr) *ParenExpr {
	n := &ParenExpr{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(ParenExprType))

	return n
}

func (n *ParenExpr) CopyFromGoAst(fset *token.FileSet, src *ast.ParenExpr) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.X != nil {
		tmpX := NewFromExpr(fset, src.X)
		tmpX.SetParent(n)
		psi.UpdateEdge(n, EdgeKindParenExprX.Singleton(), tmpX)
	}

}

func (n *ParenExpr) ToGoAst() ast.Node { return n.ToGoParenExpr(nil) }

func (n *ParenExpr) ToGoParenExpr(dst *ast.ParenExpr) *ast.ParenExpr {
	if dst == nil {
		dst = &ast.ParenExpr{}
	}
	tmpX := psi.GetEdgeOrNil[Expr](n, EdgeKindParenExprX.Singleton())
	if tmpX != nil {
		dst.X = tmpX.ToGoAst().(ast.Expr)
	}

	return dst
}

type SelectorExpr struct {
	NodeBase
}

var SelectorExprType = psi.DefineNodeType[*SelectorExpr]()

var EdgeKindSelectorExprX = psi.DefineEdgeType[Expr]("GoSelectorExprX")
var EdgeKindSelectorExprSel = psi.DefineEdgeType[*Ident]("GoSelectorExprSel")

func (n *SelectorExpr) GetX() Expr {
	return psi.GetEdgeOrNil[Expr](n, EdgeKindSelectorExprX.Singleton())
}
func (n *SelectorExpr) SetX(node Expr) { psi.UpdateEdge(n, EdgeKindSelectorExprX.Singleton(), node) }
func (n *SelectorExpr) GetSel() *Ident {
	return psi.GetEdgeOrNil[*Ident](n, EdgeKindSelectorExprSel.Singleton())
}
func (n *SelectorExpr) SetSel(node *Ident) {
	psi.UpdateEdge(n, EdgeKindSelectorExprSel.Singleton(), node)
}
func NewFromSelectorExpr(fset *token.FileSet, node *ast.SelectorExpr) *SelectorExpr {
	n := &SelectorExpr{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(SelectorExprType))

	return n
}

func (n *SelectorExpr) CopyFromGoAst(fset *token.FileSet, src *ast.SelectorExpr) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.X != nil {
		tmpX := NewFromExpr(fset, src.X)
		tmpX.SetParent(n)
		psi.UpdateEdge(n, EdgeKindSelectorExprX.Singleton(), tmpX)
	}

	if src.Sel != nil {
		tmpSel := NewFromIdent(fset, src.Sel)
		tmpSel.SetParent(n)
		psi.UpdateEdge(n, EdgeKindSelectorExprSel.Singleton(), tmpSel)
	}

}

func (n *SelectorExpr) ToGoAst() ast.Node { return n.ToGoSelectorExpr(nil) }

func (n *SelectorExpr) ToGoSelectorExpr(dst *ast.SelectorExpr) *ast.SelectorExpr {
	if dst == nil {
		dst = &ast.SelectorExpr{}
	}
	tmpX := psi.GetEdgeOrNil[Expr](n, EdgeKindSelectorExprX.Singleton())
	if tmpX != nil {
		dst.X = tmpX.ToGoAst().(ast.Expr)
	}

	tmpSel := psi.GetEdgeOrNil[*Ident](n, EdgeKindSelectorExprSel.Singleton())
	if tmpSel != nil {
		dst.Sel = tmpSel.ToGoAst().(*ast.Ident)
	}

	return dst
}

type IndexExpr struct {
	NodeBase
}

var IndexExprType = psi.DefineNodeType[*IndexExpr]()

var EdgeKindIndexExprX = psi.DefineEdgeType[Expr]("GoIndexExprX")
var EdgeKindIndexExprIndex = psi.DefineEdgeType[Expr]("GoIndexExprIndex")

func (n *IndexExpr) GetX() Expr     { return psi.GetEdgeOrNil[Expr](n, EdgeKindIndexExprX.Singleton()) }
func (n *IndexExpr) SetX(node Expr) { psi.UpdateEdge(n, EdgeKindIndexExprX.Singleton(), node) }
func (n *IndexExpr) GetIndex() Expr {
	return psi.GetEdgeOrNil[Expr](n, EdgeKindIndexExprIndex.Singleton())
}
func (n *IndexExpr) SetIndex(node Expr) { psi.UpdateEdge(n, EdgeKindIndexExprIndex.Singleton(), node) }
func NewFromIndexExpr(fset *token.FileSet, node *ast.IndexExpr) *IndexExpr {
	n := &IndexExpr{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(IndexExprType))

	return n
}

func (n *IndexExpr) CopyFromGoAst(fset *token.FileSet, src *ast.IndexExpr) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.X != nil {
		tmpX := NewFromExpr(fset, src.X)
		tmpX.SetParent(n)
		psi.UpdateEdge(n, EdgeKindIndexExprX.Singleton(), tmpX)
	}

	if src.Index != nil {
		tmpIndex := NewFromExpr(fset, src.Index)
		tmpIndex.SetParent(n)
		psi.UpdateEdge(n, EdgeKindIndexExprIndex.Singleton(), tmpIndex)
	}

}

func (n *IndexExpr) ToGoAst() ast.Node { return n.ToGoIndexExpr(nil) }

func (n *IndexExpr) ToGoIndexExpr(dst *ast.IndexExpr) *ast.IndexExpr {
	if dst == nil {
		dst = &ast.IndexExpr{}
	}
	tmpX := psi.GetEdgeOrNil[Expr](n, EdgeKindIndexExprX.Singleton())
	if tmpX != nil {
		dst.X = tmpX.ToGoAst().(ast.Expr)
	}

	tmpIndex := psi.GetEdgeOrNil[Expr](n, EdgeKindIndexExprIndex.Singleton())
	if tmpIndex != nil {
		dst.Index = tmpIndex.ToGoAst().(ast.Expr)
	}

	return dst
}

type IndexListExpr struct {
	NodeBase
}

var IndexListExprType = psi.DefineNodeType[*IndexListExpr]()

var EdgeKindIndexListExprX = psi.DefineEdgeType[Expr]("GoIndexListExprX")
var EdgeKindIndexListExprIndices = psi.DefineEdgeType[Expr]("GoIndexListExprIndices")

func (n *IndexListExpr) GetX() Expr {
	return psi.GetEdgeOrNil[Expr](n, EdgeKindIndexListExprX.Singleton())
}
func (n *IndexListExpr) SetX(node Expr)     { psi.UpdateEdge(n, EdgeKindIndexListExprX.Singleton(), node) }
func (n *IndexListExpr) GetIndices() []Expr { return psi.GetEdges(n, EdgeKindIndexListExprIndices) }
func (n *IndexListExpr) SetIndices(nodes []Expr) {
	psi.UpdateEdges(n, EdgeKindIndexListExprIndices, nodes)
}
func NewFromIndexListExpr(fset *token.FileSet, node *ast.IndexListExpr) *IndexListExpr {
	n := &IndexListExpr{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(IndexListExprType))

	return n
}

func (n *IndexListExpr) CopyFromGoAst(fset *token.FileSet, src *ast.IndexListExpr) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.X != nil {
		tmpX := NewFromExpr(fset, src.X)
		tmpX.SetParent(n)
		psi.UpdateEdge(n, EdgeKindIndexListExprX.Singleton(), tmpX)
	}

	for i, v := range src.Indices {
		tmpIndices := NewFromExpr(fset, v)
		tmpIndices.SetParent(n)
		psi.UpdateEdge(n, EdgeKindIndexListExprIndices.Indexed(int64(i)), tmpIndices)
	}

}

func (n *IndexListExpr) ToGoAst() ast.Node { return n.ToGoIndexListExpr(nil) }

func (n *IndexListExpr) ToGoIndexListExpr(dst *ast.IndexListExpr) *ast.IndexListExpr {
	if dst == nil {
		dst = &ast.IndexListExpr{}
	}
	tmpX := psi.GetEdgeOrNil[Expr](n, EdgeKindIndexListExprX.Singleton())
	if tmpX != nil {
		dst.X = tmpX.ToGoAst().(ast.Expr)
	}

	tmpIndices := psi.GetEdges(n, EdgeKindIndexListExprIndices)
	dst.Indices = make([]ast.Expr, len(tmpIndices))
	for i, v := range tmpIndices {
		dst.Indices[i] = v.ToGoAst().(ast.Expr)
	}

	return dst
}

type SliceExpr struct {
	NodeBase
	Slice3 bool `json:"Slice3"`
}

var SliceExprType = psi.DefineNodeType[*SliceExpr]()

var EdgeKindSliceExprX = psi.DefineEdgeType[Expr]("GoSliceExprX")
var EdgeKindSliceExprLow = psi.DefineEdgeType[Expr]("GoSliceExprLow")
var EdgeKindSliceExprHigh = psi.DefineEdgeType[Expr]("GoSliceExprHigh")
var EdgeKindSliceExprMax = psi.DefineEdgeType[Expr]("GoSliceExprMax")

func (n *SliceExpr) GetX() Expr       { return psi.GetEdgeOrNil[Expr](n, EdgeKindSliceExprX.Singleton()) }
func (n *SliceExpr) SetX(node Expr)   { psi.UpdateEdge(n, EdgeKindSliceExprX.Singleton(), node) }
func (n *SliceExpr) GetLow() Expr     { return psi.GetEdgeOrNil[Expr](n, EdgeKindSliceExprLow.Singleton()) }
func (n *SliceExpr) SetLow(node Expr) { psi.UpdateEdge(n, EdgeKindSliceExprLow.Singleton(), node) }
func (n *SliceExpr) GetHigh() Expr {
	return psi.GetEdgeOrNil[Expr](n, EdgeKindSliceExprHigh.Singleton())
}
func (n *SliceExpr) SetHigh(node Expr) { psi.UpdateEdge(n, EdgeKindSliceExprHigh.Singleton(), node) }
func (n *SliceExpr) GetMax() Expr      { return psi.GetEdgeOrNil[Expr](n, EdgeKindSliceExprMax.Singleton()) }
func (n *SliceExpr) SetMax(node Expr)  { psi.UpdateEdge(n, EdgeKindSliceExprMax.Singleton(), node) }
func NewFromSliceExpr(fset *token.FileSet, node *ast.SliceExpr) *SliceExpr {
	n := &SliceExpr{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(SliceExprType))

	return n
}

func (n *SliceExpr) CopyFromGoAst(fset *token.FileSet, src *ast.SliceExpr) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	n.Slice3 = src.Slice3
	if src.X != nil {
		tmpX := NewFromExpr(fset, src.X)
		tmpX.SetParent(n)
		psi.UpdateEdge(n, EdgeKindSliceExprX.Singleton(), tmpX)
	}

	if src.Low != nil {
		tmpLow := NewFromExpr(fset, src.Low)
		tmpLow.SetParent(n)
		psi.UpdateEdge(n, EdgeKindSliceExprLow.Singleton(), tmpLow)
	}

	if src.High != nil {
		tmpHigh := NewFromExpr(fset, src.High)
		tmpHigh.SetParent(n)
		psi.UpdateEdge(n, EdgeKindSliceExprHigh.Singleton(), tmpHigh)
	}

	if src.Max != nil {
		tmpMax := NewFromExpr(fset, src.Max)
		tmpMax.SetParent(n)
		psi.UpdateEdge(n, EdgeKindSliceExprMax.Singleton(), tmpMax)
	}

}

func (n *SliceExpr) ToGoAst() ast.Node { return n.ToGoSliceExpr(nil) }

func (n *SliceExpr) ToGoSliceExpr(dst *ast.SliceExpr) *ast.SliceExpr {
	if dst == nil {
		dst = &ast.SliceExpr{}
	}
	dst.Slice3 = n.Slice3
	tmpX := psi.GetEdgeOrNil[Expr](n, EdgeKindSliceExprX.Singleton())
	if tmpX != nil {
		dst.X = tmpX.ToGoAst().(ast.Expr)
	}

	tmpLow := psi.GetEdgeOrNil[Expr](n, EdgeKindSliceExprLow.Singleton())
	if tmpLow != nil {
		dst.Low = tmpLow.ToGoAst().(ast.Expr)
	}

	tmpHigh := psi.GetEdgeOrNil[Expr](n, EdgeKindSliceExprHigh.Singleton())
	if tmpHigh != nil {
		dst.High = tmpHigh.ToGoAst().(ast.Expr)
	}

	tmpMax := psi.GetEdgeOrNil[Expr](n, EdgeKindSliceExprMax.Singleton())
	if tmpMax != nil {
		dst.Max = tmpMax.ToGoAst().(ast.Expr)
	}

	return dst
}

type TypeAssertExpr struct {
	NodeBase
}

var TypeAssertExprType = psi.DefineNodeType[*TypeAssertExpr]()

var EdgeKindTypeAssertExprX = psi.DefineEdgeType[Expr]("GoTypeAssertExprX")
var EdgeKindTypeAssertExprType = psi.DefineEdgeType[Expr]("GoTypeAssertExprType")

func (n *TypeAssertExpr) GetX() Expr {
	return psi.GetEdgeOrNil[Expr](n, EdgeKindTypeAssertExprX.Singleton())
}
func (n *TypeAssertExpr) SetX(node Expr) {
	psi.UpdateEdge(n, EdgeKindTypeAssertExprX.Singleton(), node)
}
func (n *TypeAssertExpr) GetType() Expr {
	return psi.GetEdgeOrNil[Expr](n, EdgeKindTypeAssertExprType.Singleton())
}
func (n *TypeAssertExpr) SetType(node Expr) {
	psi.UpdateEdge(n, EdgeKindTypeAssertExprType.Singleton(), node)
}
func NewFromTypeAssertExpr(fset *token.FileSet, node *ast.TypeAssertExpr) *TypeAssertExpr {
	n := &TypeAssertExpr{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(TypeAssertExprType))

	return n
}

func (n *TypeAssertExpr) CopyFromGoAst(fset *token.FileSet, src *ast.TypeAssertExpr) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.X != nil {
		tmpX := NewFromExpr(fset, src.X)
		tmpX.SetParent(n)
		psi.UpdateEdge(n, EdgeKindTypeAssertExprX.Singleton(), tmpX)
	}

	if src.Type != nil {
		tmpType := NewFromExpr(fset, src.Type)
		tmpType.SetParent(n)
		psi.UpdateEdge(n, EdgeKindTypeAssertExprType.Singleton(), tmpType)
	}

}

func (n *TypeAssertExpr) ToGoAst() ast.Node { return n.ToGoTypeAssertExpr(nil) }

func (n *TypeAssertExpr) ToGoTypeAssertExpr(dst *ast.TypeAssertExpr) *ast.TypeAssertExpr {
	if dst == nil {
		dst = &ast.TypeAssertExpr{}
	}
	tmpX := psi.GetEdgeOrNil[Expr](n, EdgeKindTypeAssertExprX.Singleton())
	if tmpX != nil {
		dst.X = tmpX.ToGoAst().(ast.Expr)
	}

	tmpType := psi.GetEdgeOrNil[Expr](n, EdgeKindTypeAssertExprType.Singleton())
	if tmpType != nil {
		dst.Type = tmpType.ToGoAst().(ast.Expr)
	}

	return dst
}

type CallExpr struct {
	NodeBase
}

var CallExprType = psi.DefineNodeType[*CallExpr]()

var EdgeKindCallExprFun = psi.DefineEdgeType[Expr]("GoCallExprFun")
var EdgeKindCallExprArgs = psi.DefineEdgeType[Expr]("GoCallExprArgs")

func (n *CallExpr) GetFun() Expr         { return psi.GetEdgeOrNil[Expr](n, EdgeKindCallExprFun.Singleton()) }
func (n *CallExpr) SetFun(node Expr)     { psi.UpdateEdge(n, EdgeKindCallExprFun.Singleton(), node) }
func (n *CallExpr) GetArgs() []Expr      { return psi.GetEdges(n, EdgeKindCallExprArgs) }
func (n *CallExpr) SetArgs(nodes []Expr) { psi.UpdateEdges(n, EdgeKindCallExprArgs, nodes) }
func NewFromCallExpr(fset *token.FileSet, node *ast.CallExpr) *CallExpr {
	n := &CallExpr{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(CallExprType))

	return n
}

func (n *CallExpr) CopyFromGoAst(fset *token.FileSet, src *ast.CallExpr) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Fun != nil {
		tmpFun := NewFromExpr(fset, src.Fun)
		tmpFun.SetParent(n)
		psi.UpdateEdge(n, EdgeKindCallExprFun.Singleton(), tmpFun)
	}

	for i, v := range src.Args {
		tmpArgs := NewFromExpr(fset, v)
		tmpArgs.SetParent(n)
		psi.UpdateEdge(n, EdgeKindCallExprArgs.Indexed(int64(i)), tmpArgs)
	}

}

func (n *CallExpr) ToGoAst() ast.Node { return n.ToGoCallExpr(nil) }

func (n *CallExpr) ToGoCallExpr(dst *ast.CallExpr) *ast.CallExpr {
	if dst == nil {
		dst = &ast.CallExpr{}
	}
	tmpFun := psi.GetEdgeOrNil[Expr](n, EdgeKindCallExprFun.Singleton())
	if tmpFun != nil {
		dst.Fun = tmpFun.ToGoAst().(ast.Expr)
	}

	tmpArgs := psi.GetEdges(n, EdgeKindCallExprArgs)
	dst.Args = make([]ast.Expr, len(tmpArgs))
	for i, v := range tmpArgs {
		dst.Args[i] = v.ToGoAst().(ast.Expr)
	}

	return dst
}

type StarExpr struct {
	NodeBase
}

var StarExprType = psi.DefineNodeType[*StarExpr]()

var EdgeKindStarExprX = psi.DefineEdgeType[Expr]("GoStarExprX")

func (n *StarExpr) GetX() Expr     { return psi.GetEdgeOrNil[Expr](n, EdgeKindStarExprX.Singleton()) }
func (n *StarExpr) SetX(node Expr) { psi.UpdateEdge(n, EdgeKindStarExprX.Singleton(), node) }
func NewFromStarExpr(fset *token.FileSet, node *ast.StarExpr) *StarExpr {
	n := &StarExpr{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(StarExprType))

	return n
}

func (n *StarExpr) CopyFromGoAst(fset *token.FileSet, src *ast.StarExpr) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.X != nil {
		tmpX := NewFromExpr(fset, src.X)
		tmpX.SetParent(n)
		psi.UpdateEdge(n, EdgeKindStarExprX.Singleton(), tmpX)
	}

}

func (n *StarExpr) ToGoAst() ast.Node { return n.ToGoStarExpr(nil) }

func (n *StarExpr) ToGoStarExpr(dst *ast.StarExpr) *ast.StarExpr {
	if dst == nil {
		dst = &ast.StarExpr{}
	}
	tmpX := psi.GetEdgeOrNil[Expr](n, EdgeKindStarExprX.Singleton())
	if tmpX != nil {
		dst.X = tmpX.ToGoAst().(ast.Expr)
	}

	return dst
}

type UnaryExpr struct {
	NodeBase
	Op token.Token `json:"Op"`
}

var UnaryExprType = psi.DefineNodeType[*UnaryExpr]()

var EdgeKindUnaryExprX = psi.DefineEdgeType[Expr]("GoUnaryExprX")

func (n *UnaryExpr) GetX() Expr     { return psi.GetEdgeOrNil[Expr](n, EdgeKindUnaryExprX.Singleton()) }
func (n *UnaryExpr) SetX(node Expr) { psi.UpdateEdge(n, EdgeKindUnaryExprX.Singleton(), node) }
func NewFromUnaryExpr(fset *token.FileSet, node *ast.UnaryExpr) *UnaryExpr {
	n := &UnaryExpr{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(UnaryExprType))

	return n
}

func (n *UnaryExpr) CopyFromGoAst(fset *token.FileSet, src *ast.UnaryExpr) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	n.Op = src.Op
	if src.X != nil {
		tmpX := NewFromExpr(fset, src.X)
		tmpX.SetParent(n)
		psi.UpdateEdge(n, EdgeKindUnaryExprX.Singleton(), tmpX)
	}

}

func (n *UnaryExpr) ToGoAst() ast.Node { return n.ToGoUnaryExpr(nil) }

func (n *UnaryExpr) ToGoUnaryExpr(dst *ast.UnaryExpr) *ast.UnaryExpr {
	if dst == nil {
		dst = &ast.UnaryExpr{}
	}
	dst.Op = n.Op
	tmpX := psi.GetEdgeOrNil[Expr](n, EdgeKindUnaryExprX.Singleton())
	if tmpX != nil {
		dst.X = tmpX.ToGoAst().(ast.Expr)
	}

	return dst
}

type BinaryExpr struct {
	NodeBase
	Op token.Token `json:"Op"`
}

var BinaryExprType = psi.DefineNodeType[*BinaryExpr]()

var EdgeKindBinaryExprX = psi.DefineEdgeType[Expr]("GoBinaryExprX")
var EdgeKindBinaryExprY = psi.DefineEdgeType[Expr]("GoBinaryExprY")

func (n *BinaryExpr) GetX() Expr     { return psi.GetEdgeOrNil[Expr](n, EdgeKindBinaryExprX.Singleton()) }
func (n *BinaryExpr) SetX(node Expr) { psi.UpdateEdge(n, EdgeKindBinaryExprX.Singleton(), node) }
func (n *BinaryExpr) GetY() Expr     { return psi.GetEdgeOrNil[Expr](n, EdgeKindBinaryExprY.Singleton()) }
func (n *BinaryExpr) SetY(node Expr) { psi.UpdateEdge(n, EdgeKindBinaryExprY.Singleton(), node) }
func NewFromBinaryExpr(fset *token.FileSet, node *ast.BinaryExpr) *BinaryExpr {
	n := &BinaryExpr{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(BinaryExprType))

	return n
}

func (n *BinaryExpr) CopyFromGoAst(fset *token.FileSet, src *ast.BinaryExpr) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	n.Op = src.Op
	if src.X != nil {
		tmpX := NewFromExpr(fset, src.X)
		tmpX.SetParent(n)
		psi.UpdateEdge(n, EdgeKindBinaryExprX.Singleton(), tmpX)
	}

	if src.Y != nil {
		tmpY := NewFromExpr(fset, src.Y)
		tmpY.SetParent(n)
		psi.UpdateEdge(n, EdgeKindBinaryExprY.Singleton(), tmpY)
	}

}

func (n *BinaryExpr) ToGoAst() ast.Node { return n.ToGoBinaryExpr(nil) }

func (n *BinaryExpr) ToGoBinaryExpr(dst *ast.BinaryExpr) *ast.BinaryExpr {
	if dst == nil {
		dst = &ast.BinaryExpr{}
	}
	dst.Op = n.Op
	tmpX := psi.GetEdgeOrNil[Expr](n, EdgeKindBinaryExprX.Singleton())
	if tmpX != nil {
		dst.X = tmpX.ToGoAst().(ast.Expr)
	}

	tmpY := psi.GetEdgeOrNil[Expr](n, EdgeKindBinaryExprY.Singleton())
	if tmpY != nil {
		dst.Y = tmpY.ToGoAst().(ast.Expr)
	}

	return dst
}

type KeyValueExpr struct {
	NodeBase
}

var KeyValueExprType = psi.DefineNodeType[*KeyValueExpr]()

var EdgeKindKeyValueExprKey = psi.DefineEdgeType[Expr]("GoKeyValueExprKey")
var EdgeKindKeyValueExprValue = psi.DefineEdgeType[Expr]("GoKeyValueExprValue")

func (n *KeyValueExpr) GetKey() Expr {
	return psi.GetEdgeOrNil[Expr](n, EdgeKindKeyValueExprKey.Singleton())
}
func (n *KeyValueExpr) SetKey(node Expr) {
	psi.UpdateEdge(n, EdgeKindKeyValueExprKey.Singleton(), node)
}
func (n *KeyValueExpr) GetValue() Expr {
	return psi.GetEdgeOrNil[Expr](n, EdgeKindKeyValueExprValue.Singleton())
}
func (n *KeyValueExpr) SetValue(node Expr) {
	psi.UpdateEdge(n, EdgeKindKeyValueExprValue.Singleton(), node)
}
func NewFromKeyValueExpr(fset *token.FileSet, node *ast.KeyValueExpr) *KeyValueExpr {
	n := &KeyValueExpr{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(KeyValueExprType))

	return n
}

func (n *KeyValueExpr) CopyFromGoAst(fset *token.FileSet, src *ast.KeyValueExpr) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Key != nil {
		tmpKey := NewFromExpr(fset, src.Key)
		tmpKey.SetParent(n)
		psi.UpdateEdge(n, EdgeKindKeyValueExprKey.Singleton(), tmpKey)
	}

	if src.Value != nil {
		tmpValue := NewFromExpr(fset, src.Value)
		tmpValue.SetParent(n)
		psi.UpdateEdge(n, EdgeKindKeyValueExprValue.Singleton(), tmpValue)
	}

}

func (n *KeyValueExpr) ToGoAst() ast.Node { return n.ToGoKeyValueExpr(nil) }

func (n *KeyValueExpr) ToGoKeyValueExpr(dst *ast.KeyValueExpr) *ast.KeyValueExpr {
	if dst == nil {
		dst = &ast.KeyValueExpr{}
	}
	tmpKey := psi.GetEdgeOrNil[Expr](n, EdgeKindKeyValueExprKey.Singleton())
	if tmpKey != nil {
		dst.Key = tmpKey.ToGoAst().(ast.Expr)
	}

	tmpValue := psi.GetEdgeOrNil[Expr](n, EdgeKindKeyValueExprValue.Singleton())
	if tmpValue != nil {
		dst.Value = tmpValue.ToGoAst().(ast.Expr)
	}

	return dst
}

type ArrayType struct {
	NodeBase
}

var ArrayTypeType = psi.DefineNodeType[*ArrayType]()

var EdgeKindArrayTypeLen = psi.DefineEdgeType[Expr]("GoArrayTypeLen")
var EdgeKindArrayTypeElt = psi.DefineEdgeType[Expr]("GoArrayTypeElt")

func (n *ArrayType) GetLen() Expr     { return psi.GetEdgeOrNil[Expr](n, EdgeKindArrayTypeLen.Singleton()) }
func (n *ArrayType) SetLen(node Expr) { psi.UpdateEdge(n, EdgeKindArrayTypeLen.Singleton(), node) }
func (n *ArrayType) GetElt() Expr     { return psi.GetEdgeOrNil[Expr](n, EdgeKindArrayTypeElt.Singleton()) }
func (n *ArrayType) SetElt(node Expr) { psi.UpdateEdge(n, EdgeKindArrayTypeElt.Singleton(), node) }
func NewFromArrayType(fset *token.FileSet, node *ast.ArrayType) *ArrayType {
	n := &ArrayType{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(ArrayTypeType))

	return n
}

func (n *ArrayType) CopyFromGoAst(fset *token.FileSet, src *ast.ArrayType) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Len != nil {
		tmpLen := NewFromExpr(fset, src.Len)
		tmpLen.SetParent(n)
		psi.UpdateEdge(n, EdgeKindArrayTypeLen.Singleton(), tmpLen)
	}

	if src.Elt != nil {
		tmpElt := NewFromExpr(fset, src.Elt)
		tmpElt.SetParent(n)
		psi.UpdateEdge(n, EdgeKindArrayTypeElt.Singleton(), tmpElt)
	}

}

func (n *ArrayType) ToGoAst() ast.Node { return n.ToGoArrayType(nil) }

func (n *ArrayType) ToGoArrayType(dst *ast.ArrayType) *ast.ArrayType {
	if dst == nil {
		dst = &ast.ArrayType{}
	}
	tmpLen := psi.GetEdgeOrNil[Expr](n, EdgeKindArrayTypeLen.Singleton())
	if tmpLen != nil {
		dst.Len = tmpLen.ToGoAst().(ast.Expr)
	}

	tmpElt := psi.GetEdgeOrNil[Expr](n, EdgeKindArrayTypeElt.Singleton())
	if tmpElt != nil {
		dst.Elt = tmpElt.ToGoAst().(ast.Expr)
	}

	return dst
}

type StructType struct {
	NodeBase
	Incomplete bool `json:"Incomplete"`
}

var StructTypeType = psi.DefineNodeType[*StructType]()

var EdgeKindStructTypeFields = psi.DefineEdgeType[*FieldList]("GoStructTypeFields")

func (n *StructType) GetFields() *FieldList {
	return psi.GetEdgeOrNil[*FieldList](n, EdgeKindStructTypeFields.Singleton())
}
func (n *StructType) SetFields(node *FieldList) {
	psi.UpdateEdge(n, EdgeKindStructTypeFields.Singleton(), node)
}
func NewFromStructType(fset *token.FileSet, node *ast.StructType) *StructType {
	n := &StructType{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(StructTypeType))

	return n
}

func (n *StructType) CopyFromGoAst(fset *token.FileSet, src *ast.StructType) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	n.Incomplete = src.Incomplete
	if src.Fields != nil {
		tmpFields := NewFromFieldList(fset, src.Fields)
		tmpFields.SetParent(n)
		psi.UpdateEdge(n, EdgeKindStructTypeFields.Singleton(), tmpFields)
	}

}

func (n *StructType) ToGoAst() ast.Node { return n.ToGoStructType(nil) }

func (n *StructType) ToGoStructType(dst *ast.StructType) *ast.StructType {
	if dst == nil {
		dst = &ast.StructType{}
	}
	dst.Incomplete = n.Incomplete
	tmpFields := psi.GetEdgeOrNil[*FieldList](n, EdgeKindStructTypeFields.Singleton())
	if tmpFields != nil {
		dst.Fields = tmpFields.ToGoAst().(*ast.FieldList)
	}

	return dst
}

type FuncType struct {
	NodeBase
}

var FuncTypeType = psi.DefineNodeType[*FuncType]()

var EdgeKindFuncTypeTypeParams = psi.DefineEdgeType[*FieldList]("GoFuncTypeTypeParams")
var EdgeKindFuncTypeParams = psi.DefineEdgeType[*FieldList]("GoFuncTypeParams")
var EdgeKindFuncTypeResults = psi.DefineEdgeType[*FieldList]("GoFuncTypeResults")

func (n *FuncType) GetTypeParams() *FieldList {
	return psi.GetEdgeOrNil[*FieldList](n, EdgeKindFuncTypeTypeParams.Singleton())
}
func (n *FuncType) SetTypeParams(node *FieldList) {
	psi.UpdateEdge(n, EdgeKindFuncTypeTypeParams.Singleton(), node)
}
func (n *FuncType) GetParams() *FieldList {
	return psi.GetEdgeOrNil[*FieldList](n, EdgeKindFuncTypeParams.Singleton())
}
func (n *FuncType) SetParams(node *FieldList) {
	psi.UpdateEdge(n, EdgeKindFuncTypeParams.Singleton(), node)
}
func (n *FuncType) GetResults() *FieldList {
	return psi.GetEdgeOrNil[*FieldList](n, EdgeKindFuncTypeResults.Singleton())
}
func (n *FuncType) SetResults(node *FieldList) {
	psi.UpdateEdge(n, EdgeKindFuncTypeResults.Singleton(), node)
}
func NewFromFuncType(fset *token.FileSet, node *ast.FuncType) *FuncType {
	n := &FuncType{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(FuncTypeType))

	return n
}

func (n *FuncType) CopyFromGoAst(fset *token.FileSet, src *ast.FuncType) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.TypeParams != nil {
		tmpTypeParams := NewFromFieldList(fset, src.TypeParams)
		tmpTypeParams.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFuncTypeTypeParams.Singleton(), tmpTypeParams)
	}

	if src.Params != nil {
		tmpParams := NewFromFieldList(fset, src.Params)
		tmpParams.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFuncTypeParams.Singleton(), tmpParams)
	}

	if src.Results != nil {
		tmpResults := NewFromFieldList(fset, src.Results)
		tmpResults.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFuncTypeResults.Singleton(), tmpResults)
	}

}

func (n *FuncType) ToGoAst() ast.Node { return n.ToGoFuncType(nil) }

func (n *FuncType) ToGoFuncType(dst *ast.FuncType) *ast.FuncType {
	if dst == nil {
		dst = &ast.FuncType{}
	}
	tmpTypeParams := psi.GetEdgeOrNil[*FieldList](n, EdgeKindFuncTypeTypeParams.Singleton())
	if tmpTypeParams != nil {
		dst.TypeParams = tmpTypeParams.ToGoAst().(*ast.FieldList)
	}

	tmpParams := psi.GetEdgeOrNil[*FieldList](n, EdgeKindFuncTypeParams.Singleton())
	if tmpParams != nil {
		dst.Params = tmpParams.ToGoAst().(*ast.FieldList)
	}

	tmpResults := psi.GetEdgeOrNil[*FieldList](n, EdgeKindFuncTypeResults.Singleton())
	if tmpResults != nil {
		dst.Results = tmpResults.ToGoAst().(*ast.FieldList)
	}

	return dst
}

type InterfaceType struct {
	NodeBase
	Incomplete bool `json:"Incomplete"`
}

var InterfaceTypeType = psi.DefineNodeType[*InterfaceType]()

var EdgeKindInterfaceTypeMethods = psi.DefineEdgeType[*FieldList]("GoInterfaceTypeMethods")

func (n *InterfaceType) GetMethods() *FieldList {
	return psi.GetEdgeOrNil[*FieldList](n, EdgeKindInterfaceTypeMethods.Singleton())
}
func (n *InterfaceType) SetMethods(node *FieldList) {
	psi.UpdateEdge(n, EdgeKindInterfaceTypeMethods.Singleton(), node)
}
func NewFromInterfaceType(fset *token.FileSet, node *ast.InterfaceType) *InterfaceType {
	n := &InterfaceType{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(InterfaceTypeType))

	return n
}

func (n *InterfaceType) CopyFromGoAst(fset *token.FileSet, src *ast.InterfaceType) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	n.Incomplete = src.Incomplete
	if src.Methods != nil {
		tmpMethods := NewFromFieldList(fset, src.Methods)
		tmpMethods.SetParent(n)
		psi.UpdateEdge(n, EdgeKindInterfaceTypeMethods.Singleton(), tmpMethods)
	}

}

func (n *InterfaceType) ToGoAst() ast.Node { return n.ToGoInterfaceType(nil) }

func (n *InterfaceType) ToGoInterfaceType(dst *ast.InterfaceType) *ast.InterfaceType {
	if dst == nil {
		dst = &ast.InterfaceType{}
	}
	dst.Incomplete = n.Incomplete
	tmpMethods := psi.GetEdgeOrNil[*FieldList](n, EdgeKindInterfaceTypeMethods.Singleton())
	if tmpMethods != nil {
		dst.Methods = tmpMethods.ToGoAst().(*ast.FieldList)
	}

	return dst
}

type MapType struct {
	NodeBase
}

var MapTypeType = psi.DefineNodeType[*MapType]()

var EdgeKindMapTypeKey = psi.DefineEdgeType[Expr]("GoMapTypeKey")
var EdgeKindMapTypeValue = psi.DefineEdgeType[Expr]("GoMapTypeValue")

func (n *MapType) GetKey() Expr       { return psi.GetEdgeOrNil[Expr](n, EdgeKindMapTypeKey.Singleton()) }
func (n *MapType) SetKey(node Expr)   { psi.UpdateEdge(n, EdgeKindMapTypeKey.Singleton(), node) }
func (n *MapType) GetValue() Expr     { return psi.GetEdgeOrNil[Expr](n, EdgeKindMapTypeValue.Singleton()) }
func (n *MapType) SetValue(node Expr) { psi.UpdateEdge(n, EdgeKindMapTypeValue.Singleton(), node) }
func NewFromMapType(fset *token.FileSet, node *ast.MapType) *MapType {
	n := &MapType{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(MapTypeType))

	return n
}

func (n *MapType) CopyFromGoAst(fset *token.FileSet, src *ast.MapType) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Key != nil {
		tmpKey := NewFromExpr(fset, src.Key)
		tmpKey.SetParent(n)
		psi.UpdateEdge(n, EdgeKindMapTypeKey.Singleton(), tmpKey)
	}

	if src.Value != nil {
		tmpValue := NewFromExpr(fset, src.Value)
		tmpValue.SetParent(n)
		psi.UpdateEdge(n, EdgeKindMapTypeValue.Singleton(), tmpValue)
	}

}

func (n *MapType) ToGoAst() ast.Node { return n.ToGoMapType(nil) }

func (n *MapType) ToGoMapType(dst *ast.MapType) *ast.MapType {
	if dst == nil {
		dst = &ast.MapType{}
	}
	tmpKey := psi.GetEdgeOrNil[Expr](n, EdgeKindMapTypeKey.Singleton())
	if tmpKey != nil {
		dst.Key = tmpKey.ToGoAst().(ast.Expr)
	}

	tmpValue := psi.GetEdgeOrNil[Expr](n, EdgeKindMapTypeValue.Singleton())
	if tmpValue != nil {
		dst.Value = tmpValue.ToGoAst().(ast.Expr)
	}

	return dst
}

type ChanType struct {
	NodeBase
	Dir ast.ChanDir `json:"Dir"`
}

var ChanTypeType = psi.DefineNodeType[*ChanType]()

var EdgeKindChanTypeValue = psi.DefineEdgeType[Expr]("GoChanTypeValue")

func (n *ChanType) GetValue() Expr {
	return psi.GetEdgeOrNil[Expr](n, EdgeKindChanTypeValue.Singleton())
}
func (n *ChanType) SetValue(node Expr) { psi.UpdateEdge(n, EdgeKindChanTypeValue.Singleton(), node) }
func NewFromChanType(fset *token.FileSet, node *ast.ChanType) *ChanType {
	n := &ChanType{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(ChanTypeType))

	return n
}

func (n *ChanType) CopyFromGoAst(fset *token.FileSet, src *ast.ChanType) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	n.Dir = src.Dir
	if src.Value != nil {
		tmpValue := NewFromExpr(fset, src.Value)
		tmpValue.SetParent(n)
		psi.UpdateEdge(n, EdgeKindChanTypeValue.Singleton(), tmpValue)
	}

}

func (n *ChanType) ToGoAst() ast.Node { return n.ToGoChanType(nil) }

func (n *ChanType) ToGoChanType(dst *ast.ChanType) *ast.ChanType {
	if dst == nil {
		dst = &ast.ChanType{}
	}
	dst.Dir = n.Dir
	tmpValue := psi.GetEdgeOrNil[Expr](n, EdgeKindChanTypeValue.Singleton())
	if tmpValue != nil {
		dst.Value = tmpValue.ToGoAst().(ast.Expr)
	}

	return dst
}

type BadStmt struct {
	NodeBase
}

var BadStmtType = psi.DefineNodeType[*BadStmt]()

func NewFromBadStmt(fset *token.FileSet, node *ast.BadStmt) *BadStmt {
	n := &BadStmt{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(BadStmtType))

	return n
}

func (n *BadStmt) CopyFromGoAst(fset *token.FileSet, src *ast.BadStmt) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
}

func (n *BadStmt) ToGoAst() ast.Node { return n.ToGoBadStmt(nil) }

func (n *BadStmt) ToGoBadStmt(dst *ast.BadStmt) *ast.BadStmt {
	if dst == nil {
		dst = &ast.BadStmt{}
	}

	return dst
}

type DeclStmt struct {
	NodeBase
}

var DeclStmtType = psi.DefineNodeType[*DeclStmt]()

var EdgeKindDeclStmtDecl = psi.DefineEdgeType[Decl]("GoDeclStmtDecl")

func (n *DeclStmt) GetDecl() Decl     { return psi.GetEdgeOrNil[Decl](n, EdgeKindDeclStmtDecl.Singleton()) }
func (n *DeclStmt) SetDecl(node Decl) { psi.UpdateEdge(n, EdgeKindDeclStmtDecl.Singleton(), node) }
func NewFromDeclStmt(fset *token.FileSet, node *ast.DeclStmt) *DeclStmt {
	n := &DeclStmt{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(DeclStmtType))

	return n
}

func (n *DeclStmt) CopyFromGoAst(fset *token.FileSet, src *ast.DeclStmt) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Decl != nil {
		tmpDecl := NewFromDecl(fset, src.Decl)
		tmpDecl.SetParent(n)
		psi.UpdateEdge(n, EdgeKindDeclStmtDecl.Singleton(), tmpDecl)
	}

}

func (n *DeclStmt) ToGoAst() ast.Node { return n.ToGoDeclStmt(nil) }

func (n *DeclStmt) ToGoDeclStmt(dst *ast.DeclStmt) *ast.DeclStmt {
	if dst == nil {
		dst = &ast.DeclStmt{}
	}
	tmpDecl := psi.GetEdgeOrNil[Decl](n, EdgeKindDeclStmtDecl.Singleton())
	if tmpDecl != nil {
		dst.Decl = tmpDecl.ToGoAst().(ast.Decl)
	}

	return dst
}

type EmptyStmt struct {
	NodeBase
	Implicit bool `json:"Implicit"`
}

var EmptyStmtType = psi.DefineNodeType[*EmptyStmt]()

func NewFromEmptyStmt(fset *token.FileSet, node *ast.EmptyStmt) *EmptyStmt {
	n := &EmptyStmt{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(EmptyStmtType))

	return n
}

func (n *EmptyStmt) CopyFromGoAst(fset *token.FileSet, src *ast.EmptyStmt) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	n.Implicit = src.Implicit
}

func (n *EmptyStmt) ToGoAst() ast.Node { return n.ToGoEmptyStmt(nil) }

func (n *EmptyStmt) ToGoEmptyStmt(dst *ast.EmptyStmt) *ast.EmptyStmt {
	if dst == nil {
		dst = &ast.EmptyStmt{}
	}
	dst.Implicit = n.Implicit

	return dst
}

type LabeledStmt struct {
	NodeBase
}

var LabeledStmtType = psi.DefineNodeType[*LabeledStmt]()

var EdgeKindLabeledStmtLabel = psi.DefineEdgeType[*Ident]("GoLabeledStmtLabel")
var EdgeKindLabeledStmtStmt = psi.DefineEdgeType[Stmt]("GoLabeledStmtStmt")

func (n *LabeledStmt) GetLabel() *Ident {
	return psi.GetEdgeOrNil[*Ident](n, EdgeKindLabeledStmtLabel.Singleton())
}
func (n *LabeledStmt) SetLabel(node *Ident) {
	psi.UpdateEdge(n, EdgeKindLabeledStmtLabel.Singleton(), node)
}
func (n *LabeledStmt) GetStmt() Stmt {
	return psi.GetEdgeOrNil[Stmt](n, EdgeKindLabeledStmtStmt.Singleton())
}
func (n *LabeledStmt) SetStmt(node Stmt) {
	psi.UpdateEdge(n, EdgeKindLabeledStmtStmt.Singleton(), node)
}
func NewFromLabeledStmt(fset *token.FileSet, node *ast.LabeledStmt) *LabeledStmt {
	n := &LabeledStmt{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(LabeledStmtType))

	return n
}

func (n *LabeledStmt) CopyFromGoAst(fset *token.FileSet, src *ast.LabeledStmt) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Label != nil {
		tmpLabel := NewFromIdent(fset, src.Label)
		tmpLabel.SetParent(n)
		psi.UpdateEdge(n, EdgeKindLabeledStmtLabel.Singleton(), tmpLabel)
	}

	if src.Stmt != nil {
		tmpStmt := NewFromStmt(fset, src.Stmt)
		tmpStmt.SetParent(n)
		psi.UpdateEdge(n, EdgeKindLabeledStmtStmt.Singleton(), tmpStmt)
	}

}

func (n *LabeledStmt) ToGoAst() ast.Node { return n.ToGoLabeledStmt(nil) }

func (n *LabeledStmt) ToGoLabeledStmt(dst *ast.LabeledStmt) *ast.LabeledStmt {
	if dst == nil {
		dst = &ast.LabeledStmt{}
	}
	tmpLabel := psi.GetEdgeOrNil[*Ident](n, EdgeKindLabeledStmtLabel.Singleton())
	if tmpLabel != nil {
		dst.Label = tmpLabel.ToGoAst().(*ast.Ident)
	}

	tmpStmt := psi.GetEdgeOrNil[Stmt](n, EdgeKindLabeledStmtStmt.Singleton())
	if tmpStmt != nil {
		dst.Stmt = tmpStmt.ToGoAst().(ast.Stmt)
	}

	return dst
}

type ExprStmt struct {
	NodeBase
}

var ExprStmtType = psi.DefineNodeType[*ExprStmt]()

var EdgeKindExprStmtX = psi.DefineEdgeType[Expr]("GoExprStmtX")

func (n *ExprStmt) GetX() Expr     { return psi.GetEdgeOrNil[Expr](n, EdgeKindExprStmtX.Singleton()) }
func (n *ExprStmt) SetX(node Expr) { psi.UpdateEdge(n, EdgeKindExprStmtX.Singleton(), node) }
func NewFromExprStmt(fset *token.FileSet, node *ast.ExprStmt) *ExprStmt {
	n := &ExprStmt{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(ExprStmtType))

	return n
}

func (n *ExprStmt) CopyFromGoAst(fset *token.FileSet, src *ast.ExprStmt) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.X != nil {
		tmpX := NewFromExpr(fset, src.X)
		tmpX.SetParent(n)
		psi.UpdateEdge(n, EdgeKindExprStmtX.Singleton(), tmpX)
	}

}

func (n *ExprStmt) ToGoAst() ast.Node { return n.ToGoExprStmt(nil) }

func (n *ExprStmt) ToGoExprStmt(dst *ast.ExprStmt) *ast.ExprStmt {
	if dst == nil {
		dst = &ast.ExprStmt{}
	}
	tmpX := psi.GetEdgeOrNil[Expr](n, EdgeKindExprStmtX.Singleton())
	if tmpX != nil {
		dst.X = tmpX.ToGoAst().(ast.Expr)
	}

	return dst
}

type SendStmt struct {
	NodeBase
}

var SendStmtType = psi.DefineNodeType[*SendStmt]()

var EdgeKindSendStmtChan = psi.DefineEdgeType[Expr]("GoSendStmtChan")
var EdgeKindSendStmtValue = psi.DefineEdgeType[Expr]("GoSendStmtValue")

func (n *SendStmt) GetChan() Expr     { return psi.GetEdgeOrNil[Expr](n, EdgeKindSendStmtChan.Singleton()) }
func (n *SendStmt) SetChan(node Expr) { psi.UpdateEdge(n, EdgeKindSendStmtChan.Singleton(), node) }
func (n *SendStmt) GetValue() Expr {
	return psi.GetEdgeOrNil[Expr](n, EdgeKindSendStmtValue.Singleton())
}
func (n *SendStmt) SetValue(node Expr) { psi.UpdateEdge(n, EdgeKindSendStmtValue.Singleton(), node) }
func NewFromSendStmt(fset *token.FileSet, node *ast.SendStmt) *SendStmt {
	n := &SendStmt{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(SendStmtType))

	return n
}

func (n *SendStmt) CopyFromGoAst(fset *token.FileSet, src *ast.SendStmt) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Chan != nil {
		tmpChan := NewFromExpr(fset, src.Chan)
		tmpChan.SetParent(n)
		psi.UpdateEdge(n, EdgeKindSendStmtChan.Singleton(), tmpChan)
	}

	if src.Value != nil {
		tmpValue := NewFromExpr(fset, src.Value)
		tmpValue.SetParent(n)
		psi.UpdateEdge(n, EdgeKindSendStmtValue.Singleton(), tmpValue)
	}

}

func (n *SendStmt) ToGoAst() ast.Node { return n.ToGoSendStmt(nil) }

func (n *SendStmt) ToGoSendStmt(dst *ast.SendStmt) *ast.SendStmt {
	if dst == nil {
		dst = &ast.SendStmt{}
	}
	tmpChan := psi.GetEdgeOrNil[Expr](n, EdgeKindSendStmtChan.Singleton())
	if tmpChan != nil {
		dst.Chan = tmpChan.ToGoAst().(ast.Expr)
	}

	tmpValue := psi.GetEdgeOrNil[Expr](n, EdgeKindSendStmtValue.Singleton())
	if tmpValue != nil {
		dst.Value = tmpValue.ToGoAst().(ast.Expr)
	}

	return dst
}

type IncDecStmt struct {
	NodeBase
	Tok token.Token `json:"Tok"`
}

var IncDecStmtType = psi.DefineNodeType[*IncDecStmt]()

var EdgeKindIncDecStmtX = psi.DefineEdgeType[Expr]("GoIncDecStmtX")

func (n *IncDecStmt) GetX() Expr     { return psi.GetEdgeOrNil[Expr](n, EdgeKindIncDecStmtX.Singleton()) }
func (n *IncDecStmt) SetX(node Expr) { psi.UpdateEdge(n, EdgeKindIncDecStmtX.Singleton(), node) }
func NewFromIncDecStmt(fset *token.FileSet, node *ast.IncDecStmt) *IncDecStmt {
	n := &IncDecStmt{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(IncDecStmtType))

	return n
}

func (n *IncDecStmt) CopyFromGoAst(fset *token.FileSet, src *ast.IncDecStmt) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	n.Tok = src.Tok
	if src.X != nil {
		tmpX := NewFromExpr(fset, src.X)
		tmpX.SetParent(n)
		psi.UpdateEdge(n, EdgeKindIncDecStmtX.Singleton(), tmpX)
	}

}

func (n *IncDecStmt) ToGoAst() ast.Node { return n.ToGoIncDecStmt(nil) }

func (n *IncDecStmt) ToGoIncDecStmt(dst *ast.IncDecStmt) *ast.IncDecStmt {
	if dst == nil {
		dst = &ast.IncDecStmt{}
	}
	dst.Tok = n.Tok
	tmpX := psi.GetEdgeOrNil[Expr](n, EdgeKindIncDecStmtX.Singleton())
	if tmpX != nil {
		dst.X = tmpX.ToGoAst().(ast.Expr)
	}

	return dst
}

type AssignStmt struct {
	NodeBase
	Tok token.Token `json:"Tok"`
}

var AssignStmtType = psi.DefineNodeType[*AssignStmt]()

var EdgeKindAssignStmtLhs = psi.DefineEdgeType[Expr]("GoAssignStmtLhs")
var EdgeKindAssignStmtRhs = psi.DefineEdgeType[Expr]("GoAssignStmtRhs")

func (n *AssignStmt) GetLhs() []Expr      { return psi.GetEdges(n, EdgeKindAssignStmtLhs) }
func (n *AssignStmt) SetLhs(nodes []Expr) { psi.UpdateEdges(n, EdgeKindAssignStmtLhs, nodes) }
func (n *AssignStmt) GetRhs() []Expr      { return psi.GetEdges(n, EdgeKindAssignStmtRhs) }
func (n *AssignStmt) SetRhs(nodes []Expr) { psi.UpdateEdges(n, EdgeKindAssignStmtRhs, nodes) }
func NewFromAssignStmt(fset *token.FileSet, node *ast.AssignStmt) *AssignStmt {
	n := &AssignStmt{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(AssignStmtType))

	return n
}

func (n *AssignStmt) CopyFromGoAst(fset *token.FileSet, src *ast.AssignStmt) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	n.Tok = src.Tok
	for i, v := range src.Lhs {
		tmpLhs := NewFromExpr(fset, v)
		tmpLhs.SetParent(n)
		psi.UpdateEdge(n, EdgeKindAssignStmtLhs.Indexed(int64(i)), tmpLhs)
	}

	for i, v := range src.Rhs {
		tmpRhs := NewFromExpr(fset, v)
		tmpRhs.SetParent(n)
		psi.UpdateEdge(n, EdgeKindAssignStmtRhs.Indexed(int64(i)), tmpRhs)
	}

}

func (n *AssignStmt) ToGoAst() ast.Node { return n.ToGoAssignStmt(nil) }

func (n *AssignStmt) ToGoAssignStmt(dst *ast.AssignStmt) *ast.AssignStmt {
	if dst == nil {
		dst = &ast.AssignStmt{}
	}
	dst.Tok = n.Tok
	tmpLhs := psi.GetEdges(n, EdgeKindAssignStmtLhs)
	dst.Lhs = make([]ast.Expr, len(tmpLhs))
	for i, v := range tmpLhs {
		dst.Lhs[i] = v.ToGoAst().(ast.Expr)
	}

	tmpRhs := psi.GetEdges(n, EdgeKindAssignStmtRhs)
	dst.Rhs = make([]ast.Expr, len(tmpRhs))
	for i, v := range tmpRhs {
		dst.Rhs[i] = v.ToGoAst().(ast.Expr)
	}

	return dst
}

type GoStmt struct {
	NodeBase
}

var GoStmtType = psi.DefineNodeType[*GoStmt]()

var EdgeKindGoStmtCall = psi.DefineEdgeType[*CallExpr]("GoGoStmtCall")

func (n *GoStmt) GetCall() *CallExpr {
	return psi.GetEdgeOrNil[*CallExpr](n, EdgeKindGoStmtCall.Singleton())
}
func (n *GoStmt) SetCall(node *CallExpr) { psi.UpdateEdge(n, EdgeKindGoStmtCall.Singleton(), node) }
func NewFromGoStmt(fset *token.FileSet, node *ast.GoStmt) *GoStmt {
	n := &GoStmt{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(GoStmtType))

	return n
}

func (n *GoStmt) CopyFromGoAst(fset *token.FileSet, src *ast.GoStmt) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Call != nil {
		tmpCall := NewFromCallExpr(fset, src.Call)
		tmpCall.SetParent(n)
		psi.UpdateEdge(n, EdgeKindGoStmtCall.Singleton(), tmpCall)
	}

}

func (n *GoStmt) ToGoAst() ast.Node { return n.ToGoGoStmt(nil) }

func (n *GoStmt) ToGoGoStmt(dst *ast.GoStmt) *ast.GoStmt {
	if dst == nil {
		dst = &ast.GoStmt{}
	}
	tmpCall := psi.GetEdgeOrNil[*CallExpr](n, EdgeKindGoStmtCall.Singleton())
	if tmpCall != nil {
		dst.Call = tmpCall.ToGoAst().(*ast.CallExpr)
	}

	return dst
}

type DeferStmt struct {
	NodeBase
}

var DeferStmtType = psi.DefineNodeType[*DeferStmt]()

var EdgeKindDeferStmtCall = psi.DefineEdgeType[*CallExpr]("GoDeferStmtCall")

func (n *DeferStmt) GetCall() *CallExpr {
	return psi.GetEdgeOrNil[*CallExpr](n, EdgeKindDeferStmtCall.Singleton())
}
func (n *DeferStmt) SetCall(node *CallExpr) {
	psi.UpdateEdge(n, EdgeKindDeferStmtCall.Singleton(), node)
}
func NewFromDeferStmt(fset *token.FileSet, node *ast.DeferStmt) *DeferStmt {
	n := &DeferStmt{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(DeferStmtType))

	return n
}

func (n *DeferStmt) CopyFromGoAst(fset *token.FileSet, src *ast.DeferStmt) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Call != nil {
		tmpCall := NewFromCallExpr(fset, src.Call)
		tmpCall.SetParent(n)
		psi.UpdateEdge(n, EdgeKindDeferStmtCall.Singleton(), tmpCall)
	}

}

func (n *DeferStmt) ToGoAst() ast.Node { return n.ToGoDeferStmt(nil) }

func (n *DeferStmt) ToGoDeferStmt(dst *ast.DeferStmt) *ast.DeferStmt {
	if dst == nil {
		dst = &ast.DeferStmt{}
	}
	tmpCall := psi.GetEdgeOrNil[*CallExpr](n, EdgeKindDeferStmtCall.Singleton())
	if tmpCall != nil {
		dst.Call = tmpCall.ToGoAst().(*ast.CallExpr)
	}

	return dst
}

type ReturnStmt struct {
	NodeBase
}

var ReturnStmtType = psi.DefineNodeType[*ReturnStmt]()

var EdgeKindReturnStmtResults = psi.DefineEdgeType[Expr]("GoReturnStmtResults")

func (n *ReturnStmt) GetResults() []Expr      { return psi.GetEdges(n, EdgeKindReturnStmtResults) }
func (n *ReturnStmt) SetResults(nodes []Expr) { psi.UpdateEdges(n, EdgeKindReturnStmtResults, nodes) }
func NewFromReturnStmt(fset *token.FileSet, node *ast.ReturnStmt) *ReturnStmt {
	n := &ReturnStmt{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(ReturnStmtType))

	return n
}

func (n *ReturnStmt) CopyFromGoAst(fset *token.FileSet, src *ast.ReturnStmt) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	for i, v := range src.Results {
		tmpResults := NewFromExpr(fset, v)
		tmpResults.SetParent(n)
		psi.UpdateEdge(n, EdgeKindReturnStmtResults.Indexed(int64(i)), tmpResults)
	}

}

func (n *ReturnStmt) ToGoAst() ast.Node { return n.ToGoReturnStmt(nil) }

func (n *ReturnStmt) ToGoReturnStmt(dst *ast.ReturnStmt) *ast.ReturnStmt {
	if dst == nil {
		dst = &ast.ReturnStmt{}
	}
	tmpResults := psi.GetEdges(n, EdgeKindReturnStmtResults)
	dst.Results = make([]ast.Expr, len(tmpResults))
	for i, v := range tmpResults {
		dst.Results[i] = v.ToGoAst().(ast.Expr)
	}

	return dst
}

type BranchStmt struct {
	NodeBase
	Tok token.Token `json:"Tok"`
}

var BranchStmtType = psi.DefineNodeType[*BranchStmt]()

var EdgeKindBranchStmtLabel = psi.DefineEdgeType[*Ident]("GoBranchStmtLabel")

func (n *BranchStmt) GetLabel() *Ident {
	return psi.GetEdgeOrNil[*Ident](n, EdgeKindBranchStmtLabel.Singleton())
}
func (n *BranchStmt) SetLabel(node *Ident) {
	psi.UpdateEdge(n, EdgeKindBranchStmtLabel.Singleton(), node)
}
func NewFromBranchStmt(fset *token.FileSet, node *ast.BranchStmt) *BranchStmt {
	n := &BranchStmt{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(BranchStmtType))

	return n
}

func (n *BranchStmt) CopyFromGoAst(fset *token.FileSet, src *ast.BranchStmt) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	n.Tok = src.Tok
	if src.Label != nil {
		tmpLabel := NewFromIdent(fset, src.Label)
		tmpLabel.SetParent(n)
		psi.UpdateEdge(n, EdgeKindBranchStmtLabel.Singleton(), tmpLabel)
	}

}

func (n *BranchStmt) ToGoAst() ast.Node { return n.ToGoBranchStmt(nil) }

func (n *BranchStmt) ToGoBranchStmt(dst *ast.BranchStmt) *ast.BranchStmt {
	if dst == nil {
		dst = &ast.BranchStmt{}
	}
	dst.Tok = n.Tok
	tmpLabel := psi.GetEdgeOrNil[*Ident](n, EdgeKindBranchStmtLabel.Singleton())
	if tmpLabel != nil {
		dst.Label = tmpLabel.ToGoAst().(*ast.Ident)
	}

	return dst
}

type BlockStmt struct {
	NodeBase
}

var BlockStmtType = psi.DefineNodeType[*BlockStmt]()

var EdgeKindBlockStmtList = psi.DefineEdgeType[Stmt]("GoBlockStmtList")

func (n *BlockStmt) GetList() []Stmt      { return psi.GetEdges(n, EdgeKindBlockStmtList) }
func (n *BlockStmt) SetList(nodes []Stmt) { psi.UpdateEdges(n, EdgeKindBlockStmtList, nodes) }
func NewFromBlockStmt(fset *token.FileSet, node *ast.BlockStmt) *BlockStmt {
	n := &BlockStmt{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(BlockStmtType))

	return n
}

func (n *BlockStmt) CopyFromGoAst(fset *token.FileSet, src *ast.BlockStmt) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	for i, v := range src.List {
		tmpList := NewFromStmt(fset, v)
		tmpList.SetParent(n)
		psi.UpdateEdge(n, EdgeKindBlockStmtList.Indexed(int64(i)), tmpList)
	}

}

func (n *BlockStmt) ToGoAst() ast.Node { return n.ToGoBlockStmt(nil) }

func (n *BlockStmt) ToGoBlockStmt(dst *ast.BlockStmt) *ast.BlockStmt {
	if dst == nil {
		dst = &ast.BlockStmt{}
	}
	tmpList := psi.GetEdges(n, EdgeKindBlockStmtList)
	dst.List = make([]ast.Stmt, len(tmpList))
	for i, v := range tmpList {
		dst.List[i] = v.ToGoAst().(ast.Stmt)
	}

	return dst
}

type IfStmt struct {
	NodeBase
}

var IfStmtType = psi.DefineNodeType[*IfStmt]()

var EdgeKindIfStmtInit = psi.DefineEdgeType[Stmt]("GoIfStmtInit")
var EdgeKindIfStmtCond = psi.DefineEdgeType[Expr]("GoIfStmtCond")
var EdgeKindIfStmtBody = psi.DefineEdgeType[*BlockStmt]("GoIfStmtBody")
var EdgeKindIfStmtElse = psi.DefineEdgeType[Stmt]("GoIfStmtElse")

func (n *IfStmt) GetInit() Stmt     { return psi.GetEdgeOrNil[Stmt](n, EdgeKindIfStmtInit.Singleton()) }
func (n *IfStmt) SetInit(node Stmt) { psi.UpdateEdge(n, EdgeKindIfStmtInit.Singleton(), node) }
func (n *IfStmt) GetCond() Expr     { return psi.GetEdgeOrNil[Expr](n, EdgeKindIfStmtCond.Singleton()) }
func (n *IfStmt) SetCond(node Expr) { psi.UpdateEdge(n, EdgeKindIfStmtCond.Singleton(), node) }
func (n *IfStmt) GetBody() *BlockStmt {
	return psi.GetEdgeOrNil[*BlockStmt](n, EdgeKindIfStmtBody.Singleton())
}
func (n *IfStmt) SetBody(node *BlockStmt) { psi.UpdateEdge(n, EdgeKindIfStmtBody.Singleton(), node) }
func (n *IfStmt) GetElse() Stmt           { return psi.GetEdgeOrNil[Stmt](n, EdgeKindIfStmtElse.Singleton()) }
func (n *IfStmt) SetElse(node Stmt)       { psi.UpdateEdge(n, EdgeKindIfStmtElse.Singleton(), node) }
func NewFromIfStmt(fset *token.FileSet, node *ast.IfStmt) *IfStmt {
	n := &IfStmt{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(IfStmtType))

	return n
}

func (n *IfStmt) CopyFromGoAst(fset *token.FileSet, src *ast.IfStmt) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Init != nil {
		tmpInit := NewFromStmt(fset, src.Init)
		tmpInit.SetParent(n)
		psi.UpdateEdge(n, EdgeKindIfStmtInit.Singleton(), tmpInit)
	}

	if src.Cond != nil {
		tmpCond := NewFromExpr(fset, src.Cond)
		tmpCond.SetParent(n)
		psi.UpdateEdge(n, EdgeKindIfStmtCond.Singleton(), tmpCond)
	}

	if src.Body != nil {
		tmpBody := NewFromBlockStmt(fset, src.Body)
		tmpBody.SetParent(n)
		psi.UpdateEdge(n, EdgeKindIfStmtBody.Singleton(), tmpBody)
	}

	if src.Else != nil {
		tmpElse := NewFromStmt(fset, src.Else)
		tmpElse.SetParent(n)
		psi.UpdateEdge(n, EdgeKindIfStmtElse.Singleton(), tmpElse)
	}

}

func (n *IfStmt) ToGoAst() ast.Node { return n.ToGoIfStmt(nil) }

func (n *IfStmt) ToGoIfStmt(dst *ast.IfStmt) *ast.IfStmt {
	if dst == nil {
		dst = &ast.IfStmt{}
	}
	tmpInit := psi.GetEdgeOrNil[Stmt](n, EdgeKindIfStmtInit.Singleton())
	if tmpInit != nil {
		dst.Init = tmpInit.ToGoAst().(ast.Stmt)
	}

	tmpCond := psi.GetEdgeOrNil[Expr](n, EdgeKindIfStmtCond.Singleton())
	if tmpCond != nil {
		dst.Cond = tmpCond.ToGoAst().(ast.Expr)
	}

	tmpBody := psi.GetEdgeOrNil[*BlockStmt](n, EdgeKindIfStmtBody.Singleton())
	if tmpBody != nil {
		dst.Body = tmpBody.ToGoAst().(*ast.BlockStmt)
	}

	tmpElse := psi.GetEdgeOrNil[Stmt](n, EdgeKindIfStmtElse.Singleton())
	if tmpElse != nil {
		dst.Else = tmpElse.ToGoAst().(ast.Stmt)
	}

	return dst
}

type CaseClause struct {
	NodeBase
}

var CaseClauseType = psi.DefineNodeType[*CaseClause]()

var EdgeKindCaseClauseList = psi.DefineEdgeType[Expr]("GoCaseClauseList")
var EdgeKindCaseClauseBody = psi.DefineEdgeType[Stmt]("GoCaseClauseBody")

func (n *CaseClause) GetList() []Expr      { return psi.GetEdges(n, EdgeKindCaseClauseList) }
func (n *CaseClause) SetList(nodes []Expr) { psi.UpdateEdges(n, EdgeKindCaseClauseList, nodes) }
func (n *CaseClause) GetBody() []Stmt      { return psi.GetEdges(n, EdgeKindCaseClauseBody) }
func (n *CaseClause) SetBody(nodes []Stmt) { psi.UpdateEdges(n, EdgeKindCaseClauseBody, nodes) }
func NewFromCaseClause(fset *token.FileSet, node *ast.CaseClause) *CaseClause {
	n := &CaseClause{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(CaseClauseType))

	return n
}

func (n *CaseClause) CopyFromGoAst(fset *token.FileSet, src *ast.CaseClause) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	for i, v := range src.List {
		tmpList := NewFromExpr(fset, v)
		tmpList.SetParent(n)
		psi.UpdateEdge(n, EdgeKindCaseClauseList.Indexed(int64(i)), tmpList)
	}

	for i, v := range src.Body {
		tmpBody := NewFromStmt(fset, v)
		tmpBody.SetParent(n)
		psi.UpdateEdge(n, EdgeKindCaseClauseBody.Indexed(int64(i)), tmpBody)
	}

}

func (n *CaseClause) ToGoAst() ast.Node { return n.ToGoCaseClause(nil) }

func (n *CaseClause) ToGoCaseClause(dst *ast.CaseClause) *ast.CaseClause {
	if dst == nil {
		dst = &ast.CaseClause{}
	}
	tmpList := psi.GetEdges(n, EdgeKindCaseClauseList)
	dst.List = make([]ast.Expr, len(tmpList))
	for i, v := range tmpList {
		dst.List[i] = v.ToGoAst().(ast.Expr)
	}

	tmpBody := psi.GetEdges(n, EdgeKindCaseClauseBody)
	dst.Body = make([]ast.Stmt, len(tmpBody))
	for i, v := range tmpBody {
		dst.Body[i] = v.ToGoAst().(ast.Stmt)
	}

	return dst
}

type SwitchStmt struct {
	NodeBase
}

var SwitchStmtType = psi.DefineNodeType[*SwitchStmt]()

var EdgeKindSwitchStmtInit = psi.DefineEdgeType[Stmt]("GoSwitchStmtInit")
var EdgeKindSwitchStmtTag = psi.DefineEdgeType[Expr]("GoSwitchStmtTag")
var EdgeKindSwitchStmtBody = psi.DefineEdgeType[*BlockStmt]("GoSwitchStmtBody")

func (n *SwitchStmt) GetInit() Stmt {
	return psi.GetEdgeOrNil[Stmt](n, EdgeKindSwitchStmtInit.Singleton())
}
func (n *SwitchStmt) SetInit(node Stmt) { psi.UpdateEdge(n, EdgeKindSwitchStmtInit.Singleton(), node) }
func (n *SwitchStmt) GetTag() Expr {
	return psi.GetEdgeOrNil[Expr](n, EdgeKindSwitchStmtTag.Singleton())
}
func (n *SwitchStmt) SetTag(node Expr) { psi.UpdateEdge(n, EdgeKindSwitchStmtTag.Singleton(), node) }
func (n *SwitchStmt) GetBody() *BlockStmt {
	return psi.GetEdgeOrNil[*BlockStmt](n, EdgeKindSwitchStmtBody.Singleton())
}
func (n *SwitchStmt) SetBody(node *BlockStmt) {
	psi.UpdateEdge(n, EdgeKindSwitchStmtBody.Singleton(), node)
}
func NewFromSwitchStmt(fset *token.FileSet, node *ast.SwitchStmt) *SwitchStmt {
	n := &SwitchStmt{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(SwitchStmtType))

	return n
}

func (n *SwitchStmt) CopyFromGoAst(fset *token.FileSet, src *ast.SwitchStmt) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Init != nil {
		tmpInit := NewFromStmt(fset, src.Init)
		tmpInit.SetParent(n)
		psi.UpdateEdge(n, EdgeKindSwitchStmtInit.Singleton(), tmpInit)
	}

	if src.Tag != nil {
		tmpTag := NewFromExpr(fset, src.Tag)
		tmpTag.SetParent(n)
		psi.UpdateEdge(n, EdgeKindSwitchStmtTag.Singleton(), tmpTag)
	}

	if src.Body != nil {
		tmpBody := NewFromBlockStmt(fset, src.Body)
		tmpBody.SetParent(n)
		psi.UpdateEdge(n, EdgeKindSwitchStmtBody.Singleton(), tmpBody)
	}

}

func (n *SwitchStmt) ToGoAst() ast.Node { return n.ToGoSwitchStmt(nil) }

func (n *SwitchStmt) ToGoSwitchStmt(dst *ast.SwitchStmt) *ast.SwitchStmt {
	if dst == nil {
		dst = &ast.SwitchStmt{}
	}
	tmpInit := psi.GetEdgeOrNil[Stmt](n, EdgeKindSwitchStmtInit.Singleton())
	if tmpInit != nil {
		dst.Init = tmpInit.ToGoAst().(ast.Stmt)
	}

	tmpTag := psi.GetEdgeOrNil[Expr](n, EdgeKindSwitchStmtTag.Singleton())
	if tmpTag != nil {
		dst.Tag = tmpTag.ToGoAst().(ast.Expr)
	}

	tmpBody := psi.GetEdgeOrNil[*BlockStmt](n, EdgeKindSwitchStmtBody.Singleton())
	if tmpBody != nil {
		dst.Body = tmpBody.ToGoAst().(*ast.BlockStmt)
	}

	return dst
}

type TypeSwitchStmt struct {
	NodeBase
}

var TypeSwitchStmtType = psi.DefineNodeType[*TypeSwitchStmt]()

var EdgeKindTypeSwitchStmtInit = psi.DefineEdgeType[Stmt]("GoTypeSwitchStmtInit")
var EdgeKindTypeSwitchStmtAssign = psi.DefineEdgeType[Stmt]("GoTypeSwitchStmtAssign")
var EdgeKindTypeSwitchStmtBody = psi.DefineEdgeType[*BlockStmt]("GoTypeSwitchStmtBody")

func (n *TypeSwitchStmt) GetInit() Stmt {
	return psi.GetEdgeOrNil[Stmt](n, EdgeKindTypeSwitchStmtInit.Singleton())
}
func (n *TypeSwitchStmt) SetInit(node Stmt) {
	psi.UpdateEdge(n, EdgeKindTypeSwitchStmtInit.Singleton(), node)
}
func (n *TypeSwitchStmt) GetAssign() Stmt {
	return psi.GetEdgeOrNil[Stmt](n, EdgeKindTypeSwitchStmtAssign.Singleton())
}
func (n *TypeSwitchStmt) SetAssign(node Stmt) {
	psi.UpdateEdge(n, EdgeKindTypeSwitchStmtAssign.Singleton(), node)
}
func (n *TypeSwitchStmt) GetBody() *BlockStmt {
	return psi.GetEdgeOrNil[*BlockStmt](n, EdgeKindTypeSwitchStmtBody.Singleton())
}
func (n *TypeSwitchStmt) SetBody(node *BlockStmt) {
	psi.UpdateEdge(n, EdgeKindTypeSwitchStmtBody.Singleton(), node)
}
func NewFromTypeSwitchStmt(fset *token.FileSet, node *ast.TypeSwitchStmt) *TypeSwitchStmt {
	n := &TypeSwitchStmt{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(TypeSwitchStmtType))

	return n
}

func (n *TypeSwitchStmt) CopyFromGoAst(fset *token.FileSet, src *ast.TypeSwitchStmt) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Init != nil {
		tmpInit := NewFromStmt(fset, src.Init)
		tmpInit.SetParent(n)
		psi.UpdateEdge(n, EdgeKindTypeSwitchStmtInit.Singleton(), tmpInit)
	}

	if src.Assign != nil {
		tmpAssign := NewFromStmt(fset, src.Assign)
		tmpAssign.SetParent(n)
		psi.UpdateEdge(n, EdgeKindTypeSwitchStmtAssign.Singleton(), tmpAssign)
	}

	if src.Body != nil {
		tmpBody := NewFromBlockStmt(fset, src.Body)
		tmpBody.SetParent(n)
		psi.UpdateEdge(n, EdgeKindTypeSwitchStmtBody.Singleton(), tmpBody)
	}

}

func (n *TypeSwitchStmt) ToGoAst() ast.Node { return n.ToGoTypeSwitchStmt(nil) }

func (n *TypeSwitchStmt) ToGoTypeSwitchStmt(dst *ast.TypeSwitchStmt) *ast.TypeSwitchStmt {
	if dst == nil {
		dst = &ast.TypeSwitchStmt{}
	}
	tmpInit := psi.GetEdgeOrNil[Stmt](n, EdgeKindTypeSwitchStmtInit.Singleton())
	if tmpInit != nil {
		dst.Init = tmpInit.ToGoAst().(ast.Stmt)
	}

	tmpAssign := psi.GetEdgeOrNil[Stmt](n, EdgeKindTypeSwitchStmtAssign.Singleton())
	if tmpAssign != nil {
		dst.Assign = tmpAssign.ToGoAst().(ast.Stmt)
	}

	tmpBody := psi.GetEdgeOrNil[*BlockStmt](n, EdgeKindTypeSwitchStmtBody.Singleton())
	if tmpBody != nil {
		dst.Body = tmpBody.ToGoAst().(*ast.BlockStmt)
	}

	return dst
}

type CommClause struct {
	NodeBase
}

var CommClauseType = psi.DefineNodeType[*CommClause]()

var EdgeKindCommClauseComm = psi.DefineEdgeType[Stmt]("GoCommClauseComm")
var EdgeKindCommClauseBody = psi.DefineEdgeType[Stmt]("GoCommClauseBody")

func (n *CommClause) GetComm() Stmt {
	return psi.GetEdgeOrNil[Stmt](n, EdgeKindCommClauseComm.Singleton())
}
func (n *CommClause) SetComm(node Stmt)    { psi.UpdateEdge(n, EdgeKindCommClauseComm.Singleton(), node) }
func (n *CommClause) GetBody() []Stmt      { return psi.GetEdges(n, EdgeKindCommClauseBody) }
func (n *CommClause) SetBody(nodes []Stmt) { psi.UpdateEdges(n, EdgeKindCommClauseBody, nodes) }
func NewFromCommClause(fset *token.FileSet, node *ast.CommClause) *CommClause {
	n := &CommClause{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(CommClauseType))

	return n
}

func (n *CommClause) CopyFromGoAst(fset *token.FileSet, src *ast.CommClause) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Comm != nil {
		tmpComm := NewFromStmt(fset, src.Comm)
		tmpComm.SetParent(n)
		psi.UpdateEdge(n, EdgeKindCommClauseComm.Singleton(), tmpComm)
	}

	for i, v := range src.Body {
		tmpBody := NewFromStmt(fset, v)
		tmpBody.SetParent(n)
		psi.UpdateEdge(n, EdgeKindCommClauseBody.Indexed(int64(i)), tmpBody)
	}

}

func (n *CommClause) ToGoAst() ast.Node { return n.ToGoCommClause(nil) }

func (n *CommClause) ToGoCommClause(dst *ast.CommClause) *ast.CommClause {
	if dst == nil {
		dst = &ast.CommClause{}
	}
	tmpComm := psi.GetEdgeOrNil[Stmt](n, EdgeKindCommClauseComm.Singleton())
	if tmpComm != nil {
		dst.Comm = tmpComm.ToGoAst().(ast.Stmt)
	}

	tmpBody := psi.GetEdges(n, EdgeKindCommClauseBody)
	dst.Body = make([]ast.Stmt, len(tmpBody))
	for i, v := range tmpBody {
		dst.Body[i] = v.ToGoAst().(ast.Stmt)
	}

	return dst
}

type SelectStmt struct {
	NodeBase
}

var SelectStmtType = psi.DefineNodeType[*SelectStmt]()

var EdgeKindSelectStmtBody = psi.DefineEdgeType[*BlockStmt]("GoSelectStmtBody")

func (n *SelectStmt) GetBody() *BlockStmt {
	return psi.GetEdgeOrNil[*BlockStmt](n, EdgeKindSelectStmtBody.Singleton())
}
func (n *SelectStmt) SetBody(node *BlockStmt) {
	psi.UpdateEdge(n, EdgeKindSelectStmtBody.Singleton(), node)
}
func NewFromSelectStmt(fset *token.FileSet, node *ast.SelectStmt) *SelectStmt {
	n := &SelectStmt{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(SelectStmtType))

	return n
}

func (n *SelectStmt) CopyFromGoAst(fset *token.FileSet, src *ast.SelectStmt) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Body != nil {
		tmpBody := NewFromBlockStmt(fset, src.Body)
		tmpBody.SetParent(n)
		psi.UpdateEdge(n, EdgeKindSelectStmtBody.Singleton(), tmpBody)
	}

}

func (n *SelectStmt) ToGoAst() ast.Node { return n.ToGoSelectStmt(nil) }

func (n *SelectStmt) ToGoSelectStmt(dst *ast.SelectStmt) *ast.SelectStmt {
	if dst == nil {
		dst = &ast.SelectStmt{}
	}
	tmpBody := psi.GetEdgeOrNil[*BlockStmt](n, EdgeKindSelectStmtBody.Singleton())
	if tmpBody != nil {
		dst.Body = tmpBody.ToGoAst().(*ast.BlockStmt)
	}

	return dst
}

type ForStmt struct {
	NodeBase
}

var ForStmtType = psi.DefineNodeType[*ForStmt]()

var EdgeKindForStmtInit = psi.DefineEdgeType[Stmt]("GoForStmtInit")
var EdgeKindForStmtCond = psi.DefineEdgeType[Expr]("GoForStmtCond")
var EdgeKindForStmtPost = psi.DefineEdgeType[Stmt]("GoForStmtPost")
var EdgeKindForStmtBody = psi.DefineEdgeType[*BlockStmt]("GoForStmtBody")

func (n *ForStmt) GetInit() Stmt     { return psi.GetEdgeOrNil[Stmt](n, EdgeKindForStmtInit.Singleton()) }
func (n *ForStmt) SetInit(node Stmt) { psi.UpdateEdge(n, EdgeKindForStmtInit.Singleton(), node) }
func (n *ForStmt) GetCond() Expr     { return psi.GetEdgeOrNil[Expr](n, EdgeKindForStmtCond.Singleton()) }
func (n *ForStmt) SetCond(node Expr) { psi.UpdateEdge(n, EdgeKindForStmtCond.Singleton(), node) }
func (n *ForStmt) GetPost() Stmt     { return psi.GetEdgeOrNil[Stmt](n, EdgeKindForStmtPost.Singleton()) }
func (n *ForStmt) SetPost(node Stmt) { psi.UpdateEdge(n, EdgeKindForStmtPost.Singleton(), node) }
func (n *ForStmt) GetBody() *BlockStmt {
	return psi.GetEdgeOrNil[*BlockStmt](n, EdgeKindForStmtBody.Singleton())
}
func (n *ForStmt) SetBody(node *BlockStmt) { psi.UpdateEdge(n, EdgeKindForStmtBody.Singleton(), node) }
func NewFromForStmt(fset *token.FileSet, node *ast.ForStmt) *ForStmt {
	n := &ForStmt{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(ForStmtType))

	return n
}

func (n *ForStmt) CopyFromGoAst(fset *token.FileSet, src *ast.ForStmt) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Init != nil {
		tmpInit := NewFromStmt(fset, src.Init)
		tmpInit.SetParent(n)
		psi.UpdateEdge(n, EdgeKindForStmtInit.Singleton(), tmpInit)
	}

	if src.Cond != nil {
		tmpCond := NewFromExpr(fset, src.Cond)
		tmpCond.SetParent(n)
		psi.UpdateEdge(n, EdgeKindForStmtCond.Singleton(), tmpCond)
	}

	if src.Post != nil {
		tmpPost := NewFromStmt(fset, src.Post)
		tmpPost.SetParent(n)
		psi.UpdateEdge(n, EdgeKindForStmtPost.Singleton(), tmpPost)
	}

	if src.Body != nil {
		tmpBody := NewFromBlockStmt(fset, src.Body)
		tmpBody.SetParent(n)
		psi.UpdateEdge(n, EdgeKindForStmtBody.Singleton(), tmpBody)
	}

}

func (n *ForStmt) ToGoAst() ast.Node { return n.ToGoForStmt(nil) }

func (n *ForStmt) ToGoForStmt(dst *ast.ForStmt) *ast.ForStmt {
	if dst == nil {
		dst = &ast.ForStmt{}
	}
	tmpInit := psi.GetEdgeOrNil[Stmt](n, EdgeKindForStmtInit.Singleton())
	if tmpInit != nil {
		dst.Init = tmpInit.ToGoAst().(ast.Stmt)
	}

	tmpCond := psi.GetEdgeOrNil[Expr](n, EdgeKindForStmtCond.Singleton())
	if tmpCond != nil {
		dst.Cond = tmpCond.ToGoAst().(ast.Expr)
	}

	tmpPost := psi.GetEdgeOrNil[Stmt](n, EdgeKindForStmtPost.Singleton())
	if tmpPost != nil {
		dst.Post = tmpPost.ToGoAst().(ast.Stmt)
	}

	tmpBody := psi.GetEdgeOrNil[*BlockStmt](n, EdgeKindForStmtBody.Singleton())
	if tmpBody != nil {
		dst.Body = tmpBody.ToGoAst().(*ast.BlockStmt)
	}

	return dst
}

type RangeStmt struct {
	NodeBase
	Tok token.Token `json:"Tok"`
}

var RangeStmtType = psi.DefineNodeType[*RangeStmt]()

var EdgeKindRangeStmtKey = psi.DefineEdgeType[Expr]("GoRangeStmtKey")
var EdgeKindRangeStmtX = psi.DefineEdgeType[Expr]("GoRangeStmtX")
var EdgeKindRangeStmtBody = psi.DefineEdgeType[*BlockStmt]("GoRangeStmtBody")

func (n *RangeStmt) GetKey() Expr     { return psi.GetEdgeOrNil[Expr](n, EdgeKindRangeStmtKey.Singleton()) }
func (n *RangeStmt) SetKey(node Expr) { psi.UpdateEdge(n, EdgeKindRangeStmtKey.Singleton(), node) }
func (n *RangeStmt) GetX() Expr       { return psi.GetEdgeOrNil[Expr](n, EdgeKindRangeStmtX.Singleton()) }
func (n *RangeStmt) SetX(node Expr)   { psi.UpdateEdge(n, EdgeKindRangeStmtX.Singleton(), node) }
func (n *RangeStmt) GetBody() *BlockStmt {
	return psi.GetEdgeOrNil[*BlockStmt](n, EdgeKindRangeStmtBody.Singleton())
}
func (n *RangeStmt) SetBody(node *BlockStmt) {
	psi.UpdateEdge(n, EdgeKindRangeStmtBody.Singleton(), node)
}
func NewFromRangeStmt(fset *token.FileSet, node *ast.RangeStmt) *RangeStmt {
	n := &RangeStmt{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(RangeStmtType))

	return n
}

func (n *RangeStmt) CopyFromGoAst(fset *token.FileSet, src *ast.RangeStmt) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	n.Tok = src.Tok
	if src.Key != nil {
		tmpKey := NewFromExpr(fset, src.Key)
		tmpKey.SetParent(n)
		psi.UpdateEdge(n, EdgeKindRangeStmtKey.Singleton(), tmpKey)
	}

	if src.X != nil {
		tmpX := NewFromExpr(fset, src.X)
		tmpX.SetParent(n)
		psi.UpdateEdge(n, EdgeKindRangeStmtX.Singleton(), tmpX)
	}

	if src.Body != nil {
		tmpBody := NewFromBlockStmt(fset, src.Body)
		tmpBody.SetParent(n)
		psi.UpdateEdge(n, EdgeKindRangeStmtBody.Singleton(), tmpBody)
	}

}

func (n *RangeStmt) ToGoAst() ast.Node { return n.ToGoRangeStmt(nil) }

func (n *RangeStmt) ToGoRangeStmt(dst *ast.RangeStmt) *ast.RangeStmt {
	if dst == nil {
		dst = &ast.RangeStmt{}
	}
	dst.Tok = n.Tok
	tmpKey := psi.GetEdgeOrNil[Expr](n, EdgeKindRangeStmtKey.Singleton())
	if tmpKey != nil {
		dst.Key = tmpKey.ToGoAst().(ast.Expr)
	}

	tmpX := psi.GetEdgeOrNil[Expr](n, EdgeKindRangeStmtX.Singleton())
	if tmpX != nil {
		dst.X = tmpX.ToGoAst().(ast.Expr)
	}

	tmpBody := psi.GetEdgeOrNil[*BlockStmt](n, EdgeKindRangeStmtBody.Singleton())
	if tmpBody != nil {
		dst.Body = tmpBody.ToGoAst().(*ast.BlockStmt)
	}

	return dst
}

type Spec interface {
	Node
}

func NewFromSpec(fset *token.FileSet, node ast.Spec) Spec {
	return GoAstToPsi(fset, node).(Spec)
}

type ImportSpec struct {
	NodeBase
}

var ImportSpecType = psi.DefineNodeType[*ImportSpec]()

var EdgeKindImportSpecDoc = psi.DefineEdgeType[*CommentGroup]("GoImportSpecDoc")
var EdgeKindImportSpecName = psi.DefineEdgeType[*Ident]("GoImportSpecName")
var EdgeKindImportSpecPath = psi.DefineEdgeType[*BasicLit]("GoImportSpecPath")
var EdgeKindImportSpecComment = psi.DefineEdgeType[*CommentGroup]("GoImportSpecComment")

func (n *ImportSpec) GetDoc() *CommentGroup {
	return psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindImportSpecDoc.Singleton())
}
func (n *ImportSpec) SetDoc(node *CommentGroup) {
	psi.UpdateEdge(n, EdgeKindImportSpecDoc.Singleton(), node)
}
func (n *ImportSpec) GetName() *Ident {
	return psi.GetEdgeOrNil[*Ident](n, EdgeKindImportSpecName.Singleton())
}
func (n *ImportSpec) SetName(node *Ident) {
	psi.UpdateEdge(n, EdgeKindImportSpecName.Singleton(), node)
}
func (n *ImportSpec) GetPath() *BasicLit {
	return psi.GetEdgeOrNil[*BasicLit](n, EdgeKindImportSpecPath.Singleton())
}
func (n *ImportSpec) SetPath(node *BasicLit) {
	psi.UpdateEdge(n, EdgeKindImportSpecPath.Singleton(), node)
}
func (n *ImportSpec) GetComment() *CommentGroup {
	return psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindImportSpecComment.Singleton())
}
func (n *ImportSpec) SetComment(node *CommentGroup) {
	psi.UpdateEdge(n, EdgeKindImportSpecComment.Singleton(), node)
}
func NewFromImportSpec(fset *token.FileSet, node *ast.ImportSpec) *ImportSpec {
	n := &ImportSpec{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(ImportSpecType))

	return n
}

func (n *ImportSpec) CopyFromGoAst(fset *token.FileSet, src *ast.ImportSpec) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Doc != nil {
		tmpDoc := NewFromCommentGroup(fset, src.Doc)
		tmpDoc.SetParent(n)
		psi.UpdateEdge(n, EdgeKindImportSpecDoc.Singleton(), tmpDoc)
	}

	if src.Name != nil {
		tmpName := NewFromIdent(fset, src.Name)
		tmpName.SetParent(n)
		psi.UpdateEdge(n, EdgeKindImportSpecName.Singleton(), tmpName)
	}

	if src.Path != nil {
		tmpPath := NewFromBasicLit(fset, src.Path)
		tmpPath.SetParent(n)
		psi.UpdateEdge(n, EdgeKindImportSpecPath.Singleton(), tmpPath)
	}

	if src.Comment != nil {
		tmpComment := NewFromCommentGroup(fset, src.Comment)
		tmpComment.SetParent(n)
		psi.UpdateEdge(n, EdgeKindImportSpecComment.Singleton(), tmpComment)
	}

}

func (n *ImportSpec) ToGoAst() ast.Node { return n.ToGoImportSpec(nil) }

func (n *ImportSpec) ToGoImportSpec(dst *ast.ImportSpec) *ast.ImportSpec {
	if dst == nil {
		dst = &ast.ImportSpec{}
	}
	tmpDoc := psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindImportSpecDoc.Singleton())
	if tmpDoc != nil {
		dst.Doc = tmpDoc.ToGoAst().(*ast.CommentGroup)
	}

	tmpName := psi.GetEdgeOrNil[*Ident](n, EdgeKindImportSpecName.Singleton())
	if tmpName != nil {
		dst.Name = tmpName.ToGoAst().(*ast.Ident)
	}

	tmpPath := psi.GetEdgeOrNil[*BasicLit](n, EdgeKindImportSpecPath.Singleton())
	if tmpPath != nil {
		dst.Path = tmpPath.ToGoAst().(*ast.BasicLit)
	}

	tmpComment := psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindImportSpecComment.Singleton())
	if tmpComment != nil {
		dst.Comment = tmpComment.ToGoAst().(*ast.CommentGroup)
	}

	return dst
}

type ValueSpec struct {
	NodeBase
}

var ValueSpecType = psi.DefineNodeType[*ValueSpec]()

var EdgeKindValueSpecDoc = psi.DefineEdgeType[*CommentGroup]("GoValueSpecDoc")
var EdgeKindValueSpecNames = psi.DefineEdgeType[*Ident]("GoValueSpecNames")
var EdgeKindValueSpecType = psi.DefineEdgeType[Expr]("GoValueSpecType")
var EdgeKindValueSpecValues = psi.DefineEdgeType[Expr]("GoValueSpecValues")
var EdgeKindValueSpecComment = psi.DefineEdgeType[*CommentGroup]("GoValueSpecComment")

func (n *ValueSpec) GetDoc() *CommentGroup {
	return psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindValueSpecDoc.Singleton())
}
func (n *ValueSpec) SetDoc(node *CommentGroup) {
	psi.UpdateEdge(n, EdgeKindValueSpecDoc.Singleton(), node)
}
func (n *ValueSpec) GetNames() []*Ident      { return psi.GetEdges(n, EdgeKindValueSpecNames) }
func (n *ValueSpec) SetNames(nodes []*Ident) { psi.UpdateEdges(n, EdgeKindValueSpecNames, nodes) }
func (n *ValueSpec) GetType() Expr {
	return psi.GetEdgeOrNil[Expr](n, EdgeKindValueSpecType.Singleton())
}
func (n *ValueSpec) SetType(node Expr)      { psi.UpdateEdge(n, EdgeKindValueSpecType.Singleton(), node) }
func (n *ValueSpec) GetValues() []Expr      { return psi.GetEdges(n, EdgeKindValueSpecValues) }
func (n *ValueSpec) SetValues(nodes []Expr) { psi.UpdateEdges(n, EdgeKindValueSpecValues, nodes) }
func (n *ValueSpec) GetComment() *CommentGroup {
	return psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindValueSpecComment.Singleton())
}
func (n *ValueSpec) SetComment(node *CommentGroup) {
	psi.UpdateEdge(n, EdgeKindValueSpecComment.Singleton(), node)
}
func NewFromValueSpec(fset *token.FileSet, node *ast.ValueSpec) *ValueSpec {
	n := &ValueSpec{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(ValueSpecType))

	return n
}

func (n *ValueSpec) CopyFromGoAst(fset *token.FileSet, src *ast.ValueSpec) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Doc != nil {
		tmpDoc := NewFromCommentGroup(fset, src.Doc)
		tmpDoc.SetParent(n)
		psi.UpdateEdge(n, EdgeKindValueSpecDoc.Singleton(), tmpDoc)
	}

	for i, v := range src.Names {
		tmpNames := NewFromIdent(fset, v)
		tmpNames.SetParent(n)
		psi.UpdateEdge(n, EdgeKindValueSpecNames.Indexed(int64(i)), tmpNames)
	}

	if src.Type != nil {
		tmpType := NewFromExpr(fset, src.Type)
		tmpType.SetParent(n)
		psi.UpdateEdge(n, EdgeKindValueSpecType.Singleton(), tmpType)
	}

	for i, v := range src.Values {
		tmpValues := NewFromExpr(fset, v)
		tmpValues.SetParent(n)
		psi.UpdateEdge(n, EdgeKindValueSpecValues.Indexed(int64(i)), tmpValues)
	}

	if src.Comment != nil {
		tmpComment := NewFromCommentGroup(fset, src.Comment)
		tmpComment.SetParent(n)
		psi.UpdateEdge(n, EdgeKindValueSpecComment.Singleton(), tmpComment)
	}

}

func (n *ValueSpec) ToGoAst() ast.Node { return n.ToGoValueSpec(nil) }

func (n *ValueSpec) ToGoValueSpec(dst *ast.ValueSpec) *ast.ValueSpec {
	if dst == nil {
		dst = &ast.ValueSpec{}
	}
	tmpDoc := psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindValueSpecDoc.Singleton())
	if tmpDoc != nil {
		dst.Doc = tmpDoc.ToGoAst().(*ast.CommentGroup)
	}

	tmpNames := psi.GetEdges(n, EdgeKindValueSpecNames)
	dst.Names = make([]*ast.Ident, len(tmpNames))
	for i, v := range tmpNames {
		dst.Names[i] = v.ToGoAst().(*ast.Ident)
	}

	tmpType := psi.GetEdgeOrNil[Expr](n, EdgeKindValueSpecType.Singleton())
	if tmpType != nil {
		dst.Type = tmpType.ToGoAst().(ast.Expr)
	}

	tmpValues := psi.GetEdges(n, EdgeKindValueSpecValues)
	dst.Values = make([]ast.Expr, len(tmpValues))
	for i, v := range tmpValues {
		dst.Values[i] = v.ToGoAst().(ast.Expr)
	}

	tmpComment := psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindValueSpecComment.Singleton())
	if tmpComment != nil {
		dst.Comment = tmpComment.ToGoAst().(*ast.CommentGroup)
	}

	return dst
}

type TypeSpec struct {
	NodeBase
}

var TypeSpecType = psi.DefineNodeType[*TypeSpec]()

var EdgeKindTypeSpecDoc = psi.DefineEdgeType[*CommentGroup]("GoTypeSpecDoc")
var EdgeKindTypeSpecName = psi.DefineEdgeType[*Ident]("GoTypeSpecName")
var EdgeKindTypeSpecTypeParams = psi.DefineEdgeType[*FieldList]("GoTypeSpecTypeParams")
var EdgeKindTypeSpecType = psi.DefineEdgeType[Expr]("GoTypeSpecType")
var EdgeKindTypeSpecComment = psi.DefineEdgeType[*CommentGroup]("GoTypeSpecComment")

func (n *TypeSpec) GetDoc() *CommentGroup {
	return psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindTypeSpecDoc.Singleton())
}
func (n *TypeSpec) SetDoc(node *CommentGroup) {
	psi.UpdateEdge(n, EdgeKindTypeSpecDoc.Singleton(), node)
}
func (n *TypeSpec) GetName() *Ident {
	return psi.GetEdgeOrNil[*Ident](n, EdgeKindTypeSpecName.Singleton())
}
func (n *TypeSpec) SetName(node *Ident) { psi.UpdateEdge(n, EdgeKindTypeSpecName.Singleton(), node) }
func (n *TypeSpec) GetTypeParams() *FieldList {
	return psi.GetEdgeOrNil[*FieldList](n, EdgeKindTypeSpecTypeParams.Singleton())
}
func (n *TypeSpec) SetTypeParams(node *FieldList) {
	psi.UpdateEdge(n, EdgeKindTypeSpecTypeParams.Singleton(), node)
}
func (n *TypeSpec) GetType() Expr     { return psi.GetEdgeOrNil[Expr](n, EdgeKindTypeSpecType.Singleton()) }
func (n *TypeSpec) SetType(node Expr) { psi.UpdateEdge(n, EdgeKindTypeSpecType.Singleton(), node) }
func (n *TypeSpec) GetComment() *CommentGroup {
	return psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindTypeSpecComment.Singleton())
}
func (n *TypeSpec) SetComment(node *CommentGroup) {
	psi.UpdateEdge(n, EdgeKindTypeSpecComment.Singleton(), node)
}
func NewFromTypeSpec(fset *token.FileSet, node *ast.TypeSpec) *TypeSpec {
	n := &TypeSpec{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(TypeSpecType))

	return n
}

func (n *TypeSpec) CopyFromGoAst(fset *token.FileSet, src *ast.TypeSpec) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Doc != nil {
		tmpDoc := NewFromCommentGroup(fset, src.Doc)
		tmpDoc.SetParent(n)
		psi.UpdateEdge(n, EdgeKindTypeSpecDoc.Singleton(), tmpDoc)
	}

	if src.Name != nil {
		tmpName := NewFromIdent(fset, src.Name)
		tmpName.SetParent(n)
		psi.UpdateEdge(n, EdgeKindTypeSpecName.Singleton(), tmpName)
	}

	if src.TypeParams != nil {
		tmpTypeParams := NewFromFieldList(fset, src.TypeParams)
		tmpTypeParams.SetParent(n)
		psi.UpdateEdge(n, EdgeKindTypeSpecTypeParams.Singleton(), tmpTypeParams)
	}

	if src.Type != nil {
		tmpType := NewFromExpr(fset, src.Type)
		tmpType.SetParent(n)
		psi.UpdateEdge(n, EdgeKindTypeSpecType.Singleton(), tmpType)
	}

	if src.Comment != nil {
		tmpComment := NewFromCommentGroup(fset, src.Comment)
		tmpComment.SetParent(n)
		psi.UpdateEdge(n, EdgeKindTypeSpecComment.Singleton(), tmpComment)
	}

}

func (n *TypeSpec) ToGoAst() ast.Node { return n.ToGoTypeSpec(nil) }

func (n *TypeSpec) ToGoTypeSpec(dst *ast.TypeSpec) *ast.TypeSpec {
	if dst == nil {
		dst = &ast.TypeSpec{}
	}
	tmpDoc := psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindTypeSpecDoc.Singleton())
	if tmpDoc != nil {
		dst.Doc = tmpDoc.ToGoAst().(*ast.CommentGroup)
	}

	tmpName := psi.GetEdgeOrNil[*Ident](n, EdgeKindTypeSpecName.Singleton())
	if tmpName != nil {
		dst.Name = tmpName.ToGoAst().(*ast.Ident)
	}

	tmpTypeParams := psi.GetEdgeOrNil[*FieldList](n, EdgeKindTypeSpecTypeParams.Singleton())
	if tmpTypeParams != nil {
		dst.TypeParams = tmpTypeParams.ToGoAst().(*ast.FieldList)
	}

	tmpType := psi.GetEdgeOrNil[Expr](n, EdgeKindTypeSpecType.Singleton())
	if tmpType != nil {
		dst.Type = tmpType.ToGoAst().(ast.Expr)
	}

	tmpComment := psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindTypeSpecComment.Singleton())
	if tmpComment != nil {
		dst.Comment = tmpComment.ToGoAst().(*ast.CommentGroup)
	}

	return dst
}

type BadDecl struct {
	NodeBase
}

var BadDeclType = psi.DefineNodeType[*BadDecl]()

func NewFromBadDecl(fset *token.FileSet, node *ast.BadDecl) *BadDecl {
	n := &BadDecl{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(BadDeclType))

	return n
}

func (n *BadDecl) CopyFromGoAst(fset *token.FileSet, src *ast.BadDecl) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
}

func (n *BadDecl) ToGoAst() ast.Node { return n.ToGoBadDecl(nil) }

func (n *BadDecl) ToGoBadDecl(dst *ast.BadDecl) *ast.BadDecl {
	if dst == nil {
		dst = &ast.BadDecl{}
	}

	return dst
}

type GenDecl struct {
	NodeBase
	Tok token.Token `json:"Tok"`
}

var GenDeclType = psi.DefineNodeType[*GenDecl]()

var EdgeKindGenDeclDoc = psi.DefineEdgeType[*CommentGroup]("GoGenDeclDoc")
var EdgeKindGenDeclSpecs = psi.DefineEdgeType[Spec]("GoGenDeclSpecs")

func (n *GenDecl) GetDoc() *CommentGroup {
	return psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindGenDeclDoc.Singleton())
}
func (n *GenDecl) SetDoc(node *CommentGroup) { psi.UpdateEdge(n, EdgeKindGenDeclDoc.Singleton(), node) }
func (n *GenDecl) GetSpecs() []Spec          { return psi.GetEdges(n, EdgeKindGenDeclSpecs) }
func (n *GenDecl) SetSpecs(nodes []Spec)     { psi.UpdateEdges(n, EdgeKindGenDeclSpecs, nodes) }
func NewFromGenDecl(fset *token.FileSet, node *ast.GenDecl) *GenDecl {
	n := &GenDecl{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(GenDeclType))

	return n
}

func (n *GenDecl) CopyFromGoAst(fset *token.FileSet, src *ast.GenDecl) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	n.Tok = src.Tok
	if src.Doc != nil {
		tmpDoc := NewFromCommentGroup(fset, src.Doc)
		tmpDoc.SetParent(n)
		psi.UpdateEdge(n, EdgeKindGenDeclDoc.Singleton(), tmpDoc)
	}

	for i, v := range src.Specs {
		tmpSpecs := NewFromSpec(fset, v)
		tmpSpecs.SetParent(n)
		psi.UpdateEdge(n, EdgeKindGenDeclSpecs.Indexed(int64(i)), tmpSpecs)
	}

}

func (n *GenDecl) ToGoAst() ast.Node { return n.ToGoGenDecl(nil) }

func (n *GenDecl) ToGoGenDecl(dst *ast.GenDecl) *ast.GenDecl {
	if dst == nil {
		dst = &ast.GenDecl{}
	}
	dst.Tok = n.Tok
	tmpDoc := psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindGenDeclDoc.Singleton())
	if tmpDoc != nil {
		dst.Doc = tmpDoc.ToGoAst().(*ast.CommentGroup)
	}

	tmpSpecs := psi.GetEdges(n, EdgeKindGenDeclSpecs)
	dst.Specs = make([]ast.Spec, len(tmpSpecs))
	for i, v := range tmpSpecs {
		dst.Specs[i] = v.ToGoAst().(ast.Spec)
	}

	return dst
}

type FuncDecl struct {
	NodeBase
}

var FuncDeclType = psi.DefineNodeType[*FuncDecl]()

var EdgeKindFuncDeclDoc = psi.DefineEdgeType[*CommentGroup]("GoFuncDeclDoc")
var EdgeKindFuncDeclRecv = psi.DefineEdgeType[*FieldList]("GoFuncDeclRecv")
var EdgeKindFuncDeclName = psi.DefineEdgeType[*Ident]("GoFuncDeclName")
var EdgeKindFuncDeclType = psi.DefineEdgeType[*FuncType]("GoFuncDeclType")
var EdgeKindFuncDeclBody = psi.DefineEdgeType[*BlockStmt]("GoFuncDeclBody")

func (n *FuncDecl) GetDoc() *CommentGroup {
	return psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindFuncDeclDoc.Singleton())
}
func (n *FuncDecl) SetDoc(node *CommentGroup) {
	psi.UpdateEdge(n, EdgeKindFuncDeclDoc.Singleton(), node)
}
func (n *FuncDecl) GetRecv() *FieldList {
	return psi.GetEdgeOrNil[*FieldList](n, EdgeKindFuncDeclRecv.Singleton())
}
func (n *FuncDecl) SetRecv(node *FieldList) {
	psi.UpdateEdge(n, EdgeKindFuncDeclRecv.Singleton(), node)
}
func (n *FuncDecl) GetName() *Ident {
	return psi.GetEdgeOrNil[*Ident](n, EdgeKindFuncDeclName.Singleton())
}
func (n *FuncDecl) SetName(node *Ident) { psi.UpdateEdge(n, EdgeKindFuncDeclName.Singleton(), node) }
func (n *FuncDecl) GetType() *FuncType {
	return psi.GetEdgeOrNil[*FuncType](n, EdgeKindFuncDeclType.Singleton())
}
func (n *FuncDecl) SetType(node *FuncType) { psi.UpdateEdge(n, EdgeKindFuncDeclType.Singleton(), node) }
func (n *FuncDecl) GetBody() *BlockStmt {
	return psi.GetEdgeOrNil[*BlockStmt](n, EdgeKindFuncDeclBody.Singleton())
}
func (n *FuncDecl) SetBody(node *BlockStmt) {
	psi.UpdateEdge(n, EdgeKindFuncDeclBody.Singleton(), node)
}
func NewFromFuncDecl(fset *token.FileSet, node *ast.FuncDecl) *FuncDecl {
	n := &FuncDecl{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(FuncDeclType))

	return n
}

func (n *FuncDecl) CopyFromGoAst(fset *token.FileSet, src *ast.FuncDecl) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Doc != nil {
		tmpDoc := NewFromCommentGroup(fset, src.Doc)
		tmpDoc.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFuncDeclDoc.Singleton(), tmpDoc)
	}

	if src.Recv != nil {
		tmpRecv := NewFromFieldList(fset, src.Recv)
		tmpRecv.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFuncDeclRecv.Singleton(), tmpRecv)
	}

	if src.Name != nil {
		tmpName := NewFromIdent(fset, src.Name)
		tmpName.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFuncDeclName.Singleton(), tmpName)
	}

	if src.Type != nil {
		tmpType := NewFromFuncType(fset, src.Type)
		tmpType.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFuncDeclType.Singleton(), tmpType)
	}

	if src.Body != nil {
		tmpBody := NewFromBlockStmt(fset, src.Body)
		tmpBody.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFuncDeclBody.Singleton(), tmpBody)
	}

}

func (n *FuncDecl) ToGoAst() ast.Node { return n.ToGoFuncDecl(nil) }

func (n *FuncDecl) ToGoFuncDecl(dst *ast.FuncDecl) *ast.FuncDecl {
	if dst == nil {
		dst = &ast.FuncDecl{}
	}
	tmpDoc := psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindFuncDeclDoc.Singleton())
	if tmpDoc != nil {
		dst.Doc = tmpDoc.ToGoAst().(*ast.CommentGroup)
	}

	tmpRecv := psi.GetEdgeOrNil[*FieldList](n, EdgeKindFuncDeclRecv.Singleton())
	if tmpRecv != nil {
		dst.Recv = tmpRecv.ToGoAst().(*ast.FieldList)
	}

	tmpName := psi.GetEdgeOrNil[*Ident](n, EdgeKindFuncDeclName.Singleton())
	if tmpName != nil {
		dst.Name = tmpName.ToGoAst().(*ast.Ident)
	}

	tmpType := psi.GetEdgeOrNil[*FuncType](n, EdgeKindFuncDeclType.Singleton())
	if tmpType != nil {
		dst.Type = tmpType.ToGoAst().(*ast.FuncType)
	}

	tmpBody := psi.GetEdgeOrNil[*BlockStmt](n, EdgeKindFuncDeclBody.Singleton())
	if tmpBody != nil {
		dst.Body = tmpBody.ToGoAst().(*ast.BlockStmt)
	}

	return dst
}

type File struct {
	NodeBase
}

var FileType = psi.DefineNodeType[*File]()

var EdgeKindFileDoc = psi.DefineEdgeType[*CommentGroup]("GoFileDoc")
var EdgeKindFileName = psi.DefineEdgeType[*Ident]("GoFileName")
var EdgeKindFileDecls = psi.DefineEdgeType[Decl]("GoFileDecls")
var EdgeKindFileImports = psi.DefineEdgeType[*ImportSpec]("GoFileImports")
var EdgeKindFileComments = psi.DefineEdgeType[*CommentGroup]("GoFileComments")

func (n *File) GetDoc() *CommentGroup {
	return psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindFileDoc.Singleton())
}
func (n *File) SetDoc(node *CommentGroup)         { psi.UpdateEdge(n, EdgeKindFileDoc.Singleton(), node) }
func (n *File) GetName() *Ident                   { return psi.GetEdgeOrNil[*Ident](n, EdgeKindFileName.Singleton()) }
func (n *File) SetName(node *Ident)               { psi.UpdateEdge(n, EdgeKindFileName.Singleton(), node) }
func (n *File) GetDecls() []Decl                  { return psi.GetEdges(n, EdgeKindFileDecls) }
func (n *File) SetDecls(nodes []Decl)             { psi.UpdateEdges(n, EdgeKindFileDecls, nodes) }
func (n *File) GetImports() []*ImportSpec         { return psi.GetEdges(n, EdgeKindFileImports) }
func (n *File) SetImports(nodes []*ImportSpec)    { psi.UpdateEdges(n, EdgeKindFileImports, nodes) }
func (n *File) GetComments() []*CommentGroup      { return psi.GetEdges(n, EdgeKindFileComments) }
func (n *File) SetComments(nodes []*CommentGroup) { psi.UpdateEdges(n, EdgeKindFileComments, nodes) }
func NewFromFile(fset *token.FileSet, node *ast.File) *File {
	n := &File{}

	n.CopyFromGoAst(fset, node)
	n.Init(n, psi.WithNodeType(FileType))

	return n
}

func (n *File) CopyFromGoAst(fset *token.FileSet, src *ast.File) {
	n.StartTokenPos = src.Pos()
	n.EndTokenPos = src.End()
	if src.Doc != nil {
		tmpDoc := NewFromCommentGroup(fset, src.Doc)
		tmpDoc.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFileDoc.Singleton(), tmpDoc)
	}

	if src.Name != nil {
		tmpName := NewFromIdent(fset, src.Name)
		tmpName.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFileName.Singleton(), tmpName)
	}

	for i, v := range src.Decls {
		tmpDecls := NewFromDecl(fset, v)
		tmpDecls.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFileDecls.Indexed(int64(i)), tmpDecls)
	}

	for i, v := range src.Imports {
		tmpImports := NewFromImportSpec(fset, v)
		tmpImports.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFileImports.Indexed(int64(i)), tmpImports)
	}

	for i, v := range src.Comments {
		tmpComments := NewFromCommentGroup(fset, v)
		tmpComments.SetParent(n)
		psi.UpdateEdge(n, EdgeKindFileComments.Indexed(int64(i)), tmpComments)
	}

}

func (n *File) ToGoAst() ast.Node { return n.ToGoFile(nil) }

func (n *File) ToGoFile(dst *ast.File) *ast.File {
	if dst == nil {
		dst = &ast.File{}
	}
	tmpDoc := psi.GetEdgeOrNil[*CommentGroup](n, EdgeKindFileDoc.Singleton())
	if tmpDoc != nil {
		dst.Doc = tmpDoc.ToGoAst().(*ast.CommentGroup)
	}

	tmpName := psi.GetEdgeOrNil[*Ident](n, EdgeKindFileName.Singleton())
	if tmpName != nil {
		dst.Name = tmpName.ToGoAst().(*ast.Ident)
	}

	tmpDecls := psi.GetEdges(n, EdgeKindFileDecls)
	dst.Decls = make([]ast.Decl, len(tmpDecls))
	for i, v := range tmpDecls {
		dst.Decls[i] = v.ToGoAst().(ast.Decl)
	}

	tmpImports := psi.GetEdges(n, EdgeKindFileImports)
	dst.Imports = make([]*ast.ImportSpec, len(tmpImports))
	for i, v := range tmpImports {
		dst.Imports[i] = v.ToGoAst().(*ast.ImportSpec)
	}

	tmpComments := psi.GetEdges(n, EdgeKindFileComments)
	dst.Comments = make([]*ast.CommentGroup, len(tmpComments))
	for i, v := range tmpComments {
		dst.Comments[i] = v.ToGoAst().(*ast.CommentGroup)
	}

	return dst
}

func GoAstToPsi(fset *token.FileSet, n ast.Node) Node {
	switch n := n.(type) {
	case *ast.Comment:
		return NewFromComment(fset, n)
	case *ast.CommentGroup:
		return NewFromCommentGroup(fset, n)
	case *ast.Field:
		return NewFromField(fset, n)
	case *ast.FieldList:
		return NewFromFieldList(fset, n)
	case *ast.BadExpr:
		return NewFromBadExpr(fset, n)
	case *ast.Ident:
		return NewFromIdent(fset, n)
	case *ast.Ellipsis:
		return NewFromEllipsis(fset, n)
	case *ast.BasicLit:
		return NewFromBasicLit(fset, n)
	case *ast.FuncLit:
		return NewFromFuncLit(fset, n)
	case *ast.CompositeLit:
		return NewFromCompositeLit(fset, n)
	case *ast.ParenExpr:
		return NewFromParenExpr(fset, n)
	case *ast.SelectorExpr:
		return NewFromSelectorExpr(fset, n)
	case *ast.IndexExpr:
		return NewFromIndexExpr(fset, n)
	case *ast.IndexListExpr:
		return NewFromIndexListExpr(fset, n)
	case *ast.SliceExpr:
		return NewFromSliceExpr(fset, n)
	case *ast.TypeAssertExpr:
		return NewFromTypeAssertExpr(fset, n)
	case *ast.CallExpr:
		return NewFromCallExpr(fset, n)
	case *ast.StarExpr:
		return NewFromStarExpr(fset, n)
	case *ast.UnaryExpr:
		return NewFromUnaryExpr(fset, n)
	case *ast.BinaryExpr:
		return NewFromBinaryExpr(fset, n)
	case *ast.KeyValueExpr:
		return NewFromKeyValueExpr(fset, n)
	case *ast.ArrayType:
		return NewFromArrayType(fset, n)
	case *ast.StructType:
		return NewFromStructType(fset, n)
	case *ast.FuncType:
		return NewFromFuncType(fset, n)
	case *ast.InterfaceType:
		return NewFromInterfaceType(fset, n)
	case *ast.MapType:
		return NewFromMapType(fset, n)
	case *ast.ChanType:
		return NewFromChanType(fset, n)
	case *ast.BadStmt:
		return NewFromBadStmt(fset, n)
	case *ast.DeclStmt:
		return NewFromDeclStmt(fset, n)
	case *ast.EmptyStmt:
		return NewFromEmptyStmt(fset, n)
	case *ast.LabeledStmt:
		return NewFromLabeledStmt(fset, n)
	case *ast.ExprStmt:
		return NewFromExprStmt(fset, n)
	case *ast.SendStmt:
		return NewFromSendStmt(fset, n)
	case *ast.IncDecStmt:
		return NewFromIncDecStmt(fset, n)
	case *ast.AssignStmt:
		return NewFromAssignStmt(fset, n)
	case *ast.GoStmt:
		return NewFromGoStmt(fset, n)
	case *ast.DeferStmt:
		return NewFromDeferStmt(fset, n)
	case *ast.ReturnStmt:
		return NewFromReturnStmt(fset, n)
	case *ast.BranchStmt:
		return NewFromBranchStmt(fset, n)
	case *ast.BlockStmt:
		return NewFromBlockStmt(fset, n)
	case *ast.IfStmt:
		return NewFromIfStmt(fset, n)
	case *ast.CaseClause:
		return NewFromCaseClause(fset, n)
	case *ast.SwitchStmt:
		return NewFromSwitchStmt(fset, n)
	case *ast.TypeSwitchStmt:
		return NewFromTypeSwitchStmt(fset, n)
	case *ast.CommClause:
		return NewFromCommClause(fset, n)
	case *ast.SelectStmt:
		return NewFromSelectStmt(fset, n)
	case *ast.ForStmt:
		return NewFromForStmt(fset, n)
	case *ast.RangeStmt:
		return NewFromRangeStmt(fset, n)
	case *ast.ImportSpec:
		return NewFromImportSpec(fset, n)
	case *ast.ValueSpec:
		return NewFromValueSpec(fset, n)
	case *ast.TypeSpec:
		return NewFromTypeSpec(fset, n)
	case *ast.BadDecl:
		return NewFromBadDecl(fset, n)
	case *ast.GenDecl:
		return NewFromGenDecl(fset, n)
	case *ast.FuncDecl:
		return NewFromFuncDecl(fset, n)
	case *ast.File:
		return NewFromFile(fset, n)
	default:
		panic(fmt.Errorf("unknown node type %T", n))
	}
}
