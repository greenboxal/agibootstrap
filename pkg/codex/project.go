package codex

import (
	"context"
	"fmt"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/google/uuid"
	"github.com/greenboxal/aip/aip-langchain/pkg/chunkers"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	"github.com/ipfs/go-datastore"
	badger "github.com/ipfs/go-ds-badger"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/gpt/cache"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/fti"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/graphindex"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/graphstore"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	tasks "github.com/greenboxal/agibootstrap/pkg/platform/tasks"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/analysis"
)

const SourceFileEdge psi.TypedEdgeKind[psi.SourceFile] = "SourceFile"

type Project struct {
	psi.NodeBase

	logger *zap.SugaredLogger

	uuid   string
	config project.Config

	indexedGraph *graphstore.IndexedGraph
	indexManager *graphindex.Manager
	embedder     llm.Embedder

	ds datastore.Batching
	fs repofs.FS

	repo *fti.Repository
	tm   *tasks.Manager
	lm   *thoughtdb.Repo
	sm   *SyncManager

	rootPath string
	rootNode *vfs.Directory

	langRegistry *project.Registry

	fset *token.FileSet
}

var ProjectType = psi.DefineNodeType[*Project](psi.WithRuntimeOnly())

func NewBareProject(rootPath string) (*Project, error) {
	rootFs, err := repofs.NewFS(rootPath)

	if err != nil {
		return nil, err
	}

	repo, err := fti.NewRepository(rootPath)

	if err != nil {
		return nil, err
	}

	dsOpts := badger.DefaultOptions

	dsPath := repo.ResolveDbPath("codex", "db")

	if err := os.MkdirAll(dsPath, 0755); err != nil {
		return nil, errors.Wrap(err, "failed to create datastore directory")
	}

	ds, err := badger.NewDatastore(dsPath, &dsOpts)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create datastore")
	}

	p := &Project{
		logger: logging.GetLogger("codex"),

		rootPath: rootPath,

		ds:   ds,
		fs:   rootFs,
		repo: repo,

		fset: token.NewFileSet(),

		embedder: cache.NewCachedEmbedder(ds, gpt.GlobalEmbedder),
	}

	return p, nil
}

// LoadProject creates a new codex project with the given root path.
// It initializes the project file system, repository, and other required data structures.
// It returns a pointer to the created Project object and an error if any.
func LoadProject(ctx context.Context, rootPath string) (*Project, error) {
	p, err := NewBareProject(rootPath)

	if err != nil {
		return nil, err
	}

	if err := p.Load(ctx); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Project) Init(self psi.Node, projectUuid string) {
	p.uuid = projectUuid

	p.NodeBase.Init(self, psi.WithNodeType(ProjectType))

	scope := analysis.NewScope(p)
	scope.SetParent(p)
	analysis.SetNodeScope(p, scope)
}

func (p *Project) IsProjectValid(ctx context.Context) (bool, error) {
	projectUuid, err := p.ds.Get(ctx, datastore.NewKey("project-uuid"))

	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			projectUuid = nil
		} else {
			return false, errors.Wrap(err, "failed to get project uuid")
		}
	}

	if len(projectUuid) == 0 {
		return false, nil
	}

	return true, nil
}

func (p *Project) Create(ctx context.Context) error {
	if p.uuid != "" {
		return errors.New("project already initialized")
	}

	projectUuid := []byte(uuid.New().String())

	if err := os.MkdirAll(path.Join(p.rootPath, ".codex"), 0755); err != nil {
		return errors.Wrap(err, "failed to create codex directory")
	}

	if err := os.WriteFile(path.Join(p.rootPath, ".codex", "project-uuid"), projectUuid, 0644); err != nil {
		return errors.Wrap(err, "failed to write project uuid")
	}

	err := p.ds.Put(ctx, datastore.NewKey("project-uuid"), projectUuid)

	if err != nil {
		return errors.Wrap(err, "failed to put project uuid")
	}

	if err := p.Initialize(ctx, string(projectUuid)); err != nil {
		return err
	}

	if err := p.indexedGraph.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (p *Project) Load(ctx context.Context) error {
	projectUuid, err := p.ds.Get(ctx, datastore.NewKey("project-uuid"))

	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			projectUuid = nil
		} else {
			return errors.Wrap(err, "failed to get project uuid")
		}
	}

	if len(projectUuid) == 0 {
		return errors.Wrap(err, "project uuid not found")
	}

	if err := p.Initialize(ctx, string(projectUuid)); err != nil {
		return err
	}

	return nil
}

func (p *Project) Initialize(ctx context.Context, projectUuid string) error {
	var err error

	if p.uuid != "" {
		return errors.New("project already initialized")
	}

	p.Init(p, projectUuid)

	if err := p.loadConfig(ctx); err != nil {
		return err
	}

	p.indexedGraph, err = graphstore.NewIndexedGraph(p.ds, p.repo.ResolveDbPath("codex", "wal"), p)

	if err != nil {
		return errors.Wrap(err, "failed to create graph")
	}

	p.indexManager = graphindex.NewManager(p.indexedGraph, p.repo.ResolveDbPath("codex", "index"))

	p.langRegistry = project.NewRegistry(p)

	p.tm = tasks.NewManager()
	p.tm.SetParent(p)

	p.lm = thoughtdb.NewRepo(p.indexedGraph)
	p.lm.SetParent(p)

	p.sm = NewSyncManager(p)
	p.sm.SetParent(p)

	p.rootNode = vfs.NewDirectoryNode(p.fs, p.rootPath, "srcs")
	p.rootNode.SetParent(p)

	pathIndex, err := p.indexManager.OpenNodeIndex(ctx, "node-by-path", &graphindex.AnchoredEmbedder{
		Base:    p.embedder,
		Root:    p,
		Anchor:  p.rootNode,
		Chunker: &chunkers.TikToken{},
	})

	if err != nil {
		return errors.Wrap(err, "failed to open path index")
	}

	p.indexedGraph.AddListener(graphstore.IndexedGraphListenerFunc(func(node psi.Node) {
		if err := pathIndex.IndexNode(context.Background(), node); err != nil {
			p.logger.Error(err)
		}
	}))

	if err := p.indexedGraph.RefreshNode(ctx, p); err != nil && !errors.Is(err, psi.ErrNodeNotFound) {
		return errors.Wrap(err, "failed to refresh project node")
	}

	if err := p.indexedGraph.RefreshNode(ctx, p.lm); err != nil && !errors.Is(err, psi.ErrNodeNotFound) {
		return errors.Wrap(err, "failed to refresh thoughtdb node")
	}

	if err := p.bootstrap(ctx); err != nil {
		return err
	}

	return nil
}

func (p *Project) bootstrap(ctx context.Context) error {
	if err := p.sm.RequestSync(ctx, true); err != nil {
		return err
	}

	for _, lang := range p.config.Project.EnabledLanguages {
		if err := p.bootstrapLanguage(ctx, lang); err != nil {
			return err
		}
	}

	/*for _, mod := range p.config.Modules {
		if err := p.bootstrapModule(ctx, mod); err != nil {
			return err
		}
	}*/

	if err := p.sm.WaitForInitialSync(ctx); err != nil {
		return err
	}

	p.logger.Info("Project bootstrapped")

	return nil
}

func (p *Project) bootstrapLanguage(ctx context.Context, name string) error {
	factory := project.GetLanguageFactory(psi.LanguageID(name))

	if factory == nil {
		return errors.Errorf("language %s not found", name)
	}

	p.langRegistry.Register(factory(p))

	return nil
}

func (p *Project) bootstrapModule(ctx context.Context, mod project.ModuleConfig) error {
	lang := p.langRegistry.GetLanguage(psi.LanguageID(mod.Language))

	if lang == nil {
		return errors.Errorf("language %s not configured", mod.Language)
	}

	root, err := psi.Resolve(ctx, p.rootNode, mod.Path)

	if err != nil {
		return errors.Wrapf(err, "failed to resolve module root %s", mod.Path)
	}

	m, err := NewModule(p, mod, lang, root.(*vfs.Directory))

	if err != nil {
		return err
	}

	m.SetParent(p)

	return nil
}

func (p *Project) loadConfig(ctx context.Context) error {
	configPath := path.Join(p.rootPath, "Codex.project.toml")

	data, err := os.ReadFile(configPath)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	if _, err := toml.Decode(string(data), &p.config); err != nil {
		return errors.Wrap(err, "failed to decode config")
	}

	return nil
}

func (p *Project) UUID() string                      { return p.uuid }
func (p *Project) TaskManager() *tasks.Manager       { return p.tm }
func (p *Project) LogManager() *thoughtdb.Repo       { return p.lm }
func (p *Project) IndexManager() *graphindex.Manager { return p.indexManager }

// RootPath returns the root path of the project.
func (p *Project) RootPath() string   { return p.rootPath }
func (p *Project) RootNode() psi.Node { return p.rootNode }

// FS returns the file system interface of the project.
// It provides methods for managing the project's files and directories.
func (p *Project) FS() repofs.FS { return p.fs }

func (p *Project) Graph() *graphstore.IndexedGraph     { return p.indexedGraph }
func (p *Project) LanguageProvider() *project.Registry { return p.langRegistry }

func (p *Project) Repo() *fti.Repository { return p.repo }

func (p *Project) FileSet() *token.FileSet { return p.fset }
func (p *Project) Embedder() llm.Embedder  { return p.embedder }

// Sync synchronizes the project with the file system.
// It scans all files in the project directory and imports
// valid files into the project. Valid files are those that
// have valid file extensions specified in the `validExtensions`
// slice. The function returns an error if any occurs during
// the sync process.
func (p *Project) Sync(ctx context.Context) error {
	return p.sm.RequestSync(ctx, false)
}

// GetSourceFile retrieves the source file with the given filename from the project.
// It returns a pointer to the psi.SourceFile and any error that occurred during the process.
func (p *Project) GetSourceFile(ctx context.Context, filename string) (_ psi.SourceFile, err error) {
	relPath, err := filepath.Rel(p.rootPath, filename)

	if err != nil {
		return nil, err
	}

	psiPath := psi.MustParsePath(relPath)
	fileNode, err := psi.ResolvePath(ctx, p.rootNode, psiPath)

	if err != nil {
		return nil, err
	}

	existing := psi.ResolveEdge(fileNode, SourceFileEdge.Singleton())

	if existing == nil {
		if err != nil {
			return nil, err
		}

		lang := p.langRegistry.ResolveExtension(filepath.Base(filename))

		if lang == nil {
			return nil, fmt.Errorf("failed to resolve language for file %s", filename)
		}

		existing = lang.CreateSourceFile(ctx, filename, &repofs.FsFileHandle{
			FS:   p.fs,
			Path: strings.TrimPrefix(filename, p.rootPath+"/"),
		})

		existing.SetParent(fileNode)
		fileNode.SetEdge(SourceFileEdge.Singleton(), existing)

		err = existing.Load(ctx)

		if err != nil {
			p.logger.Warn(err)
			return nil, err
		}
	}

	return existing, nil
}

// Reindex is a method that performs the reindexing operation for the project.
// It updates the index of the project to reflect any changes made to its files.
// The function returns an error if any error occurs during the reindexing process.
func (p *Project) Reindex() error {
	return nil
}

func (p *Project) Shutdown(ctx context.Context) error {
	time.Sleep(5 * time.Second)
	if err := p.indexManager.Close(); err != nil {
		return errors.Wrap(err, "failed to close index manager")
	}

	if err := p.indexedGraph.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to close graph")
	}

	return p.ds.Close()
}
