package repofs

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	git "github.com/libgit2/git2go/v31"
)

type gitFS struct {
	repoPath string
	repo     *git.Repository
}

func New(repoPath string) (fs.FS, error) {
	repo, err := git.OpenRepository(repoPath)
	if err != nil {
		return nil, err
	}
	return &gitFS{repoPath: repoPath, repo: repo}, nil
}
func (g *gitFS) Open(name string) (fs.File, error) {
	repoPath, branch, filePath, err := splitGitPath(name)
	if err != nil {
		return nil, err
	}
	if err := g.updateRepo(repoPath); err != nil {
		return nil, err
	}
	if _, err := g.repo.LookupBranch(branch, git.BranchLocal); err != nil {
		return nil, err
	}
	commit, err := g.repo.Head()
	if err != nil {
		return nil, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}
	entry, err := tree.EntryByPath(filePath)
	if err != nil {
		return nil, err
	}
	if entry.Type != git.ObjectBlob {
		return nil, os.ErrNotExist
	}
	blob, err := g.repo.LookupBlob(entry.Id)
	if err != nil {
		return nil, err
	}
	return newGitFile(name, blob), nil
}

func (g *gitFS) Close() error {
	return nil
}
func (g *gitFS) updateRepo(repoPath string) error {
	if repoPath != g.repoPath {
		repo, err := git.OpenRepository(repoPath)
		if err != nil {
			return err
		}
		g.repo = repo
	}
	return nil
}
func splitGitPath(name string) (repoPath, branch, filePath string, err error) {
	absPath, err := filepath.Abs(name)
	if err != nil {
		return
	}
	volLen := len(filepath.VolumeName(absPath))
	if runtime.GOOS == "windows" {
		volLen += len(`\`)
	}
	if volLen == len(absPath) {
		err = os.ErrInvalid
		return
	}
	var firstSlash, secondSlash int
	for i := volLen; i < len(absPath); i++ {
		if absPath[i] == '/' || absPath[i] == '\\' {
			if firstSlash == 0 {
				firstSlash = i
			} else {
				secondSlash = i
				break
			}
		}
	}
	if firstSlash == 0 || secondSlash == 0 {
		err = os.ErrInvalid
		return
	}
	repoPath = absPath[:firstSlash]
	branch = absPath[firstSlash+1 : secondSlash]
	filePath = absPath[secondSlash+1:]
	return
}

type gitFile struct {
	fs.FileInfo
	content []byte
}

func newGitFile(name string, blob *git.Blob) *gitFile {
	return &gitFile{FileInfo: fileInfo{name: name, size: blob.Size(), mode: 0644, modTime: time.Now(), isDir: false}, content: blob.Contents()}
}
func (f *gitFile) Read(b []byte) (int, error) {
	if len(f.content) == 0 {
		return 0, io.EOF
	}
	n := copy(b, f.content)
	f.content = f.content[n:]
	return n, nil
}

type fileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (fi fileInfo) Name() string {
	return filepath.Base(fi.name)
}
func (fi fileInfo) Size() int64 {
	return fi.size
}
func (fi fileInfo) Mode() fs.FileMode {
	return fi.mode
}
func (fi fileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi fileInfo) IsDir() bool {
	return fi.isDir
}
func (fi fileInfo) Sys() interface{} {
	return nil
}

type dirIter struct {
	dir     *git.Tree
	dirPath string
	names   []string
	err     error
	m       sync.Mutex
}

func (d *dirIter) Next() bool {
	d.m.Lock()
	defer d.m.Unlock()
	if len(d.names) > 0 {
		return true
	}
	if d.err != nil {
		return false
	}
	if err := d.dir.Walk(func(currPath string, entry *git.TreeEntry) int {
		if entry.Type != git.ObjectBlob {
			return 0
		}
		d.names = append(d.names, filepath.Join(d.dirPath, currPath, entry.Name))
		return 0
	}); err != nil {
		d.err = err
		return false
	}
	return len(d.names) > 0
}
func (d *dirIter) FileInfo() (fs.FileInfo, error) {
	if len(d.names) == 0 {
		return nil, d.err
	}
	return newGitFileInfo(d.names[0], d.dir.EntryByName(filepath.ToSlash(d.names[0]))), nil
}
func (d *dirIter) Err() error {
	return d.err
}
func (d *dirIter) Close() error {
	return nil
}

type gitFileInfo struct {
	name string
	*git.TreeEntry
}

func newGitFileInfo(name string, entry *git.TreeEntry) *gitFileInfo {
	return &gitFileInfo{name: name, TreeEntry: entry}
}
func (fi *gitFileInfo) Name() string {
	return filepath.Base(fi.name)
}
func (fi *gitFileInfo) Size() int64 {
	return fi.Filemode().Size()
}
func (fi *gitFileInfo) Mode() fs.FileMode {
	return fi.Filemode().Filemode()
}
func (fi *gitFileInfo) ModTime() time.Time {
	return time.Now()
}
func (fi *gitFileInfo) IsDir() bool {
	return fi.Filemode().IsDir()
}
func (fi *gitFileInfo) Sys() interface{} {
	return nil
}

type gitFSFile struct{ fs.File }

func (f *gitFSFile) Stat() (fs.FileInfo, error) {
	if gf, ok := f.File.(*gitFile); ok {
		return gf.FileInfo, nil
	}
	return f.File.Stat()
}
func (g *gitFS) ReadDir(name string) ([]fs.DirEntry, error) {
	repoPath, branch, dirPath, err := splitGitPath(name)
	if err != nil {
		return nil, err
	}
	if err := g.updateRepo(repoPath); err != nil {
		return nil, err
	}
	if _, err := g.repo.LookupBranch(branch, git.BranchLocal); err != nil {
		return nil, err
	}
	commit, err := g.repo.Head()
	if err != nil {
		return nil, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}
	if dirPath == "" {
		return g.rootDirEntries(tree)
	}
	dir, err := tree.EntryByPath(dirPath)
	if err != nil {
		return nil, err
	}
	if dir.Type != git.ObjectTree {
		return nil, os.ErrNotExist
	}
	iter := &dirIter{dir: tree.Subdir(dirPath), dirPath: dirPath}
	var res []fs.DirEntry
	for iter.Next() {
		fi, err := iter.FileInfo()
		if err != nil {
			return nil, err
		}
		res = append(res, fi)
	}
	return res, iter.Err()
}
func (g *gitFS) rootDirEntries(tree *git.Tree) ([]fs.DirEntry, error) {
	var res []fs.DirEntry
	if err := tree.Walk(func(currPath string, entry *git.TreeEntry) error {
		if entry.Type != git.ObjectTree {
			return nil
		}
		res = append(res, newGitFileInfo(currPath, entry))
		return nil
	}); err != nil {
		return nil, err
	}
	return res, nil
}
func (g *gitFS) OpenFS() (fs.FS, error) {
	return g, nil
}
func (g *gitFS) Remove(name string) error {
	return os.ErrPermission
}
func (g *gitFS) Rename(oldname, newname string) error {
	return os.ErrPermission
}
func (g *gitFS) Stat(name string) (fs.FileInfo, error) {
	repoPath, branch, filePath, err := splitGitPath(name)
	if err != nil {
		return nil, err
	}
	if err := g.updateRepo(repoPath); err != nil {
		return nil, err
	}
	if _, err := g.repo.LookupBranch(branch, git.BranchLocal); err != nil {
		return nil, err
	}
	commit, err := g.repo.Head()
	if err != nil {
		return nil, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}
	entry, err := tree.EntryByPath(filePath)
	if err != nil {
		return nil, err
	}
	if entry.Type == git.ObjectTree {
		return newGitFileInfo(name, entry), nil
	}
	if entry.Type != git.ObjectBlob {
		return nil, os.ErrNotExist
	}
	blob, err := g.repo.LookupBlob(entry.Id)
	if err != nil {
		return nil, err
	}
	return newGitFile(name, blob).FileInfo, nil
}
func (g *gitFS) Mkdir(name string, perm fs.FileMode) error {
	return os.ErrPermission
}
func (g *gitFS) Create(name string) (fs.File, error) {
	return nil, os.ErrPermission
}
