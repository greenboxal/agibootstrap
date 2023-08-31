package coreapi

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

type NullCheckpoint struct{}

func (ckpt NullCheckpoint) Get() (uint64, bool, error) {
	return 0, false, nil
}

func (ckpt NullCheckpoint) Update(xid uint64, onlyIfGreater bool) error {
	return nil
}

func (ckpt NullCheckpoint) Close() error {
	return nil
}

type inMemoryCheckpoint struct {
	mu    sync.RWMutex
	value uint64
}

func NewInMemoryCheckpoint() Checkpoint {
	return &inMemoryCheckpoint{}
}

func (ckpt *inMemoryCheckpoint) Get() (uint64, bool, error) {
	ckpt.mu.RLock()
	defer ckpt.mu.RUnlock()

	return ckpt.value, true, nil
}

func (ckpt *inMemoryCheckpoint) Update(xid uint64, onlyIfGreater bool) error {
	ckpt.mu.Lock()
	defer ckpt.mu.Unlock()

	if onlyIfGreater {
		if xid <= ckpt.value {
			return nil
		}
	}

	ckpt.value = xid

	return nil
}

func (ckpt *inMemoryCheckpoint) Close() error {
	return nil
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

func (ckpt *fileCheckpoint) Get() (uint64, bool, error) {
	ckpt.mu.RLock()
	if ckpt.loaded {
		v := ckpt.value
		ckpt.mu.RUnlock()
		return v, true, nil
	}
	ckpt.mu.RUnlock()
	ckpt.mu.Lock()
	defer ckpt.mu.Unlock()

	return ckpt.getOrLoadUnlocked()
}

func (ckpt *fileCheckpoint) getOrLoadUnlocked() (uint64, bool, error) {
	var buffer [8]byte

	if ckpt.loaded {
		return ckpt.value, true, nil
	}

	if pos, err := ckpt.file.Seek(0, io.SeekEnd); err != nil {
		return 0, false, err
	} else if pos == 0 {
		return 0, true, nil
	}

	if _, err := ckpt.file.Seek(0, 0); err != nil {
		return 0, false, err
	}

	if _, err := io.ReadFull(ckpt.file, buffer[:]); err != nil {
		return 0, false, err
	}

	xid := binary.BigEndian.Uint64(buffer[:])

	ckpt.value = xid
	ckpt.loaded = true

	return xid, true, nil
}

func (ckpt *fileCheckpoint) Update(xid uint64, onlyIfGreater bool) error {
	var buffer [8]byte

	binary.BigEndian.PutUint64(buffer[:], xid)

	ckpt.mu.Lock()
	defer ckpt.mu.Unlock()

	if onlyIfGreater {
		current, _, err := ckpt.getOrLoadUnlocked()

		if err != nil {
			return err
		}

		if xid <= current {
			return nil
		}

		return nil
	}

	if _, err := ckpt.file.Seek(0, 0); err != nil {
		return err
	}

	if _, err := ckpt.file.Write(buffer[:]); err != nil {
		return err
	}

	ckpt.loaded = true
	ckpt.value = xid

	return nil
}

func (ckpt *fileCheckpoint) Close() error {
	return ckpt.file.Close()
}
