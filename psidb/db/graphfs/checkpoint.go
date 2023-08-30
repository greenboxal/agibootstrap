package graphfs

import (
	"encoding/binary"
	"io"
	"os"
	"sync"
)

type Checkpoint interface {
	Get() (uint64, bool, error)
	Update(xid uint64, onlyIfGreater bool) error
	Close() error
}

type fileCheckpoint struct {
	mu   sync.RWMutex
	file *os.File

	value  uint64
	loaded bool
}

func OpenFileCheckpoint(path string) (Checkpoint, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		return nil, err
	}

	return NewFileCheckpoint(f), nil
}

func NewFileCheckpoint(f *os.File) Checkpoint {
	return &fileCheckpoint{
		file: f,
	}
}

func (cp *fileCheckpoint) Get() (uint64, bool, error) {
	cp.mu.RLock()
	if cp.loaded {
		v := cp.value
		cp.mu.RUnlock()
		return v, true, nil
	}
	cp.mu.RUnlock()
	cp.mu.Lock()
	defer cp.mu.Unlock()

	return cp.getOrLoadUnlocked()
}

func (cp *fileCheckpoint) getOrLoadUnlocked() (uint64, bool, error) {
	var buffer [8]byte

	if cp.loaded {
		return cp.value, true, nil
	}

	if pos, err := cp.file.Seek(0, io.SeekEnd); err != nil {
		return 0, false, err
	} else if pos == 0 {
		return 0, true, nil
	}

	if _, err := cp.file.Seek(0, 0); err != nil {
		return 0, false, err
	}

	if _, err := io.ReadFull(cp.file, buffer[:]); err != nil {
		return 0, false, err
	}

	xid := binary.BigEndian.Uint64(buffer[:])

	cp.value = xid
	cp.loaded = true

	return xid, true, nil
}

func (cp *fileCheckpoint) Update(xid uint64, onlyIfGreater bool) error {
	var buffer [8]byte

	binary.BigEndian.PutUint64(buffer[:], xid)

	cp.mu.Lock()
	defer cp.mu.Unlock()

	if onlyIfGreater {
		current, _, err := cp.getOrLoadUnlocked()

		if err != nil {
			return err
		}

		if xid <= current {
			return nil
		}

		return nil
	}

	if _, err := cp.file.Seek(0, 0); err != nil {
		return err
	}

	if _, err := cp.file.Write(buffer[:]); err != nil {
		return err
	}

	cp.loaded = true
	cp.value = xid

	return nil
}

func (cp *fileCheckpoint) Close() error {
	return cp.file.Close()
}
