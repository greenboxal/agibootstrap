package graphfs

import (
	"encoding/binary"
	"io"
	"os"
	"sync"
)

type Checkpoint interface {
	Get() (uint64, bool, error)
	Update(xid uint64) error
	Close() error
}

type fileCheckpoint struct {
	mu   sync.Mutex
	file *os.File
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
	var buffer [8]byte

	cp.mu.Lock()
	defer cp.mu.Unlock()

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

	return xid, true, nil
}

func (cp *fileCheckpoint) Update(xid uint64) error {
	var buffer [8]byte

	binary.BigEndian.PutUint64(buffer[:], xid)

	cp.mu.Lock()
	defer cp.mu.Unlock()

	if _, err := cp.file.Seek(0, 0); err != nil {
		return err
	}

	if _, err := cp.file.Write(buffer[:]); err != nil {
		return err
	}

	return nil
}

func (cp *fileCheckpoint) Close() error {
	return cp.file.Close()
}
