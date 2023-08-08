package jobs

import "github.com/greenboxal/agibootstrap/pkg/psi"

type Job struct {
	psi.NodeBase

	Name    string   `json:"name"`
	Handler psi.Path `json:"handler"`
}

var _ psi.Node = (*Job)(nil)

func (j *Job) PsiNodeName() string { return j.Name }
