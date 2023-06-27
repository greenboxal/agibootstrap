package obsfx

type SelectStep func(any) Observable

type SelectBinding[T any] struct {
	ObservableValueBase[T]

	steps []SelectStep
	props []Observable
	deps  []Observable
	valid bool
	value T
}

func NewSelectBinding[T any](root Observable, steps []SelectStep) *SelectBinding[T] {
	sb := &SelectBinding[T]{}

	sb.steps = steps
	sb.props = make([]Observable, len(steps)+1)
	sb.props[0] = root

	root.AddListener(sb)

	return sb
}

func (s *SelectBinding[T]) observableValue() ObservableValue[T] {
	for i := 0; i < len(s.props)-1; i++ {
		prop := s.props[i].(ObservableValue[Observable])
		value := prop.Value()

		s.props[i+1] = value
		s.props[i+1].AddListener(s)
	}

	s.updateDependencies()

	return s.props[len(s.props)-1].(ObservableValue[T])
}

func (o *SelectBinding[T]) RawValue() any {
	return o.Value()
}
func (s *SelectBinding[T]) Value() T {
	if !s.valid {
		obs := s.observableValue()

		if obs == nil {
			return EmptyValue[T]()
		}

		s.value = obs.Value()
		s.valid = true
	}

	return s.value
}

func (s *SelectBinding[T]) Dependencies() []Observable {
	if len(s.deps) == 0 {
		s.updateDependencies()
	}

	return s.deps
}

func (s *SelectBinding[T]) Invalidate() {
	s.markInvalid()
}

func (s *SelectBinding[T]) IsValid() bool {
	return s.valid
}

func (s *SelectBinding[T]) Close() {
	s.unregisterDependencies()
}

func (s *SelectBinding[T]) unregisterDependencies() {
	for i, dep := range s.props {
		if i == 0 || dep == nil {
			continue
		}

		dep.RemoveListener(s)

		s.props[i] = nil
	}

	s.updateDependencies()
}
func (s *SelectBinding[T]) updateDependencies() {
	s.deps = s.deps[0:0]

	for _, v := range s.props {
		if v == nil {
			continue
		}

		s.deps = append(s.deps, v)
	}
}

func (s *SelectBinding[T]) markInvalid() {
	if s.valid {
		s.valid = false

		s.unregisterDependencies()

		s.ObservableValueBase.OnInvalidated(s)
	}
}
