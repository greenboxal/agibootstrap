package singularity

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"

	agents2 "github.com/greenboxal/agibootstrap/pkg/gpt/agents"
	"github.com/greenboxal/agibootstrap/pkg/gpt/agents/profiles"
	"github.com/greenboxal/agibootstrap/pkg/gpt/featureextractors"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type Singularity struct {
	psi.NodeBase

	self agents2.Agent

	world      *agents2.Colab
	worldState *WorldState
	scheduler  *agents2.RoundRobinScheduler
}

func NewSingularity(lm *thoughtdb.Repo) (*Singularity, error) {
	var err error

	s := &Singularity{
		worldState: NewWorldState(),
		scheduler:  &agents2.RoundRobinScheduler{},
	}

	s.Init(s)

	globalLog := lm.CreateBranch()

	allProfiles := []*agents2.Profile{profiles.SingularityProfile}
	allProfiles = append(allProfiles, profiles.MajorArcanas...)

	allAgents := make([]agents2.Agent, len(allProfiles))

	for i, profile := range allProfiles {
		a := &agents2.AgentBase{}

		al := lm.CreateBranch()

		a.Init(a, profile, nil, al, s.worldState)

		allAgents[i] = a
	}

	s.world, err = agents2.NewColab(
		s.worldState,
		globalLog,
		s.scheduler,
		allAgents[0],
		allAgents[1:]...,
	)

	if err != nil {
		return nil, err
	}

	s.world.SetParent(s)

	return s, nil
}

func (s *Singularity) Router() agents2.Router         { return s.world.Router() }
func (s *Singularity) WorldState() agents2.WorldState { return s.worldState }

func (s *Singularity) Step(ctx context.Context) ([]*thoughtdb.Thought, error) {
	s.worldState.Cycle++

	s.Router().ResetOutbox()

	if err := s.Router().RouteIncomingMessages(ctx); err != nil {
		return nil, err
	}

	plan := agents2.GetState(s.worldState, profiles.CtxPlannerPlan)

	if len(plan.Steps) == 0 {
		if err := s.runSteps(
			ctx,
			profiles.PairProfile.Name,
			profiles.SingularityProfile.Name,
			profiles.DirectorProfile.Name,
			profiles.ManagerProfile.Name,
			profiles.PlannerProfile.Name,
			profiles.LibrarianProfile.Name,
			profiles.JournalistProfile.Name,
		); err != nil {
			return s.Router().OutgoingMessages(), err
		}
	}

	obj, err := featureextractors.QueryObjective(ctx, s.self.History())

	if err != nil {
		return s.Router().OutgoingMessages(), err
	}

	agents2.SetState(s.worldState, profiles.CtxGlobalObjective, obj)

	for {
		progress := agents2.GetState(s.worldState, profiles.CtxGoalStatus)

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
	s.worldState.Time = time.Now()

	kvJson, err := json.Marshal(s.worldState.KV)

	if err != nil {
		return err
	}

	availableMsg := "Available profiles:\n"

	for _, a := range s.world.Members() {
		p := a.Profile()

		availableMsg += fmt.Sprintf("  - **%s:** %s\n", p.Name, p.Description)
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
`, s.worldState.Epoch, s.worldState.Cycle, s.worldState.Step, s.worldState.Time.Format("2006/01/02 - 15:04:05"), kvJson))),

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
