package simworld

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Agent struct {
	psi.NodeBase

	World *World

	Position psi.Node
	Focus    psi.Node
}

type InteractionRequest struct {
	Target psi.Node `json:"target"`
}

func (a *Agent) Interact(ctx context.Context, req *InteractionRequest) error {

	return nil
}

type InspectRequest struct {
	Target psi.Node `json:"target"`
}

func (a *Agent) Inspect(ctx context.Context, req *InspectRequest) error {

	return nil
}

type StepRequest struct {
	Step int    `json:"step"`
	Time uint64 `json:"time"`
}

func (a *Agent) Step(ctx context.Context, step *StepRequest) error {

	return nil
}

func (a *Agent) MoveTo(ctx context.Context, node psi.Node) error {
	// TODO: Find a physical path to the node and move to it.

	a.Position = node

	return nil
}

func (a *Agent) LookAt(ctx context.Context, node psi.Node) error {
	// TODO: Find a mental path to the node and look at it.

	a.Focus = node

	return nil
}

type PredictionRequest struct {
	Raw string `json:"raw"`
}

type Prediction struct {
	Raw string `json:"raw"`
}

func (a *Agent) Predict(ctx context.Context, req *PredictionRequest) (*Prediction, error) {
	var prediction Prediction

	return &prediction, nil
}

func (a *Agent) buildActionsFor(node psi.Node) []*openai.FunctionDefine {
	var functions []*openai.FunctionDefine

	for _, iface := range node.PsiNodeType().Interfaces() {
		for _, action := range iface.Interface().Actions() {
			var args openai.FunctionParams

			args.Type = "object"

			if action.RequestType != nil {
				def := action.RequestType.JsonSchema()
				args.Type = def.Type
				args.Properties = def.Properties
				args.Required = def.Required
			}

			functions = append(functions, &openai.FunctionDefine{
				Name:        fmt.Sprintf("%s.%s", iface.Interface().Name(), action.Name),
				Description: "",
				Parameters:  &args,
			})
		}
	}

	return functions
}
