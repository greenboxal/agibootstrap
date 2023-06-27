package agents

type Scheduler interface {
	NextSpeaker(ctx AgentContext, candidates ...Agent) (Agent, error)
}

type RoundRobinScheduler struct {
	current int
}

func (r *RoundRobinScheduler) NextSpeaker(ctx AgentContext, candidates ...Agent) (Agent, error) {
	if len(candidates) == 0 {
		return nil, nil
	}

	if r.current >= len(candidates) {
		r.current = 0
	}

	next := candidates[r.current]
	r.current++

	return next, nil
}
