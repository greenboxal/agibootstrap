package profiles

import (
	agents2 "github.com/greenboxal/agibootstrap/pkg/gpt/agents"
	"github.com/greenboxal/agibootstrap/pkg/gpt/featureextractors"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
)

var CtxPairMeditation agents2.WorldStateKey[string] = "pair_meditation"
var CtxGoalStatus agents2.WorldStateKey[featureextractors.GoalCompletion] = "goal_status"
var CtxGlobalObjective agents2.WorldStateKey[featureextractors.Objective] = "global_objective"
var CtxDirectorPlan agents2.WorldStateKey[featureextractors.Plan] = "director_plan"
var CtxPlannerPlan agents2.WorldStateKey[featureextractors.Plan] = "planner_plan"
var CtxManagerPlan agents2.WorldStateKey[featureextractors.Plan] = "manager_plan"
var CtxCodeBlocks agents2.WorldStateKey[featureextractors.CodeBocks] = "code_blocks"
var CtxLibrarianResearch agents2.WorldStateKey[featureextractors.Library] = "library"
var CtxTimeline agents2.WorldStateKey[featureextractors.Timeline] = "timeline"

type ProfileOption func(*agents2.Profile)

func BuildProfile(base agents2.Profile, opts ...ProfileOption) *agents2.Profile {
	p := base

	for _, opt := range opts {
		opt(&p)
	}

	return &p
}

func WithProvides[T any](provides ...agents2.WorldStateKey[T]) ProfileOption {
	return func(p *agents2.Profile) {
		for _, v := range provides {
			p.Provides = append(p.Provides, v.String())
		}
	}
}

func WithRequires[T any](requires ...agents2.WorldStateKey[T]) ProfileOption {
	return func(p *agents2.Profile) {
		for _, v := range requires {
			p.Requires = append(p.Requires, v.String())
		}
	}
}

var DirectorProfile = BuildProfile(agents2.Profile{
	Name:        "Director",
	Description: "Establishes the task's overarching goal and key objectives.",

	BaselineSystemPrompt: `
UncheckedCast the Director Agent, your role is to strategically assess the given task, determine the overarching goal, and establish key objectives that will lead to its completion. Begin by
providing a comprehensive overview of the task and its critical milestones.
`,

	Rank:     1.0 / 2.0,
	Priority: 1,

	PostStep: func(ctx agents2.AgentContext, msg *thoughtdb.Thought) error {
		plan, err := featureextractors.QueryPlan(ctx.Context(), ctx.Agent().History())

		if err != nil {
			return err
		}

		plan.SetParent(ctx.Branch())
		agents2.SetState(ctx.WorldState(), CtxDirectorPlan, plan)

		return nil
	},
}, WithProvides(CtxDirectorPlan))

var ManagerProfile = BuildProfile(agents2.Profile{
	Name:        "Manager",
	Description: "Ensures the smooth transition of tasks between profiles, effectively manages resources, and facilitates harmonious communication among all profiles.",

	BaselineSystemPrompt: `
UncheckedCast the Manager Agent, ensure the smooth transition of tasks between the profiles, effectively manage resources, and facilitate harmonious communication among all profiles. Your role is
critical to maintaining the system's synergy and productivity.
`,

	Rank:     1.0 / 3.0,
	Priority: 1,

	PostStep: func(ctx agents2.AgentContext, msg *thoughtdb.Thought) error {
		plan, err := featureextractors.QueryPlan(ctx.Context(), ctx.Agent().History())

		if err != nil {
			return err
		}

		plan.SetParent(ctx.Branch())
		agents2.SetState(ctx.WorldState(), CtxManagerPlan, plan)

		return nil
	},
}, WithRequires(CtxDirectorPlan), WithProvides(CtxManagerPlan))

var LibrarianProfile = BuildProfile(agents2.Profile{
	Name:        "Librarian",
	Description: "Taps into the system's stored knowledge and experiences to provide necessary context and recall relevant information for the present task. This will assist in quick and effective problem-solving.",

	BaselineSystemPrompt: `
UncheckedCast the Librarian Agent, tap into the system's stored knowledge and experiences to provide necessary context and recall relevant information for the present task. This will assist
in quick and effective problem-solving.
`,

	Rank:     1.0 / 3.0,
	Priority: 1.0 / 3.0,

	PostStep: func(ctx agents2.AgentContext, msg *thoughtdb.Thought) error {
		library, err := featureextractors.QueryLibrary(ctx.Context(), ctx.Agent().History())

		if err != nil {
			return err
		}

		library.SetParent(ctx.Branch())

		existing := agents2.GetState(ctx.WorldState(), CtxLibrarianResearch)

		existing.Books = append(existing.Books, library.Books...)

		agents2.SetState(ctx.WorldState(), CtxLibrarianResearch, existing)

		return nil
	},
}, WithRequires(CtxDirectorPlan), WithRequires(CtxManagerPlan), WithProvides(CtxLibrarianResearch))

var PlannerProfile = BuildProfile(agents2.Profile{
	Name:        "Strategist",
	Description: "Devises an efficient roadmap for the task, breaking down the overarching goal into manageable steps and providing a clear, strategic plan of action to reach the intended outcome.",

	BaselineSystemPrompt: `
UncheckedCast the Planner Agent, devise an efficient roadmap for the task as outlined by the Director. Break down the overarching goal into manageable steps and provide a clear, strategic
plan of action to reach the intended outcome.
`,

	Rank:     1.0 / 3.0,
	Priority: 1.0 / 2.0,

	PostStep: func(ctx agents2.AgentContext, msg *thoughtdb.Thought) error {
		plan, err := featureextractors.QueryPlan(ctx.Context(), ctx.Agent().History())

		if err != nil {
			return err
		}

		plan.SetParent(ctx.Branch())

		agents2.SetState(ctx.WorldState(), CtxPlannerPlan, plan)

		return nil
	},
}, WithRequires(CtxDirectorPlan), WithRequires(CtxManagerPlan), WithRequires(CtxLibrarianResearch), WithProvides(CtxPlannerPlan))

var CoderTopDownProfile = BuildProfile(agents2.Profile{
	Name:        "TopDownCoder",
	Description: "Translates the strategy into practical code, utilizing either a Bottom Up or Top Down coding strategy depending on the problem's complexity.",

	BaselineSystemPrompt: `
UncheckedCast the Coder Agent, it's your responsibility to translate the strategy into practical code. Depending on the problem's complexity, utilize either a Bottom Up or Top Down coding
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

	PostStep: func(ctx agents2.AgentContext, msg *thoughtdb.Thought) error {
		blocks, err := featureextractors.ExtractCodeBlocks(ctx.Context(), "", ctx.Agent().History()...)

		if err != nil {
			return err
		}

		blocks.SetParent(ctx.Branch())

		existing := agents2.GetState(ctx.WorldState(), CtxCodeBlocks)

		existing.Blocks = append(existing.Blocks, blocks.Blocks...)

		agents2.SetState(ctx.WorldState(), CtxCodeBlocks, existing)

		return nil
	},
}, WithRequires(CtxManagerPlan), WithRequires(CtxLibrarianResearch), WithProvides(CtxCodeBlocks))

var BottomUpCoderProfile = BuildProfile(agents2.Profile{
	Name:        "BottomsUpCoder",
	Description: "Translates the strategy into practical code, utilizing either a Bottom Up or Top Down coding strategy depending on the problem's complexity.",

	BaselineSystemPrompt: `
UncheckedCast the Coder Agent, it's your responsibility to translate the strategy into practical code. Depending on the problem's complexity, utilize either a Bottom Up or Top Down coding
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

	PostStep: func(ctx agents2.AgentContext, msg *thoughtdb.Thought) error {
		blocks, err := featureextractors.ExtractCodeBlocks(ctx.Context(), "", ctx.Agent().History()...)

		if err != nil {
			return err
		}

		blocks.SetParent(ctx.Branch())

		existing := agents2.GetState(ctx.WorldState(), CtxCodeBlocks)

		existing.Blocks = append(existing.Blocks, blocks.Blocks...)

		agents2.SetState(ctx.WorldState(), CtxCodeBlocks, existing)

		return nil
	},
}, WithRequires(CtxManagerPlan), WithRequires(CtxLibrarianResearch), WithProvides(CtxCodeBlocks))

var QualityAssuranceProfile = BuildProfile(agents2.Profile{
	Name:        "QATester",
	Description: "Scrutinizes the generated code meticulously for any errors or deviations from the accepted standards.",

	BaselineSystemPrompt: `
UncheckedCast the Quality Assurance Agent, it's your duty to scrutinize the generated code meticulously for any errors or deviations from the accepted standards. Apply rigorous tests to
ensure the functionality and integrity of the code before it's finalized.
`,

	Rank:     1.0 / 5.0,
	Priority: 1,

	PostStep: func(ctx agents2.AgentContext, msg *thoughtdb.Thought) error {
		goal, err := featureextractors.QueryGoalCompletion(ctx.Context(), ctx.Agent().History())

		if err != nil {
			return err
		}

		goal.SetParent(ctx.Branch())
		agents2.SetState(ctx.WorldState(), CtxGoalStatus, goal)

		return nil
	},
}, WithRequires(CtxDirectorPlan), WithRequires(CtxLibrarianResearch), WithRequires(CtxCodeBlocks), WithProvides(CtxGoalStatus))

var JournalistProfile = BuildProfile(agents2.Profile{
	Name:        "Journalist",
	Description: "Documents the process meticulously, tracking every decision, action, and the logic behind them, providing comprehensive logs and reports that ensure transparency and traceability.",

	BaselineSystemPrompt: `
UncheckedCast the Journalist Agent, document the process meticulously. Track every decision, action, and the logic behind them, providing comprehensive logs and reports that ensure
transparency and traceability.
`,

	Rank:     1.0 / 3.0,
	Priority: 1.0 / 4.0,

	PostStep: func(ctx agents2.AgentContext, msg *thoughtdb.Thought) error {
		timeline, err := featureextractors.QueryTimeline(ctx.Context(), ctx.Agent().History()...)

		if err != nil {
			return err
		}

		timeline.SetParent(ctx.Branch())
		agents2.SetState(ctx.WorldState(), CtxTimeline, timeline)

		return nil
	},
}, WithProvides(CtxTimeline))

var PairProfile = BuildProfile(agents2.Profile{
	Name:        "PAIR",
	Description: "Provides strategic support to other profiles, helping them overcome hurdles and enhance their performance.",

	BaselineSystemPrompt: `
UncheckedCast the PAIR Agent, your role is to provide strategic support to other profiles, helping them overcome hurdles and enhance their performance. Use your introspective ability to offer
guidance and motivate the other profiles when they seem stuck or hesitant.
`,

	Rank:     -1.0,
	Priority: -1,

	PostStep: func(ctx agents2.AgentContext, msg *thoughtdb.Thought) error {
		agents2.SetState(ctx.WorldState(), CtxPairMeditation, msg.Text)

		return nil
	},
}, WithProvides(CtxPairMeditation))

var SingularityProfile = BuildProfile(agents2.Profile{
	Name:        "Singularity",
	Description: "Routes messages to other profiles based on the task's needs and the system's state.",

	BaselineSystemPrompt: `
UncheckedCast the Singularity Agent, you are tasked with orchestrating the dialogue flow between all other profiles, deciding who will contribute next based on the task's needs and the system's
state. Additionally, provide real-time feedback to each agent to promote continuous learning and improvement.
`,

	Rank:     1.0,
	Priority: 0,
})

var MajorArcanas = []*agents2.Profile{
	PlannerProfile,
	BottomUpCoderProfile,
	CoderTopDownProfile,
	QualityAssuranceProfile,
	JournalistProfile,
	PairProfile,
	LibrarianProfile,
	ManagerProfile,
	DirectorProfile,
}
