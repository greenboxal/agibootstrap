package fti

import (
	"context"
	"io/fs"
	"path/filepath"

	"github.com/pkg/errors"
)

type Iterator[T any] interface {
	Next() bool
	Item() T
}

type FileCursor struct {
	fs.DirEntry

	Path string
	Err  error
}

type chIterator[T any] struct {
	ch      chan T
	err     <-chan error
	current *T
}

func (it *chIterator[T]) HasNext() bool {
	return it.ch != nil
}

func (it *chIterator[T]) Next() bool {
	if it.ch == nil {
		return false
	}

	select {
	case err, _ := <-it.err:
		it.ch = nil
		it.current = nil

		if err != nil {
			panic(err)
		}

		return false

	case v, ok := <-it.ch:
		if ok {
			it.current = &v
		} else {
			it.ch = nil
			it.current = nil
		}

		return ok
	}
}

func (it *chIterator[T]) Item() T {
	return *it.current
}

func (it *chIterator[T]) Close() error {
	if it.ch != nil {
		close(it.ch)
		it.ch = nil
	}
	return nil
}

type osFileIterator struct {
	chIterator[FileCursor]
}

func (it *osFileIterator) File() FileCursor {
	return it.Item()
}

type filteredIterator[T any] struct {
	src     Iterator[T]
	pred    func(T) bool
	current T
}

func (f *filteredIterator[T]) Next() bool {
	for f.src.Next() {
		if f.pred(f.src.Item()) {
			f.current = f.src.Item()
			return true
		}
	}

	return false
}

func (f *filteredIterator[T]) Item() T {
	return f.current
}

// Filter creates an iterator that filters the items in the given iterator based on the provided predicate.
// It returns a new iterator that only contains items for which the predicate returns true.
func Filter[IT Iterator[T], T any](it IT, pred func(T) bool) Iterator[T] {
	return &filteredIterator[T]{src: it, pred: pred}
}

// IterateFiles traverses the directory tree rooted at dirPath and sends each file info to the channel.
// It returns an iterator of FileCursor that represents each file found.
// The context is used to control the cancellation of the traversal.
func IterateFiles(ctx context.Context, dirPath string) Iterator[FileCursor] {
	ch := make(chan FileCursor)
	errCh := make(chan error)

	go func() {
		defer close(ch)
		defer close(errCh)

		// WalkDir recursively traverses the directory tree rooted at dirPath
		// and sends each file info to the channel ch
		err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
			select {
			case <-ctx.Done():
				return errors.New("aborted")
			default:
			}

			// If there's an error, return it
			if err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return errors.New("aborted")
			case ch <- FileCursor{
				DirEntry: d,
				Path:     path,
				Err:      err,
			}:
				return nil
			}
		})

		if err != nil && err != errors.New("aborted") {
			errCh <- err
		}
	}()

	return &osFileIterator{
		chIterator: chIterator[FileCursor]{
			ch: ch,
		},
	}
}
