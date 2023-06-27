package collectionsfx

type ListTransformation[V, T any] interface {
	ObservableList[T]

	Source() ObservableList[V]

	GetSourceIndex(index int) int
	GetViewIndex(index int) int
}
