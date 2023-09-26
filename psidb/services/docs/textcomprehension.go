package docs

import (
	"context"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules/gpt"
	"github.com/greenboxal/agibootstrap/psidb/modules/gpt/autoform"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type TextComprehensionStepRequest struct {
	CurrentSteps int `json:"currentSteps,omitempty"`
	MaxSteps     int `json:"maxSteps,omitempty"`
}

type ITextComprehension interface {
	Step(ctx context.Context, req *TextComprehensionStepRequest) error
	RunToCompletion(ctx context.Context) error
}

type TextComprehension struct {
	psi.NodeBase

	Name string `json:"name"`

	Content        string `json:"content"`
	TokensPerChunk int    `json:"linesPerChunk"`
	TokensPerPage  int    `json:"linesPerPage"`

	CurrentIndex int      `json:"currentIndex"`
	Observations []string `json:"observations"`
}

var TextComprehensionInterface = psi.DefineNodeInterface[ITextComprehension]()
var TextComprehensionType = psi.DefineNodeType[*TextComprehension](psi.WithInterfaceFromNode(TextComprehensionInterface))
var _ ITextComprehension = (*TextComprehension)(nil)

func (t *TextComprehension) PsiNodeName() string { return t.Name }

func (t *TextComprehension) Init(self psi.Node) {
	t.NodeBase.Init(self, psi.WithNodeType(TextComprehensionType))

	if t.TokensPerChunk == 0 {
		t.TokensPerChunk = 128
	}

	if t.TokensPerPage == 0 {
		t.TokensPerPage = 128
	}
}

func (t *TextComprehension) Step(ctx context.Context, req *TextComprehensionStepRequest) error {
	tc := autoform.NewTextComprehension(gpt.GlobalModelTokenizer, t.Content, t.TokensPerChunk, t.TokensPerPage)

	tc.CurrentIndex = t.CurrentIndex
	tc.Observations = t.Observations

	if err := tc.Step(ctx); err != nil {
		return err
	}

	t.CurrentIndex = tc.CurrentIndex
	t.Observations = tc.Observations
	t.Invalidate()

	if !tc.IsComplete() && (req.CurrentSteps < req.MaxSteps || req.MaxSteps == -1) {
		req.CurrentSteps++

		if err := coreapi.DispatchSelf(ctx, t, TextComprehensionInterface, "Step", req); err != nil {
			return err
		}
	}

	return t.Update(ctx)
}

func (t *TextComprehension) RunToCompletion(ctx context.Context) error {
	return coreapi.DispatchSelf(ctx, t, TextComprehensionInterface, "Step", &TextComprehensionStepRequest{
		MaxSteps: -1,
	})
}
