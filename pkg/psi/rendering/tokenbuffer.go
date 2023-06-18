package rendering

import (
	"bytes"
	"io"

	"github.com/greenboxal/aip/aip-langchain/pkg/tokenizers"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type TokenBuffer struct {
	tokenizer  tokenizers.BasicTokenizer
	tokenCount int
	tokenLimit int

	state *NodeState
	head  *tokenRope
	tail  *tokenRope
}

type tokenRope struct {
	tb   *TokenBuffer
	node *NodeState
	buf  bytes.Buffer

	isTerminal bool
	tokenCount int

	prev *tokenRope
	next *tokenRope
}

func (r *tokenRope) InsertBefore(other *tokenRope) *tokenRope {
	tail := other

	if r.tb.head == r {
		head := other

		for head.prev != nil && head != r {
			head = head.prev
		}

		r.tb.head = head
	}

	for tail.next != nil && tail != r {
		tail = tail.next
	}

	other.prev = r.prev

	if r.prev != nil {
		r.prev.next = other
	}

	tail.next = r
	r.prev = tail

	return other
}

func (r *tokenRope) InsertAfter(other *tokenRope) *tokenRope {
	tail := other

	for tail.next != nil && tail.next != r {
		tail = tail.next
	}

	other.next = r.next

	if r.next != nil {
		r.next.prev = other
	}

	tail.next = r
	r.next = tail

	if r.tb.tail == r {
		r.tb.tail = tail
	}

	return other
}

func (r *tokenRope) Append() *tokenRope {
	return r.tb.tail.InsertBefore(&tokenRope{tb: r.tb})
}

func (r *tokenRope) Prepend() *tokenRope {
	return r.tb.head.InsertBefore(&tokenRope{tb: r.tb})
}

func (r *tokenRope) PrependBefore() *tokenRope {
	return r.InsertAfter(&tokenRope{tb: r.tb})
}

func (r *tokenRope) AppendAfter() *tokenRope {
	return r.InsertAfter(&tokenRope{tb: r.tb})
}

func NewTokenBuffer(tokenizer tokenizers.BasicTokenizer, limit int) *TokenBuffer {
	tb := &TokenBuffer{
		tokenizer:  tokenizer,
		tokenLimit: limit,
	}

	tb.Reset()

	return tb
}

func (w *TokenBuffer) TokenLimit() int { return w.tokenCount }
func (w *TokenBuffer) TokenCount() int { return w.tokenCount }

func (w *TokenBuffer) SetTokenLimit(limit int) { w.tokenLimit = limit }

func (w *TokenBuffer) IsValid() bool {
	return w.tokenCount <= w.tokenLimit
}

func (r *tokenRope) invalidate() error {
	count, err := r.tb.tokenizer.Count(r.buf.String())

	if err != nil {
		return err
	}

	r.tokenCount = count

	return nil
}

func (w *TokenBuffer) WriteNode(renderer *PruningRenderer, node psi.Node) (total int, err error) {
	r := w.head.Append()
	r.node = renderer.getState(node)
	r.isTerminal = false

	if err := r.node.Update(renderer); err != nil {
		return total, err
	}

	r.tokenCount = r.node.Buffer.TokenCount()
	w.tokenCount += r.tokenCount

	return total, nil
}

func (w *TokenBuffer) Write(data []byte) (int, error) {
	r := w.head.Append()
	r.isTerminal = true
	r.node = w.state

	n, err := r.buf.Write(data)

	if err != nil {
		return n, err
	}

	if err := r.invalidate(); err != nil {
		return 0, err
	}

	w.tokenCount += r.tokenCount

	return n, nil
}

func (w *TokenBuffer) Reset() {
	w.head = &tokenRope{tb: w}
	w.tail = w.head
	w.tokenCount = 0
}

func (w *TokenBuffer) WriteTo(writer io.Writer) (total int64, err error) {
	for buf := w.head; buf != nil; buf = buf.next {
		if buf.isTerminal {
			n, err := writer.Write(buf.buf.Bytes())

			total += int64(n)

			if err != nil {
				return total, err
			}
		} else if buf.node != nil {
			n, err := buf.node.WriteTo(writer)

			total += n

			if err != nil {
				return total, err
			}
		}

		if buf.next == w.head {
			break
		}
	}

	return
}

func (w *TokenBuffer) Bytes() []byte {
	buf := bytes.NewBuffer(nil)

	_, err := w.WriteTo(buf)

	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func (w *TokenBuffer) String() string {
	buf := bytes.NewBuffer(nil)

	_, err := w.WriteTo(buf)

	if err != nil {
		panic(err)
	}

	return buf.String()
}
