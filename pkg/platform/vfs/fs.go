package vfs

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/jbenet/goprocess"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
)

type FS interface {
	fs.FS
}

type fileSystem struct {
	logger *zap.SugaredLogger

	root string

	mu       sync.Mutex
	watchMap map[string]Node
	watcher  *fsnotify.Watcher
	proc     goprocess.Process

	stopCh chan struct{}
}

func NewFS(rootPath string) (FS, error) {
	if _, err := os.Stat(rootPath); err != nil {
		return nil, errors.Wrap(err, "invalid root path")
	}

	w, err := fsnotify.NewWatcher()

	if err != nil {
		return nil, err
	}

	bfs := &fileSystem{
		logger: logging.GetLogger("vfs"),

		root: rootPath,

		watchMap: map[string]Node{},
		watcher:  w,

		stopCh: make(chan struct{}),
	}

	bfs.proc = goprocess.Go(bfs.run)

	return bfs, nil
}

func (bfs *fileSystem) Watch(node Node) error {
	bfs.mu.Lock()
	defer bfs.mu.Unlock()

	if err := bfs.watcher.Add(node.Path()); err != nil {
		return err
	}

	bfs.watchMap[node.Path()] = node

	return nil
}

func (bfs *fileSystem) Unwatch(node Node) error {
	bfs.mu.Lock()
	defer bfs.mu.Unlock()

	if _, ok := bfs.watchMap[node.Path()]; !ok {
		return nil
	}

	if err := bfs.watcher.Remove(node.Path()); err != nil {
		return err
	}

	return nil
}

func (bfs *fileSystem) Open(name string) (fs.File, error) {
	fullname, err := bfs.join(name)

	if err != nil {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: err}
	}

	f, err := os.Open(fullname)

	if err != nil {
		// DirFS takes a string appropriate for GOOS,
		// while the name argument here is always slash separated.
		// bfs.join will have mixed the two; undo that for
		// error reporting.
		err.(*os.PathError).Path = name
		return nil, err
	}

	return f, nil
}

func (bfs *fileSystem) Stat(name string) (fs.FileInfo, error) {
	fullname, err := bfs.join(name)

	if err != nil {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: err}
	}

	f, err := os.Stat(fullname)

	if err != nil {
		err.(*fs.PathError).Path = name
		return nil, err
	}

	return f, nil
}

// join returns the path for name in dir.
func (bfs *fileSystem) join(name string) (string, error) {
	var relPath, absPath string

	if bfs.root == "" {
		return "", errors.New("repofs: fileSystem with empty root")
	}

	if path.IsAbs(name) {
		absPath = name

		rel, relErr := filepath.Rel(bfs.root, name)

		if relErr == nil {
			relPath = rel
		}
	} else {
		relPath = name

		abs, err := filepath.Abs(path.Join(bfs.root, relPath))

		if err == nil {
			absPath = abs
		}
	}

	if relPath == "" || absPath == "" {
		return "", errors.New("file outside of project")
	}

	return absPath, nil
}

func (bfs *fileSystem) Close() error {
	if bfs.watcher != nil {
		if err := bfs.watcher.Close(); err != nil {
			return err
		}

		bfs.watcher = nil
	}

	return nil
}

func (bfs *fileSystem) run(proc goprocess.Process) {
	for {
		select {
		case <-proc.Closing():
			return

		case <-bfs.stopCh:
			return

		case ev := <-bfs.watcher.Events:
			if err := bfs.handleEvent(ev); err != nil {
				bfs.logger.Error(err)
			}

		case err := <-bfs.watcher.Errors:
			bfs.logger.Error(err)
		}
	}
}

func (bfs *fileSystem) handleEvent(ev fsnotify.Event) error {
	node := bfs.watchMap[ev.Name]

	if node == nil {
		return nil
	}

	return node.onWatchEvent(ev)
}
