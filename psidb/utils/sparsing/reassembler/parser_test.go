package reassembler

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/greenboxal/agibootstrap/psidb/typesystem"
	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing"
	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing/gensparse"
	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing/jsonstream"
)

type TestStruct struct {
	Int    int     `json:"int,omitempty"`
	Float  float64 `json:"float,omitempty"`
	String string  `json:"string,omitempty"`

	Nested *TestStruct `json:"nested,omitempty"`
}

var TestData1 = TestStruct{
	Int:    1,
	Float:  1.0,
	String: "hello",
	Nested: &TestStruct{
		Int:    2,
		Float:  2.0,
		String: "world",
	},
}

var TestData1JSON = must(json.Marshal(TestData1))

func TestParser(t *testing.T) {
	jsonParser := jsonstream.NewParser(func(path jsonstream.ParserPath, node gensparse.Node) error {
		return nil
	})

	rp := NewReassembler(typesystem.GetType[TestStruct]())

	nodeIndexCounter := 0
	jsonParser.PushNodeConsumer(sparsing.ParserNodeHandlerFunc(func(ctx sparsing.StreamingParserContext, node sparsing.Node) error {
		v, ok := node.(jsonstream.Value)

		if !ok {
			return nil
		}

		index := nodeIndexCounter
		nodeIndexCounter++

		return rp.WriteToken(&sparsing.NodeToken[jsonstream.Value]{
			Token: sparsing.Token{
				Kind:     -0x1000,
				Value:    "",
				Start:    v.GetStartToken().GetStart(),
				End:      v.GetEndToken().GetEnd(),
				Index:    index,
				UserData: nil,
			},

			Node: v,
			Path: ctx.CurrentPath(),
		})
	}))

	n, err := jsonParser.Write(TestData1JSON)

	require.NoError(t, err)
	require.Equal(t, len(TestData1JSON), n)

}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}

	return v
}
