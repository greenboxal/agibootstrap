package agents

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
)

type ChatLogListener func(msg Message)

type ChatLog struct {
	mu sync.RWMutex

	name      string
	listeners []ChatLogListener

	messages      []Message
	lastMessageTs time.Time

	f   *os.File
	log *bitcask.Bitcask
}

func NewChatLog(name string) (*ChatLog, error) {
	if err := os.MkdirAll("/tmp/agib-agent-logs", 0755); err != nil {
		return nil, err
	}

	f, err := os.OpenFile("/tmp/agib-agent-logs/"+name+".md", os.O_CREATE|os.O_APPEND|os.O_WRONLY|os.O_SYNC, 0644)

	if err != nil {
		return nil, err
	}

	logPath := path.Join("/tmp/agib-agent-logs/", name+".cask")

	log, err := bitcask.Open(logPath)

	if err != nil {
		return nil, err
	}

	return &ChatLog{
		name: name,

		f:   f,
		log: log,
	}, nil
}

func (cl *ChatLog) Messages() []Message {
	return cl.messages
}

func (cl *ChatLog) AddListener(l ChatLogListener) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.listeners = append(cl.listeners, l)
}

func (cl *ChatLog) Push(m Message) error {
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

func (cl *ChatLog) Close() error {
	if cl.log != nil {
		return cl.log.Close()
	}

	return nil
}

func (cl *ChatLog) ForkTemporary() *ChatLog {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	messages := make([]Message, len(cl.messages))
	copy(messages, cl.messages)

	return &ChatLog{
		name:          cl.name,
		messages:      messages,
		lastMessageTs: cl.lastMessageTs,
	}
}

func (cl *ChatLog) GC() error {
	if cl.log != nil {
		if err := cl.GC(); err != nil {
			return err
		}
	}

	return nil
}
