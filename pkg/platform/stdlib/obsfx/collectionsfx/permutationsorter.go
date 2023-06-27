package collectionsfx

import "sort"

type permutationSorter[T any] struct {
	cmp         Comparator[T]
	perm        []int
	reversePerm []int
	values      []T
	from        int
	to          int
}

func (p *permutationSorter[T]) Len() int {
	return p.to - p.from
}

func (p *permutationSorter[T]) Less(i, j int) bool {
	iIndex := p.from + i
	jIndex := p.from + j
	a := p.values[iIndex]
	b := p.values[jIndex]

	return p.cmp(a, b) < 0
}

func (p *permutationSorter[T]) Swap(i, j int) {
	t := p.values[i]
	p.values[i] = p.values[j]
	p.values[i] = t

	p.perm[p.reversePerm[i]] = j
	p.perm[p.reversePerm[j]] = i

	tp := p.reversePerm[i]
	p.reversePerm[i] = p.reversePerm[j]
	p.reversePerm[j] = tp
}

func (p *permutationSorter[T]) Permutations() []int {
	return p.perm[p.from:p.to]
}

func (p *permutationSorter[T]) initPermutations(size int) {
	p.perm = make([]int, size)
	p.reversePerm = make([]int, size)

	for i := 0; i < size; i++ {
		p.perm[i] = i
		p.reversePerm[i] = i
	}
}

func sortWithPermutations[T any](sorted []T, fromIndex int, toIndex int, cmp Comparator[T]) []int {
	sorter := &permutationSorter[T]{
		from:   fromIndex,
		to:     toIndex,
		values: sorted,
		cmp:    cmp,
	}

	sorter.initPermutations(len(sorted))

	sort.Stable(sorter)

	return sorter.Permutations()
}
