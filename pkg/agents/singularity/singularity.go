package singularity

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"

	agents "github.com/greenboxal/agibootstrap/pkg/agents"
	"github.com/greenboxal/agibootstrap/pkg/gpt/featureextractors"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Singularity struct {
	psi.NodeBase

	self *Agent

	world      *Colab
	worldState *WorldState
	scheduler  *RoundRobinScheduler
}

func NewSingularity(lm *thoughtstream.Manager) (*Singularity, error) {
	globalLog, err := lm.GetOrCreateStream("GLOBAL")

	if err != nil {
		panic(err)
	}

	s := &Singularity{
		worldState: NewWorldState(),
		scheduler:  &RoundRobinScheduler{},
	}

	s.Init(s, "")

	s.world, err = NewColab(s.worldState, globalLog, s.scheduler, SingularityProfile, MajorArcanas...)

	if err != nil {
		return nil, err
	}

	s.world.SetParent(s)

	return s, nil
}

func (s *Singularity) Router() *Router               { return s.world.Router() }
func (s *Singularity) WorldState() agents.WorldState { return s.worldState }

func (s *Singularity) Step(ctx context.Context) ([]*thoughtstream.Thought, error) {
	s.worldState.Cycle++

	s.Router().ResetOutbox()

	plan := agents.GetState(s.worldState, CtxPlannerPlan)

	if len(plan.Steps) == 0 {
		if err := s.runSteps(
			ctx,
			PairProfile.Name,
			SingularityProfile.Name,
			DirectorProfile.Name,
			ManagerProfile.Name,
			PlannerProfile.Name,
			LibrarianProfile.Name,
			JournalistProfile.Name,
		); err != nil {
			return s.Router().OutgoingMessages(), err
		}
	}

	obj, err := featureextractors.QueryObjective(ctx, s.self.log.Messages())

	if err != nil {
		return s.Router().OutgoingMessages(), err
	}

	agents.SetState(s.worldState, CtxGlobalObjective, obj)

	for {
		progress := agents.GetState(s.worldState, CtxGoalStatus)

		if progress.Completed {
			break
		}

		if err := s.world.Step(ctx); err != nil {
			return s.Router().OutgoingMessages(), err
		}
	}

	return s.Router().OutgoingMessages(), nil
}

func (s *Singularity) runSteps(ctx context.Context, steps ...string) error {
	for _, step := range steps {
		if err := s.doStep(ctx, step); err != nil {
			return err
		}
	}

	return nil
}

func (s *Singularity) doStep(ctx context.Context, profileName string) error {
	s.worldState.Step++

	kvJson, err := json.Marshal(s.worldState.KV)

	if err != nil {
		return err
	}

	availableMsg := "Available profiles:\n"

	for _, a := range s.world.Members() {
		availableMsg += fmt.Sprintf("  - **%s:** %s\n", a.Name, a.Description)
	}

	s.worldState.SystemMessages = []chat.Message{
		chat.Compose(chat.Entry(msn.RoleSystem, fmt.Sprintf(`
===
**System Epoch:** %d:%d.%d
**System Clock:** %s
**Global State:**
`+"```json"+`
%s
`+"```"+`
===
`, s.worldState.Epoch, s.worldState.Cycle, s.worldState.Step, time.Now().Format("2006/01/02 - 15:04:05"), kvJson))),

		chat.Compose(chat.Entry(msn.RoleSystem, availableMsg)),
	}

	if err := s.world.StepWith(ctx, profileName); err != nil {
		return err
	}

	data, err := json.Marshal(s.worldState)

	if err != nil {
		return err
	}

	err = os.WriteFile("/tmp/agib-agent-logs/state.json", data, 0644)

	if err != nil {
		return err
	}

	return nil
}
