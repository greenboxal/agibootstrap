package reassembler

import (
	"os"

	"github.com/alecthomas/repr"
	"github.com/go-errors/errors"
	"github.com/ipld/go-ipld-prime"

	"github.com/greenboxal/agibootstrap/psidb/typesystem"
	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing"
	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing/gensparse"
	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing/jsonstream"
)

type ValueToken = sparsing.INodeToken[jsonstream.Value]

type ParserStream = gensparse.ParserStream[ValueToken, *ValueNode]
type ParserStreamContext = gensparse.StreamingParserContext[ValueToken, *ValueNode]

type ValueNode struct {
	sparsing.TerminalNodeBase

	Type  typesystem.Type
	Value typesystem.Value
}

type Reassembler struct {
	ParserStream

	rootType typesystem.Type

	builderStack []ipld.NodeBuilder
	pathStack    []sparsing.Node
	typeStack    []typesystem.Type
}

func NewReassembler(rootType typesystem.Type) *Reassembler {
	r := &Reassembler{
		rootType: rootType,
	}

	r.ParserStream = gensparse.NewParserStream[ValueToken, *ValueNode]()
	r.PushTokenParser(r)

	return r
}

func (r *Reassembler) ConsumeTokenStream(ctx ParserStreamContext) error {
	for ctx.RemainingTokens() > 0 {
		switch any(ctx.CurrentNode()).(type) {
		case *ValueNode:
			if err := r.ConsumeValue(ctx); err != nil {
				return err
			}
		default:
			return errors.Errorf("unexpected node type: %T", ctx.CurrentNode())
		}
	}

	return nil
}

func (r *Reassembler) getTypeForPath(p []sparsing.Node) (typesystem.Type, error) {
	typ := r.rootType

	for i := len(p) - 1; i >= 0; i-- {
		for j := len(r.pathStack); j >= 0; j-- {
			if p[i] == r.pathStack[j] {
				return r.typeStack[j], nil
			}
		}
	}

	for i, seg := range p {
		if i == 0 {
			continue
		}

		switch seg := seg.(type) {
		case *jsonstream.Pair:
			if seg.Key == nil {
				return nil, nil
			}

			if typ.PrimitiveKind() == typesystem.PrimitiveKindStruct {
				st := typ.Struct()
				fld := st.Field(seg.Key.Value)

				if fld == nil {
					return nil, errors.Errorf("unknown field %q in struct %q", seg.Key.Value, typ.Name())
				}

				typ = fld.Type()
			} else if typ.PrimitiveKind() == typesystem.PrimitiveKindList {
				typ = typ.List().Elem()
			} else if typ.PrimitiveKind() == typesystem.PrimitiveKindMap {
				typ = typ.Map().Value()
			} else {
				return nil, errors.Errorf("unexpected type %q", typ.Name())
			}
		}
	}

	return typ, nil
}

func (r *Reassembler) getBuilderForPath(p []sparsing.Node) (ipld.NodeBuilder, error) {
	for i := len(p) - 1; i >= 0; i-- {
		for j := len(r.pathStack); j >= 0; j-- {
			if p[i] == r.pathStack[j] {
				if i == len(p)-1 {
					return r.builderStack[j], nil
				}

				panic("not implemented")
			}
		}
	}

	typ, err := r.getTypeForPath(p)

	if err != nil {
		return nil, err
	}

	builder := typ.IpldPrototype().NewBuilder()

	return builder, nil
}

func (r *Reassembler) ConsumeValue(ctx ParserStreamContext) error {
	tk := ctx.PeekToken(0)

	if tk.GetKind() == sparsing.TokenKindEOS {
		return nil
	}

	vtk, ok := tk.(ValueToken)

	if !ok {
		return errors.Errorf("unexpected token type: %T", tk)
	}

	n := vtk.GetNode()

	nb, err := r.getBuilderForPath(ctx.CurrentPath())

	if err != nil {
		return err
	}

	switch n.GetPrimitiveType() {
	case typesystem.PrimitiveKindInt:
		if err := nb.AssignInt(n.GetValue().(int64)); err != nil {
			return err
		}
	case typesystem.PrimitiveKindFloat:
		if err := nb.AssignFloat(n.GetValue().(float64)); err != nil {
			return err
		}
	case typesystem.PrimitiveKindString:
		if err := nb.AssignString(n.GetValue().(string)); err != nil {
			return err
		}
	case typesystem.PrimitiveKindBoolean:
		if err := nb.AssignBool(n.GetValue().(bool)); err != nil {
			return err
		}
	}

	return nil
}

var reprer = repr.New(os.Stdout, repr.NoIndent())
