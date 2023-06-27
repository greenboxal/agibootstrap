package thoughtstream

import (
	"encoding/binary"
	"time"

	"github.com/ipfs/go-cid"
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

type Pointer struct {
	Parent    cid.Cid
	Previous  cid.Cid
	Timestamp time.Time
	Level     int64
	Clock     int64
}

func (p Pointer) Address() cid.Cid {
	buf := make([]byte, 8*3+p.Parent.ByteLen()+p.Previous.ByteLen())

	binary.BigEndian.PutUint64(buf[0:8], uint64(p.Level))
	binary.BigEndian.PutUint64(buf[8:16], uint64(p.Clock))
	binary.BigEndian.PutUint64(buf[16:24], uint64(p.Timestamp.UnixNano()))

	copy(buf[24:], p.Parent.Bytes())
	copy(buf[24+p.Parent.ByteLen():], p.Previous.Bytes())

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
	return p.Level == 0 && p.Clock == 0 && p.Timestamp.IsZero() && p.Parent == cid.Undef
}

func (p Pointer) IsHead() bool {
	return p.Level == -1 && p.Clock == -1 && p.Timestamp.UnixMilli() == -1
}

func (p Pointer) IsRoot() bool {
	return p.Level == 0 && p.Clock == 0 && p.Timestamp.IsZero() && p.Parent == cid.Undef
}

func (p Pointer) IsParentOf(other Pointer) bool {
	return p.Level == other.Level-1 && p.Clock < other.Clock && p.Timestamp.Sub(other.Timestamp) < 0 && other.Parent == p.Address()
}

func (p Pointer) IsChildOf(other Pointer) bool {
	return other.IsParentOf(p)
}

func (p Pointer) IsSiblingOf(other Pointer) bool {
	return p.Level == other.Level && p.Parent == other.Parent
}

func (p Pointer) Less(other Pointer) bool {
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

func (p Pointer) Next() Pointer {
	return Pointer{
		Parent:    p.Parent,
		Previous:  p.Address(),
		Timestamp: time.Now(),
		Level:     p.Level,
		Clock:     p.Clock + 1,
	}
}

func HEAD() Pointer {
	return Pointer{
		Level:     -1,
		Clock:     -1,
		Timestamp: time.UnixMilli(-1),
	}
}
