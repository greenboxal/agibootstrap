package graphstore

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/tidwall/wal"
)

type WalOp uint8

const RecordLength = 64

const (
	WalOpUpdateInvalid = 0
	WalOpUpdateNode    = 1
	WalOpUpdateEdge    = 2
	WalOpRemoveNode    = 3
	WalOpRemoveEdge    = 4
	WalOpFence         = 5
)

type WalRecord struct {
	HdrLen  uint8
	Op      WalOp
	Counter uint64
	Ts      int64
	Len     uint8
	Payload *cid.Cid
}

func BuildWalRecord(op WalOp, payload cid.Cid) WalRecord {
	p := &payload

	if payload == cid.Undef {
		p = nil
	}

	return WalRecord{
		HdrLen:  20,
		Ts:      time.Now().UnixNano(),
		Op:      op,
		Len:     uint8(payload.ByteLen()),
		Payload: p,
	}
}

func (r *WalRecord) MarshalBinary() ([]byte, error) {
	if r.Payload != nil && r.Payload.ByteLen() != 37 && r.Payload.ByteLen() != 0 {
		return nil, fmt.Errorf("payload must be 37 bytes")
	}

	data := make([]byte, RecordLength)

	data[0] = r.HdrLen
	data[1] = byte(r.Op)

	binary.BigEndian.PutUint64(data[2:10], r.Counter)
	binary.BigEndian.PutUint64(data[10:18], uint64(r.Ts))
	data[18] = r.Len

	if r.Payload != nil {
		copy(data[r.HdrLen:r.HdrLen+r.Len], r.Payload.Bytes())
	}

	return data, nil
}

func (r *WalRecord) UnmarshalBinary(data []byte) error {
	if len(data) != RecordLength {
		return fmt.Errorf("data must be %d bytes", RecordLength)
	}

	r.HdrLen = data[0]
	r.Op = WalOp(data[1])
	r.Counter = binary.BigEndian.Uint64(data[2:10])
	r.Ts = int64(binary.BigEndian.Uint64(data[10:18]))
	r.Len = data[18]

	if r.Len > 0 {
		payload, err := cid.Cast(data[r.HdrLen : r.HdrLen+r.Len])

		if err != nil {
			return err
		}

		r.Payload = &payload
	}

	return nil
}

type WriteAheadLog struct {
	mu  sync.RWMutex
	log *wal.Log

	counter atomic.Uint64
}

func NewWriteAheadLog(path string) (*WriteAheadLog, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}

	opts := *wal.DefaultOptions
	opts.LogFormat = wal.Binary

	log, err := wal.Open(path, &opts)

	if err != nil {
		return nil, err
	}

	wl := &WriteAheadLog{
		log: log,
	}

	if err := wl.initializeFromLog(); err != nil {
		return nil, err
	}

	return wl, nil
}

func (w *WriteAheadLog) initializeFromLog() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	lastIdx, err := w.log.LastIndex()

	if err != nil {
		return err
	}

	if lastIdx > 0 {
		rec, err := w.readRecord(lastIdx)

		if err != nil {
			return err
		}

		w.counter.Store(rec.Counter)
	}

	return nil
}

func (w *WriteAheadLog) LastRecordIndex() uint64 {
	return w.counter.Load()
}

func (w *WriteAheadLog) WriteRecords(records ...WalRecord) (last uint64, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	last = 0

	for _, rec := range records {
		rec.HdrLen = 20
		rec.Counter = w.counter.Add(1)
		rec.Ts = time.Now().UnixNano()

		if rec.Payload != nil {
			rec.Len = uint8(rec.Payload.ByteLen())
		} else {
			rec.Len = 0
		}

		data, err := rec.MarshalBinary()

		if err != nil {
			return 0, err
		}

		if err := w.log.Write(rec.Counter, data); err != nil {
			return 0, err
		}

		last = rec.Counter
	}

	return last, err
}

func (w *WriteAheadLog) ReadRecord(recordIndex uint64) (*WalRecord, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.readRecord(recordIndex)
}

func (w *WriteAheadLog) readRecord(recordIndex uint64) (*WalRecord, error) {
	data, err := w.log.Read(recordIndex)

	if err != nil {
		return nil, err
	}

	rec := &WalRecord{}

	if err := rec.UnmarshalBinary(data); err != nil {
		return nil, err
	}

	return rec, nil
}

func (w *WriteAheadLog) Close() error {
	return w.log.Close()
}
