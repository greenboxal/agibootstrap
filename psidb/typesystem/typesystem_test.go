package typesystem

import (
	"testing"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/jaswdr/faker"
	"github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/require"
)

type TestEmbeddedStruct struct {
	Meta1 string `json:"meta_1"`
}

type TestInlined string

func (t *TestInlined) UnmarshalText(text []byte) error {
	*t = TestInlined(text)
	return nil
}

func (t TestInlined) MarshalText() (text []byte, err error) {
	return []byte(t), nil
}

type TestInline struct {
	TestInlined `ipld:",inline"`
}

type TestNamedEmbeddedStruct struct {
	ID TestInline `json:"id"`

	TestEmbeddedStruct
}

type TestScalars struct {
	A bool
	B int
	C string

	Time     time.Time
	Duration time.Duration
}

type TestComplexScalars struct {
	Cid     cid.Cid
	CidLink cidlink.Link
}

type TestPointers struct {
	A *bool
	B *int
	C *string

	Time     *time.Time
	Duration *time.Duration
}

type TestNils struct {
	List []string
}

type TestLists struct {
	TestListScalar []string
	TestListStruct []TestScalars
}

type TestMaps struct {
	TestMapStringStruct map[string]TestScalars
	TestMapIntStruct    map[int]TestScalars
	TestMapUint64Struct map[uint64]TestScalars

	TestMapStringString map[string]string
	TestMapIntInt       map[int]int
	TestMapUint64Uint64 map[uint64]uint64
}

type TestStruct1 struct {
	TestNamedEmbeddedStruct `json:"metadata"`

	Nils    TestNils
	Scalars TestScalars
	Lists   TestLists
	Maps    TestMaps
	Complex TestComplexScalars

	TestPointersNull TestPointers
	TestPointers     TestPointers

	TestIfaceNode interface{}
}

var testHash multihash.Multihash
var testCid cid.Cid

func init() {
	var err error

	testHash, err = multihash.Sum([]byte("LOL"), multihash.SHA2_256, -1)

	if err != nil {
		panic(err)
	}

	testCid = cid.NewCidV1(cid.Raw, testHash)
}

func TestTS(t *testing.T) {
	f := faker.New()

	ifaceValue := TestScalars{}
	f.Struct().Fill(&ifaceValue)

	initialValue := TestStruct1{}

	initialValue.Maps.TestMapIntInt = map[int]int{0: 1, 4342: 435345}
	initialValue.Maps.TestMapStringString = map[string]string{"saddas": "", "": "osdkfods", "asofdasa": "adoad"}
	initialValue.Maps.TestMapUint64Uint64 = map[uint64]uint64{0: 1, 4342: 435345}
	initialValue.Maps.TestMapIntStruct = map[int]TestScalars{0: ifaceValue, 4342: ifaceValue}
	initialValue.Maps.TestMapStringStruct = map[string]TestScalars{"saddas": ifaceValue, "": ifaceValue, "asofdasa": ifaceValue}
	initialValue.Maps.TestMapUint64Struct = map[uint64]TestScalars{0: ifaceValue, 4342: ifaceValue}

	f.Struct().Fill(&initialValue)

	initialValue.Nils = TestNils{}
	initialValue.TestIfaceNode = ifaceValue
	initialValue.Complex.Cid = testCid
	initialValue.Complex.CidLink = cidlink.Link{Cid: testCid}

	typ := TypeOf(TestStruct1{})
	wrapped := Wrap(initialValue)

	require.NotNil(t, typ)
	require.NotNil(t, wrapped)

	iface, err := wrapped.LookupByString("TestIfaceNode")

	require.NoError(t, err)
	require.NotNil(t, iface)

	data, err := ipld.Encode(wrapped, dagjson.Encode)

	require.NoError(t, err)
	require.NotNil(t, data)

	node, err := ipld.Decode(data, dagjson.Decode)

	require.NoError(t, err)
	require.NotNil(t, node)

	nodeWithProto, err := ipld.DecodeUsingPrototype(data, dagjson.Decode, typ.IpldPrototype())

	require.NoError(t, err)
	require.NotNil(t, nodeWithProto)

	unwrapped := Unwrap(nodeWithProto)

	require.NotNil(t, unwrapped)
	require.EqualValues(t, initialValue, unwrapped)
}
