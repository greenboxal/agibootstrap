package collectionsfx

import "golang.org/x/exp/slices"

type ListChangeBuilder[T any] struct {
	*ObservableListBase[T]

	List ObservableList[T]

	changeLock        int
	addRemoveChanges  []*listChangeEvent[T]
	updateChanges     []*listChangeEvent[T]
	permutationChange *listChangeEvent[T]
}

func (b *ListChangeBuilder[T]) NextRemove(index int, value T) {
	var last *listChangeEvent[T]

	b.checkState()

	if len(b.addRemoveChanges) > 0 {
		last = b.addRemoveChanges[len(b.addRemoveChanges)-1]
	}

	if last != nil && last.To() == index {
		last.removedSlice = append(last.removedSlice, value)
	} else if last != nil && last.from == index+1 {
		last.from--
		last.to--
		last.removedSlice = append([]T{value}, last.removedSlice...)
	} else {
		b.insertRemoved(index, value)
	}

	if len(b.updateChanges) > 0 {
		upos := b.findSubChange(b.updateChanges, index)

		if upos < 0 {
			upos = ^upos
		} else {
			ch := b.updateChanges[upos]

			if ch.from == ch.to-1 {
				b.updateChanges = slices.Delete(b.updateChanges, upos, upos+1)
			} else {
				ch.to--
				upos++
			}
		}

		for i := upos; i < len(b.updateChanges); i++ {
			b.updateChanges[i].from--
			b.updateChanges[i].to--
		}
	}
}

func (b *ListChangeBuilder[T]) NextRemoveRange(index int, removed []T) {
	for _, v := range removed {
		b.NextRemove(index, v)
	}
}

func (b *ListChangeBuilder[T]) NextAdd(from, to int) {
	var last *listChangeEvent[T]

	b.checkState()

	if len(b.addRemoveChanges) > 0 {
		last = b.addRemoveChanges[len(b.addRemoveChanges)-1]
	}

	count := to - from

	if last != nil && last.To() == from {
		last.to = to
	} else if last != nil && from >= last.from && from < last.to {
		last.to += count
	} else {
		b.insertAdd(from, to)
	}

	if len(b.updateChanges) > 0 {
		upos := b.findSubChange(b.updateChanges, from)

		if upos < 0 {
			upos = ^upos
		} else {
			change := b.updateChanges[upos]

			cev := &listChangeEvent[T]{
				list:    b.List,
				from:    to,
				to:      change.to + to - from,
				updated: true,
			}

			b.updateChanges = slices.Insert(b.updateChanges, upos+1, cev)

			change.to = from
			upos += 2
		}

		for i := upos; i < len(b.updateChanges); i++ {
			b.updateChanges[i].from += count
			b.updateChanges[i].to += count
		}
	}
}

func (b *ListChangeBuilder[T]) NextPermutation(from, to int, perm []int) {
	b.checkState()

	prePermFrom := from
	prePermTo := to
	prePerm := perm

	if len(b.addRemoveChanges) > 0 {
		//Because there were already some changes to the list, we need
		// to "reconstruct" the original list and create a permutation
		// as-if there were no changes to the list. We can then
		// merge this with the permutation we already did

		// This maps elements from current list to the original list.
		// -1 means the map was not in the original list.
		// Note that for performance reasons, the map is permutated when created
		// by the permutation. So it basically contains the order in which the original
		// items were permutated by our new permutation.
		mapToOriginal := make([]int, b.List.Len())
		// Marks the original-list indexes that were removed
		removed := map[int]bool{}
		last := 0
		offset := 0

		for i := 0; i < len(b.addRemoveChanges); i++ {
			change := b.addRemoveChanges[i]
			for j := last; j < change.from; j++ {
				var idx int

				if j < from || j >= to {
					idx = j
				} else {
					idx = perm[j-from]
				}

				mapToOriginal[idx] = j + offset
			}

			for j := change.from; j < change.to; j++ {
				var idx int

				if j < from || j >= to {
					idx = j
				} else {
					idx = perm[j-from]
				}

				mapToOriginal[idx] = -1
			}

			last = change.to
			removedSize := len(change.removedSlice)
			upTo := change.from + offset + removedSize

			for j := change.from + offset; j < upTo; j++ {
				removed[j] = true
			}

			offset += removedSize - (change.to - change.from)
		}

		// from the last add/remove change to the end of the list
		for i := last; i < len(mapToOriginal); i++ {
			var idx int

			if i < from || i >= to {
				idx = i
			} else {
				idx = perm[i-from]
			}

			mapToOriginal[idx] = i + offset
		}

		newPerm := make([]int, b.List.Len()+offset)
		mapPtr := 0
		for i := 0; i < len(newPerm); i++ {
			if removed[i] {
				newPerm[i] = i
			} else {
				for mapToOriginal[mapPtr] == -1 {
					mapPtr++
				}

				newPerm[mapToOriginal[mapPtr]] = i
				mapPtr++
			}
		}

		// We could theoretically find the first and last items such that
		// newPerm[i] != i and trim the permutation, but it is not necessary
		prePermFrom = 0
		prePermTo = len(newPerm)
		prePerm = newPerm
	}

	if b.permutationChange != nil {
		if prePermFrom == b.permutationChange.from && prePermTo == b.permutationChange.to {
			for i := 0; i < len(prePerm); i++ {
				b.permutationChange.perm[i] = prePerm[b.permutationChange.perm[i]-prePermFrom]
			}
		} else {
			newTo := b.permutationChange.to

			if newTo < prePermTo {
				newTo = prePermTo
			}

			newFrom := b.permutationChange.from

			if newFrom > prePermFrom {
				newFrom = prePermFrom
			}

			newPerm := make([]int, newTo-newFrom)

			for i := newFrom; i < newTo; i++ {
				if i < b.permutationChange.from || i >= b.permutationChange.to {
					newPerm[i-newFrom] = prePerm[i-prePermFrom]
				} else {
					p := b.permutationChange.perm[i-b.permutationChange.from]

					if p < prePermFrom || p >= prePermTo {
						newPerm[i-newFrom] = p
					} else {
						newPerm[i-newFrom] = prePerm[p-prePermFrom]
					}
				}
			}

			b.permutationChange.from = newFrom
			b.permutationChange.to = newTo
			b.permutationChange.perm = newPerm
		}
	} else {
		b.permutationChange = &listChangeEvent[T]{
			list: b.List,
			from: prePermFrom,
			to:   prePermTo,
			perm: prePerm,
		}
	}

	if len(b.addRemoveChanges) > 0 {
		newAdded := map[int]bool{}
		newRemoved := map[int][]T{}

		for i := 0; i < len(b.addRemoveChanges); i++ {
			change := b.addRemoveChanges[i]

			for cIndex := change.from; cIndex < change.to; cIndex++ {
				if cIndex < from || cIndex >= to {
					newAdded[cIndex] = true
				} else {
					newAdded[perm[cIndex-from]] = true
				}
			}

			if len(change.removedSlice) > 0 {
				if change.from < from || change.from >= to {
					newRemoved[change.from] = change.removedSlice
				} else {
					newRemoved[perm[change.from-from]] = change.removedSlice
				}
			}
		}

		b.addRemoveChanges = b.addRemoveChanges[0:0]

		var lastChange *listChangeEvent[T]

		for i := range newAdded {
			if lastChange == nil || lastChange.to != i {
				lastChange = &listChangeEvent[T]{
					list: b.List,
					from: i,
					to:   i + 1,
				}

				b.addRemoveChanges = append(b.addRemoveChanges, lastChange)
			} else {
				lastChange.to = i + 1
			}

			removed := newRemoved[i]
			delete(newRemoved, i)

			if len(removed) > 0 {
				lastChange.removedSlice = append(lastChange.removedSlice, removed...)
			}
		}

		for at, removed := range newRemoved {
			idx := b.findSubChange(b.addRemoveChanges, at)

			if idx >= 0 {
				panic("")
			}

			change := &listChangeEvent[T]{
				list:         b.List,
				from:         at,
				to:           at,
				removedSlice: removed,
			}

			idx = ^idx

			b.addRemoveChanges = slices.Insert(b.addRemoveChanges, idx, change)
		}
	}

	// TODO: Fixup updates
}

func (b *ListChangeBuilder[T]) NextReplace(from, to int, removed []T) {
	b.NextRemoveRange(from, removed)
	b.NextAdd(from, to)
}

func (b *ListChangeBuilder[T]) NextSet(index int, removed T) {
	b.NextRemove(index, removed)
	b.NextAdd(index, index+1)
}

func (b *ListChangeBuilder[T]) NextUpdate(index int) {
	var last *listChangeEvent[T]

	b.checkState()

	if len(b.updateChanges) > 0 {
		last = b.updateChanges[len(b.updateChanges)-1]
	}

	if last != nil && last.To() == index {
		last.to = index + 1
	} else {
		b.insertUpdate(index)
	}
}

func (b *ListChangeBuilder[T]) findSubChange(list []*listChangeEvent[T], index int) int {
	from := 0
	to := len(list) - 1

	for from <= to {
		changeIdx := (from + to) / 2
		ev := list[changeIdx]

		if index >= ev.to {
			from = changeIdx + 1
		} else if index < ev.from {
			to = changeIdx - 1
		} else {
			return changeIdx
		}
	}

	return ^from
}

func (b *ListChangeBuilder[T]) insertUpdate(index int) {
	var previous, last *listChangeEvent[T]

	idx := b.findSubChange(b.updateChanges, index)

	if idx < 0 {
		idx = ^idx

		if idx > 0 {
			previous = b.updateChanges[idx-1]
		}

		if idx < len(b.updateChanges) {
			last = b.updateChanges[idx]
		}

		if previous != nil && previous.to == index {
			previous.to = index + 1
		} else if last != nil && last.from == index+1 {
			last.from = index
		} else {
			cev := &listChangeEvent[T]{
				list:    b.List,
				from:    index,
				to:      index + 1,
				updated: true,
			}

			b.updateChanges = slices.Insert(b.updateChanges, idx, cev)
		}
	}
}

func (b *ListChangeBuilder[T]) insertRemoved(pos int, removed T) {
	var previous, last *listChangeEvent[T]

	idx := b.findSubChange(b.updateChanges, pos)

	if idx < 0 {
		idx = ^idx

		if idx > 0 {
			previous = b.addRemoveChanges[idx-1]
		}

		if idx < len(b.addRemoveChanges) {
			last = b.addRemoveChanges[idx]
		}

		if previous != nil && previous.to == pos {
			previous.removedSlice = append(previous.removedSlice, removed)
			idx--
		} else if last != nil && last.from == pos+1 {
			last.from--
			last.to--
			previous.removedSlice = append([]T{removed}, previous.removedSlice...)
		} else {
			cev := &listChangeEvent[T]{
				list:         b.List,
				from:         pos,
				to:           pos,
				removedSlice: []T{removed},
			}

			b.addRemoveChanges = slices.Insert(b.addRemoveChanges, idx, cev)
		}
	} else {
		ev := b.addRemoveChanges[idx]
		ev.to--

		if ev.from == ev.to && len(ev.removedSlice) == 0 {
			b.addRemoveChanges = slices.Delete(b.addRemoveChanges, idx, idx+1)
		}
	}

	for i := idx + 1; i < len(b.addRemoveChanges); i++ {
		ev := b.addRemoveChanges[idx]
		ev.from--
		ev.to--
	}
}

func (b *ListChangeBuilder[T]) insertAdd(from, to int) {
	var previous *listChangeEvent[T]

	idx := b.findSubChange(b.updateChanges, from)
	count := to - from

	if idx < 0 {
		idx = ^idx

		if idx > 0 {
			previous = b.addRemoveChanges[idx-1]
		}

		if previous != nil && previous.to == from {
			previous.to = to
			idx--
		} else {
			cev := &listChangeEvent[T]{
				list: b.List,
				from: from,
				to:   to,
			}

			b.addRemoveChanges = slices.Insert(b.addRemoveChanges, idx, cev)
		}
	} else {
		ev := b.addRemoveChanges[idx]
		ev.to += count
	}

	for i := idx + 1; i < len(b.addRemoveChanges); i++ {
		ev := b.addRemoveChanges[idx]
		ev.from += count
		ev.to += count
	}
}

func (b *ListChangeBuilder[T]) Begin() {
	b.changeLock++
}

func (b *ListChangeBuilder[T]) End() {
	if b.changeLock <= 0 {
		panic("called End before Begin")
	}

	b.changeLock--

	b.commit()
}

func (b *ListChangeBuilder[T]) compress(events []*listChangeEvent[T]) ([]*listChangeEvent[T], int) {
	if len(events) == 0 {
		return events, 0
	}

	removed := 0
	prev := events[0]

	for i := 1; i < len(events); i++ {
		cur := events[i]

		if cur == nil {
			continue
		}

		if prev.to == cur.from {
			prev.to = cur.from
			prev.removedSlice = append(prev.removedSlice, cur.removedSlice...)

			events[i] = nil
			removed++
		} else {
			prev = cur
		}
	}

	return events, removed
}

func (b *ListChangeBuilder[T]) checkState() {
	if b.changeLock <= 0 {
		panic("Begin was not called on this builder")
	}
}

func (b *ListChangeBuilder[T]) commit() {
	var removed int

	hasAddRemove := len(b.addRemoveChanges) > 0 && b.addRemoveChanges[0] != nil
	hasUpdates := len(b.updateChanges) > 0 && b.updateChanges[0] != nil
	hasPerm := b.permutationChange != nil

	if b.changeLock != 0 || !(hasAddRemove || hasUpdates || hasPerm) {
		return
	}

	totalSize := len(b.addRemoveChanges) + len(b.updateChanges)

	if hasPerm {
		totalSize++
	}

	events := make([]*listChangeEvent[T], 0, totalSize)

	if hasPerm {
		events = append(events, b.permutationChange)
	}

	if hasAddRemove {
		compressed, r := b.compress(b.addRemoveChanges)

		b.addRemoveChanges = compressed

		removed += r
	}

	if hasUpdates {
		compressed, r := b.compress(b.updateChanges)

		b.updateChanges = compressed

		removed += r
	}

	events = append(events, b.addRemoveChanges...)
	events = append(events, b.updateChanges...)

	events = b.finalizeSubChanges(events)

	if len(events) == 1 && events[0] != nil {
		b.FireListChanged(events[0])
	} else {
		it := &listChangeIterator[T]{
			events: events,
		}

		b.FireListChanged(it)
	}

	b.addRemoveChanges = b.addRemoveChanges[0:0]
	b.updateChanges = b.updateChanges[0:0]
	b.permutationChange = nil
}

func (b *ListChangeBuilder[T]) finalizeSubChanges(events []*listChangeEvent[T]) []*listChangeEvent[T] {
	for _, ev := range events {
		b.finalizeSubChange(ev)
	}

	return events
}

func (b *ListChangeBuilder[T]) finalizeSubChange(ev *listChangeEvent[T]) {

}
