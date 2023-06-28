package thoughtstream

type ForkOptions struct {
}

type ForkOption func(*ForkOptions)

func NewForkOptions(opts ...ForkOption) ForkOptions {
	var fo ForkOptions
	for _, opt := range opts {
		opt(&fo)
	}
	return fo
}

type MergeOptions struct {
}

type MergeOption func(*MergeOptions)

func NewMergeOptions(opts ...MergeOption) MergeOptions {
	var mo MergeOptions
	for _, opt := range opts {
		opt(&mo)
	}
	return mo
}
