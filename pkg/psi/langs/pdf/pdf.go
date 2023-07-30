package pdf

import (
	"io"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"

	"github.com/greenboxal/agibootstrap/pkg/platform/project"
)

type Node interface {
	project.AstNode
}

type NodeBase struct {
	project.AstNodeBase
}

func Parse(rs io.ReadSeeker) error {
	cfg := model.NewDefaultConfiguration()
	ctx, err := api.ReadContext(rs, cfg)

	if err != nil {
		return err
	}

	ctx.Page
}
