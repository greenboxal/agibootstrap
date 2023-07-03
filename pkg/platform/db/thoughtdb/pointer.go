package thoughtdb

import (
	"encoding/binary"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/multiformats/go-multihash"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Reference struct {
	Path psi.Path
}

type Interval struct {
	Start Pointer
	End   Pointer
}

func (i Interval) CompareTo(other Interval) int {
	if i.Start.CompareTo(other.Start) == 0 {
		return i.End.CompareTo(other.End)
	}

	return i.Start.CompareTo(other.Start)
}

type Pointer struct {
	Parent    ipld.Link `json:"parent"`
	Previous  ipld.Link `json:"previous"`
	Timestamp time.Time `json:"timestamp"`
	Level     int64     `json:"level"`
	Clock     int64     `json:"clock"`
}

func (p Pointer) Address() cid.Cid {
	var parent, previous cid.Cid

	if p.Parent != nil {
		parent = p.Parent.(cidlink.Link).Cid
	}

	if p.Previous != nil {
		previous = p.Previous.(cidlink.Link).Cid
	}

	buf := make([]byte, 8*3+parent.ByteLen()+previous.ByteLen())

	binary.BigEndian.PutUint64(buf[0:8], uint64(p.Level))
	binary.BigEndian.PutUint64(buf[8:16], uint64(p.Clock))
	binary.BigEndian.PutUint64(buf[16:24], uint64(p.Timestamp.UnixNano()))

	copy(buf[24:], parent.Bytes())
	copy(buf[24+parent.ByteLen():], previous.Bytes())

	mh, err := multihash.Sum(buf, multihash.SHA2_256, -1)

	if err != nil {
		panic(err)
	}

	return cid.NewCidV1(cid.Raw, mh)
}

func (p Pointer) String() string {
	return p.Address().String()
}

func (p Pointer) IsZero() bool {
	return p.Level == 0 && p.Clock == 0 && p.Timestamp.IsZero() && p.Parent == nil
}

func (p Pointer) IsHead() bool {
	return p.Level == -1 && p.Clock == -1 && p.Timestamp.UnixMilli() == -1
}

func (p Pointer) IsRoot() bool {
	return p.Level == 0 && p.Clock == 0 && p.Timestamp.IsZero() && p.Parent == nil
}

func (p Pointer) IsSiblingOf(other Pointer) bool {
	return p.Level == other.Level && p.Parent == other.Parent
}

func (p Pointer) Less(other Pointer) bool {
	if p.IsHead() {
		return false
	}

	if other.IsHead() {
		return true
	}

	if p.Level < other.Level {
		return true
	}

	if p.Level > other.Level {
		return false
	}

	if p.Clock < other.Clock {
		return true
	}

	if p.Clock > other.Clock {
		return false
	}

	return p.Timestamp.Sub(other.Timestamp) < 0
}

func (p Pointer) CompareTo(other Pointer) int {
	if p == other {
		return 0
	}

	if p.Level < other.Level {
		return -1
	}

	if p.Clock < other.Clock {
		return -1
	}

	if p.Timestamp.Before(other.Timestamp) {
		return -1
	}

	return 1
}

func (p Pointer) Next(previous ipld.Link) Pointer {
	return Pointer{
		Parent:    p.Parent,
		Previous:  previous,
		Timestamp: time.Now(),
		Level:     p.Level,
		Clock:     p.Clock + 1,
	}
}

func Head() Pointer {
	return Pointer{
		Level:     -1,
		Clock:     -1,
		Timestamp: time.UnixMilli(-1),
	}
}

func RootPointer() Pointer {
	return Pointer{
		Parent:    cidlink.Link{Cid: rootCid},
		Previous:  cidlink.Link{Cid: rootCid},
		Timestamp: time.UnixMilli(0),
		Level:     0,
		Clock:     0,
	}
}

var rootCid cid.Cid

func init() {
	mh, err := multihash.Sum(nil, multihash.SHA2_256, -1)

	if err != nil {
		panic(err)
	}

	rootCid = cid.NewCidV1(cid.Raw, mh)
}
