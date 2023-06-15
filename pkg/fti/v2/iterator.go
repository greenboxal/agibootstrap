package fti

import (
	"context"
	"io/fs"
	"path/filepath"
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

func Filter[IT Iterator[T], T any](it IT, pred func(T) bool) Iterator[T] {
	return &filteredIterator[T]{src: it, pred: pred}
}

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
				return ErrAbort
			default:
			}

			// If there's an error, return it
			if err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return ErrAbort
			case ch <- FileCursor{
				DirEntry: d,
				Path:     path,
				Err:      err,
			}:
				return nil
			}
		})

		if err != nil && err != ErrAbort {
			errCh <- err
		}
	}()

	return &osFileIterator{
		chIterator: chIterator[FileCursor]{
			ch: ch,
		},
	}
}
