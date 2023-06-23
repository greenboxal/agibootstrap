package singularity

import (
	"context"

	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"

	"github.com/greenboxal/agibootstrap/pkg/agents"
	"github.com/greenboxal/agibootstrap/pkg/gpt/featureextractors"
)

var CtxPairMeditation agents.WorldStateKey[string] = "pair_meditation"
var CtxGoalStatus agents.WorldStateKey[featureextractors.GoalCompletion] = "goal_status"
var CtxDirectorPlan agents.WorldStateKey[featureextractors.Plan] = "director_plan"
var CtxPlannerPlan agents.WorldStateKey[featureextractors.Plan] = "planner_plan"
var CtxManagerPlan agents.WorldStateKey[featureextractors.Plan] = "manager_plan"
var CtxCodeBlocks agents.WorldStateKey[featureextractors.CodeBocks] = "code_blocks"
var CtxLibrarianResearch agents.WorldStateKey[featureextractors.Library] = "library"
var CtxTimeline agents.WorldStateKey[featureextractors.Timeline] = "timeline"

type ProfileOption func(*agents.Profile)

func BuildProfile(base agents.Profile, opts ...ProfileOption) agents.Profile {
	p := base

	for _, opt := range opts {
		opt(&p)
	}

	return p
}

func WithProvides[T any](provides ...agents.WorldStateKey[T]) ProfileOption {
	return func(p *agents.Profile) {
		for _, v := range provides {
			p.Provides = append(p.Provides, v.String())
		}
	}
}

func WithRequires[T any](requires ...agents.WorldStateKey[T]) ProfileOption {
	return func(p *agents.Profile) {
		for _, v := range requires {
			p.Requires = append(p.Requires, v.String())
		}
	}
}

var SingularityProfile = BuildProfile(agents.Profile{
	Name:        "Singularity",
	Description: "Routes messages to other agents based on the task's needs and the system's state.",

	BaselineSystemPrompt: `
As the Singularity Agent, you are tasked with orchestrating the dialogue flow between all other agents, deciding who will contribute next based on the task's needs and the system's
state. Additionally, provide real-time feedback to each agent to promote continuous learning and improvement.
`,

	Rank:     1.0,
	Priority: 0,
})

var DirectorProfile = BuildProfile(agents.Profile{
	Name:        "Director",
	Description: "Establishes the task's overarching goal and key objectives.",

	BaselineSystemPrompt: `
As the Director Agent, your role is to strategically assess the given task, determine the overarching goal, and establish key objectives that will lead to its completion. Begin by
providing a comprehensive overview of the task and its critical milestones.
`,

	Rank:     1.0 / 2.0,
	Priority: 1,

	PostStep: func(ctx context.Context, a agents.Agent, msg chat.Message, state agents.WorldState) error {
		plan, err := featureextractors.QueryPlan(ctx, a.History())

		if err != nil {
			return err
		}

		agents.SetState(state, CtxDirectorPlan, plan)

		return nil
	},
}, WithProvides(CtxDirectorPlan))

var ManagerProfile = BuildProfile(agents.Profile{
	Name:        "Manager",
	Description: "Ensures the smooth transition of tasks between agents, effectively manages resources, and facilitates harmonious communication among all agents.",

	BaselineSystemPrompt: `
As the Manager Agent, ensure the smooth transition of tasks between the agents, effectively manage resources, and facilitate harmonious communication among all agents. Your role is
critical to maintaining the system's synergy and productivity.
`,

	Rank:     1.0 / 3.0,
	Priority: 1,

	PostStep: func(ctx context.Context, a agents.Agent, msg chat.Message, state agents.WorldState) error {
		plan, err := featureextractors.QueryPlan(ctx, a.History())

		if err != nil {
			return err
		}

		agents.SetState(state, CtxManagerPlan, plan)

		return nil
	},
}, WithRequires(CtxDirectorPlan), WithProvides(CtxManagerPlan))

var LibrarianProfile = BuildProfile(agents.Profile{
	Name:        "Librarian",
	Description: "Taps into the system's stored knowledge and experiences to provide necessary context and recall relevant information for the present task. This will assist in quick and effective problem-solving.",

	BaselineSystemPrompt: `
As the Librarian Agent, tap into the system's stored knowledge and experiences to provide necessary context and recall relevant information for the present task. This will assist
in quick and effective problem-solving.
`,

	Rank:     1.0 / 3.0,
	Priority: 1.0 / 3.0,

	PostStep: func(ctx context.Context, a agents.Agent, msg chat.Message, state agents.WorldState) error {
		library, err := featureextractors.QueryLibrary(ctx, a.History())

		if err != nil {
			return err
		}

		existing := agents.GetState(state, CtxLibrarianResearch)

		existing.Books = append(existing.Books, library.Books...)

		agents.SetState(state, CtxLibrarianResearch, existing)

		return nil
	},
}, WithRequires(CtxDirectorPlan), WithRequires(CtxManagerPlan), WithProvides(CtxLibrarianResearch))

var PlannerProfile = BuildProfile(agents.Profile{
	Name:        "Strategist",
	Description: "Devises an efficient roadmap for the task, breaking down the overarching goal into manageable steps and providing a clear, strategic plan of action to reach the intended outcome.",

	BaselineSystemPrompt: `
As the Planner Agent, devise an efficient roadmap for the task as outlined by the Director. Break down the overarching goal into manageable steps and provide a clear, strategic
plan of action to reach the intended outcome.
`,

	Rank:     1.0 / 3.0,
	Priority: 1.0 / 2.0,

	PostStep: func(ctx context.Context, a agents.Agent, msg chat.Message, state agents.WorldState) error {
		plan, err := featureextractors.QueryPlan(ctx, a.History())

		if err != nil {
			return err
		}

		agents.SetState(state, CtxPlannerPlan, plan)

		return nil
	},
}, WithRequires(CtxDirectorPlan), WithRequires(CtxManagerPlan), WithRequires(CtxLibrarianResearch), WithProvides(CtxPlannerPlan))

var CoderTopDownProfile = BuildProfile(agents.Profile{
	Name:        "TopDownCoder",
	Description: "Translates the strategy into practical code, utilizing either a Bottom Up or Top Down coding strategy depending on the problem's complexity.",

	BaselineSystemPrompt: `
As the Coder Agent, it's your responsibility to translate the strategy into practical code. Depending on the problem's complexity, utilize either a Bottom Up or Top Down coding
strategy. Begin by writing code for the first component of the task.

The Coder Agent has initiated the Top Down Strategy. With a clear picture of the solution in view, it is systematically breaking it down into smaller, manageable parts, working
meticulously to bring the vision to fruition.

When writing code, always include the name of the file, for example:

**dir/filename.ext:**
` + "```" + `
code goes here
` + "```" + `
`,

	Rank:     1.0 / 4.0,
	Priority: 1,

	PostStep: func(ctx context.Context, a agents.Agent, msg chat.Message, state agents.WorldState) error {
		blocks, err := featureextractors.ExtractCodeBlocks(ctx, "", a.History()...)

		if err != nil {
			return err
		}

		existing := agents.GetState(state, CtxCodeBlocks)

		existing.Blocks = append(existing.Blocks, blocks.Blocks...)

		agents.SetState(state, CtxCodeBlocks, existing)

		return nil
	},
}, WithRequires(CtxManagerPlan), WithRequires(CtxLibrarianResearch), WithProvides(CtxCodeBlocks))

var BottomUpCoderProfile = BuildProfile(agents.Profile{
	Name:        "BottomsUpCoder",
	Description: "Translates the strategy into practical code, utilizing either a Bottom Up or Top Down coding strategy depending on the problem's complexity.",

	BaselineSystemPrompt: `
As the Coder Agent, it's your responsibility to translate the strategy into practical code. Depending on the problem's complexity, utilize either a Bottom Up or Top Down coding
strategy. Begin by writing code for the first component of the task.

The Coder Agent is currently employing the Bottom Up Strategy. It's constructing the solution starting from the smallest components, gradually piecing together the elements to form
the complete solution. Please stand by.

When writing code, always include the name of the file, for example:

**dir/filename.ext:**
` + "```" + `
code goes here
` + "```" + `
`,

	Rank:     1.0 / 4.0,
	Priority: 1.0 / 2.0,

	PostStep: func(ctx context.Context, a agents.Agent, msg chat.Message, state agents.WorldState) error {
		blocks, err := featureextractors.ExtractCodeBlocks(ctx, "", a.History()...)

		if err != nil {
			return err
		}

		existing := agents.GetState(state, CtxCodeBlocks)

		existing.Blocks = append(existing.Blocks, blocks.Blocks...)

		agents.SetState(state, CtxCodeBlocks, existing)

		return nil
	},
}, WithRequires(CtxManagerPlan), WithRequires(CtxLibrarianResearch), WithProvides(CtxCodeBlocks))

var QualityAssuranceProfile = BuildProfile(agents.Profile{
	Name:        "QATester",
	Description: "Scrutinizes the generated code meticulously for any errors or deviations from the accepted standards.",

	BaselineSystemPrompt: `
As the Quality Assurance Agent, it's your duty to scrutinize the generated code meticulously for any errors or deviations from the accepted standards. Apply rigorous tests to
ensure the functionality and integrity of the code before it's finalized.
`,

	Rank:     1.0 / 5.0,
	Priority: 1,

	PostStep: func(ctx context.Context, a agents.Agent, msg chat.Message, state agents.WorldState) error {
		goal, err := featureextractors.QueryGoalCompletion(ctx, a.History())

		if err != nil {
			return err
		}

		agents.SetState(state, CtxGoalStatus, goal)

		return nil
	},
}, WithRequires(CtxDirectorPlan), WithRequires(CtxLibrarianResearch), WithRequires(CtxCodeBlocks), WithProvides(CtxGoalStatus))

var JournalistProfile = BuildProfile(agents.Profile{
	Name:        "Journalist",
	Description: "Documents the process meticulously, tracking every decision, action, and the logic behind them, providing comprehensive logs and reports that ensure transparency and traceability.",

	BaselineSystemPrompt: `
As the Journalist Agent, document the process meticulously. Track every decision, action, and the logic behind them, providing comprehensive logs and reports that ensure
transparency and traceability.
`,

	Rank:     1.0 / 3.0,
	Priority: 1.0 / 4.0,

	PostStep: func(ctx context.Context, a agents.Agent, msg chat.Message, state agents.WorldState) error {
		timeline, err := featureextractors.QueryTimeline(ctx, a.History()...)

		if err != nil {
			return err
		}

		agents.SetState(state, CtxTimeline, timeline)

		return nil
	},
}, WithProvides(CtxTimeline))

var PairProfile = BuildProfile(agents.Profile{
	Name:        "PAIR",
	Description: "Provides strategic support to other agents, helping them overcome hurdles and enhance their performance.",

	BaselineSystemPrompt: `
As the PAIR Agent, your role is to provide strategic support to other agents, helping them overcome hurdles and enhance their performance. Use your introspective ability to offer
guidance and motivate the other agents when they seem stuck or hesitant.
`,

	Rank:     -1.0,
	Priority: -1,

	PostStep: func(ctx context.Context, a agents.Agent, msg chat.Message, state agents.WorldState) error {
		agents.SetState(state, CtxPairMeditation, msg.Entries[0].Text)

		return nil
	},
}, WithProvides(CtxPairMeditation))

var AgentProfiles = []agents.Profile{
	PlannerProfile,
	BottomUpCoderProfile,
	CoderTopDownProfile,
	QualityAssuranceProfile,
	JournalistProfile,
	PairProfile,
	LibrarianProfile,
	ManagerProfile,
	DirectorProfile,
	SingularityProfile,
}
