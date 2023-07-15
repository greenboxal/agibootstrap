package codex

import (
	"go/token"
	"os"
	"path"

	"github.com/ipfs/go-datastore"
	badger "github.com/ipfs/go-ds-badger"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/fti"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/graphindex"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/graphstore"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/tasks"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
)

type Platform struct {
	logger *zap.SugaredLogger

	indexedGraph *graphstore.IndexedGraph
	indexManager *graphindex.Manager

	ds datastore.Batching
	fs repofs.FS

	repo *fti.Repository
	tm   *tasks.Manager
	lm   *thoughtdb.Repo

	langRegistry *project.FileTypeProvider
	fset         *token.FileSet

	databasePath string
}

func NewPlatform(rootPath string, databasePath string) (*Platform, error) {
	rootFs, err := repofs.NewFS(rootPath)

	if err != nil {
		return nil, err
	}

	repo, err := fti.NewRepository(rootPath)

	if err != nil {
		return nil, err
	}

	dsOpts := badger.DefaultOptions

	dsPath := path.Join(databasePath, "ds")

	if err := os.MkdirAll(dsPath, 0755); err != nil {
		return nil, errors.Wrap(err, "failed to create datastore directory")
	}

	ds, err := badger.NewDatastore(dsPath, &dsOpts)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create datastore")
	}

	p := &Platform{
		logger: logging.GetLogger("platform"),

		ds:   ds,
		fs:   rootFs,
		repo: repo,

		databasePath: databasePath,

		fset: token.NewFileSet(),
	}

	p.tm = tasks.NewManager()
	p.lm = thoughtdb.NewRepo(p.indexedGraph)

	return p, nil
}
