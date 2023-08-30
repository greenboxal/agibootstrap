package vfs

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"
	"github.com/pkg/errors"
	`github.com/uptrace/opentelemetry-go-extra/otelzap`

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type FileSystem interface {
	fs.FS

	WriteFile(p string, src io.Reader) error
}

type FileSystemOption func(*fileSystem)

func WithPathFilter(pathFilter func(string) bool) FileSystemOption {
	return func(fsys *fileSystem) {
		fsys.pathFilter = pathFilter
	}
}

type fileSystem struct {
	psi.NodeBase

	logger *otelzap.SugaredLogger

	root    string
	manager *Manager

	mu      sync.Mutex
	nodeMap map[string]Node
	watcher *fsnotify.Watcher
	proc    goprocess.Process

	stopCh chan struct{}

	pathFilter func(string) bool
}

var FileSystemType = psi.DefineNodeType[*fileSystem](psi.WithRuntimeOnly())

func (bfs *fileSystem) Init(self psi.Node) {
	bfs.NodeBase.Init(self, psi.WithNodeType(FileSystemType))
}

func newLocalFS(m *Manager, rootPath string, options ...FileSystemOption) (*fileSystem, error) {
	if _, err := os.Stat(rootPath); err != nil {
		return nil, errors.Wrap(err, "invalid root Path")
	}

	w, err := fsnotify.NewWatcher()

	if err != nil {
		return nil, err
	}

	bfs := &fileSystem{
		logger: logging.GetLogger("vfs"),

		root:    rootPath,
		manager: m,

		nodeMap: map[string]Node{},
		watcher: w,

		stopCh: make(chan struct{}),
	}

	for _, option := range options {
		option(bfs)
	}

	bfs.Init(bfs)

	bfs.proc = goprocess.Go(bfs.run)

	return bfs, nil
}

func (bfs *fileSystem) GetNodeForPath(ctx context.Context, path string) (n Node, err error) {
	ok, err := isChildPath(bfs.root, path)

	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fs.ErrNotExist
	}

	if bfs.pathFilter != nil && !bfs.pathFilter(path) {
		return nil, fs.ErrNotExist
	}

	bfs.mu.Lock()
	defer bfs.mu.Unlock()

	return bfs.getNodeForPathUnlocked(ctx, path)
}

func (bfs *fileSystem) getNodeForPathUnlocked(ctx context.Context, path string) (n Node, err error) {
	if n = bfs.nodeMap[path]; n != nil {
		return n, nil
	}

	stat, err := os.Stat(path)

	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		n = newDirectoryNode(bfs, path, "")
	} else if stat.Mode().Type() == os.ModeSymlink {
		n = newFileNode(bfs, path)
	}

	parentPath := filepath.Dir(path)

	if ok, err := isChildPath(bfs.root, parentPath); err == nil && ok && parentPath != bfs.root {
		parent, err := bfs.getNodeForPathUnlocked(ctx, parentPath)

		if err != nil {
			return nil, err
		}

		n.SetParent(parent)
	}

	if err := n.Update(ctx); err != nil {
		return nil, err
	}

	bfs.nodeMap[path] = n

	return n, nil
}

func (bfs *fileSystem) Open(name string) (fs.File, error) {
	fullname, err := bfs.join(name)

	if err != nil {
		return nil, &fs.PathError{Op: "lastStat", Path: name, Err: err}
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

func (bfs *fileSystem) WriteFile(p string, src io.Reader) error {
	w, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		return err
	}

	_, err = io.Copy(w, src)

	if err != nil {
		return err
	}

	return nil
}

func (bfs *fileSystem) Stat(name string) (fs.FileInfo, error) {
	fullname, err := bfs.join(name)

	if err != nil {
		return nil, &fs.PathError{Op: "lastStat", Path: name, Err: err}
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
	defer bfs.manager.notifyClose(bfs)

	bfs.mu.Lock()
	defer bfs.mu.Unlock()

	if bfs.proc != nil {
		close(bfs.stopCh)

		if err := bfs.proc.Close(); err != nil {
			return err
		}

		bfs.proc = nil
	}

	if bfs.watcher != nil {
		if err := bfs.watcher.Close(); err != nil {
			return err
		}

		bfs.watcher = nil
	}

	return nil
}

func (bfs *fileSystem) run(proc goprocess.Process) {
	ctx := goprocessctx.OnClosingContext(proc)

	for {
		select {
		case <-proc.Closing():
			return

		case <-bfs.stopCh:
			return

		case ev := <-bfs.watcher.Events:
			func() {
				defer func() {
					if r := recover(); r != nil {
						bfs.logger.Error(r)
					}
				}()

				if err := bfs.handleEvent(ctx, ev); err != nil {
					bfs.logger.Error(err)
				}
			}()

		case err := <-bfs.watcher.Errors:
			bfs.logger.Error(err)
		}
	}
}

func (bfs *fileSystem) handleEvent(ctx context.Context, ev fsnotify.Event) error {
	node := bfs.nodeMap[ev.Name]

	if node == nil {
		return nil
	}

	return node.onWatchEvent(ctx, ev)
}

func (bfs *fileSystem) addWatch(nb *NodeBase) error {
	bfs.mu.Lock()
	defer bfs.mu.Unlock()

	return bfs.watcher.Add(nb.GetPath())
}

func (bfs *fileSystem) removeWatch(nb *NodeBase) error {
	bfs.mu.Lock()
	defer bfs.mu.Unlock()

	return bfs.watcher.Remove(nb.GetPath())
}
