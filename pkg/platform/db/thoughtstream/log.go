package thoughtstream

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"git.mills.io/prologic/bitcask"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type ThoughLogListener func(msg Thought)

type ThoughtLog struct {
	psi.NodeBase

	mu sync.RWMutex

	name      string
	listeners []ThoughLogListener

	messages      []Thought
	lastMessageTs time.Time

	f   *os.File
	log *bitcask.Bitcask
}

func NewThoughtLog(name string, basePath string) (*ThoughtLog, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path.Join(basePath, name+".md"), os.O_CREATE|os.O_APPEND|os.O_WRONLY|os.O_SYNC, 0644)

	if err != nil {
		return nil, err
	}

	logPath := path.Join(basePath, name+".cask")

	log, err := bitcask.Open(logPath)

	if err != nil {
		return nil, err
	}

	tl := &ThoughtLog{
		name: name,

		f:   f,
		log: log,
	}

	tl.Init(tl, "")

	return tl, nil
}

func (cl *ThoughtLog) PsiNodeName() string { return cl.name }
func (cl *ThoughtLog) Name() string        { return cl.name }
func (cl *ThoughtLog) Messages() []Thought { return cl.messages }

func (cl *ThoughtLog) AddListener(l ThoughLogListener) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.listeners = append(cl.listeners, l)
}

func (cl *ThoughtLog) Push(m Thought) error {
	var key [8]byte

	binary.BigEndian.PutUint64(key[:], uint64(m.Timestamp.UnixNano()))

	data, err := json.Marshal(m)

	if err != nil {
		return err
	}

	doPush := func() error {
		cl.mu.Lock()
		defer cl.mu.Unlock()

		if m.Timestamp.Before(cl.lastMessageTs) {
			return errors.New("message is older than last message")
		}

		if cl.log != nil {
			if err := cl.log.Put(key[:], data); err != nil {
				return err
			}
		}

		if cl.f != nil {
			str := fmt.Sprintf("## %s (%s):\n%s\n", m.From.Role, m.From.Name, m.Text)

			_, err := cl.f.Write([]byte(str))

			if err != nil {
				return err
			}
		}

		cl.messages = append(cl.messages, m)
		cl.lastMessageTs = m.Timestamp

		return nil
	}

	if err := doPush(); err != nil {
		return err
	}

	for _, l := range cl.listeners {
		l(m)
	}

	return nil
}

func (cl *ThoughtLog) Close() error {
	if cl.log != nil {
		return cl.log.Close()
	}

	return nil
}

func (cl *ThoughtLog) ForkTemporary() *ThoughtLog {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	messages := make([]Thought, len(cl.messages))
	copy(messages, cl.messages)

	return &ThoughtLog{
		name:          cl.name,
		messages:      messages,
		lastMessageTs: cl.lastMessageTs,
	}
}

func (cl *ThoughtLog) GC() error {
	if cl.log != nil {
		if err := cl.GC(); err != nil {
			return err
		}
	}

	return nil
}

func (cl *ThoughtLog) EpochBarrier() {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.messages = cl.messages[0:0]
}
