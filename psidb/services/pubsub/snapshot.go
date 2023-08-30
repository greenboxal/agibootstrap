package pubsub

import (
	"time"

	"github.com/samber/lo"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/db/graphfs"
)

type SchedulerSnapshot struct {
	Queues   map[string][]*PendingItemSnapshot             `json:"queues,omitempty"`
	Promises map[psi.PromiseHandle]*PendingPromiseSnapshot `json:"promises,omitempty"`
}

type PendingItemSnapshot struct {
	Key       scheduledItemKey `json:"key"`
	Start     time.Time        `json:"start"`
	Deadline  time.Time        `json:"deadline"`
	Scheduled bool             `json:"scheduled,omitempty"`

	Dependencies []psi.PromiseHandle `json:"dependencies,omitempty"`

	Notification *graphfs.JournalEntry `json:"notification,omitempty"`
	Confirmation *graphfs.JournalEntry `json:"confirmation,omitempty"`
}

type PendingPromiseSnapshot struct {
	Key        psi.PromiseHandle `json:"key"`
	Resolved   bool              `json:"resolved"`
	Count      int               `json:"count"`
	Dependees  []string          `json:"dependees,omitempty"`
	References int               `json:"references"`
}

func (sch *Scheduler) DumpStatistics() *SchedulerSnapshot {
	result := &SchedulerSnapshot{
		Queues:   map[string][]*PendingItemSnapshot{},
		Promises: map[psi.PromiseHandle]*PendingPromiseSnapshot{},
	}

	func() {
		sch.pendingMutex.RLock()
		defer sch.pendingMutex.RUnlock()

		for _, item := range sch.pending {
			result.Queues[item.qkey] = append(result.Queues[item.qkey], &PendingItemSnapshot{
				Key:          item.skey,
				Start:        item.start,
				Deadline:     item.deadline,
				Scheduled:    item.scheduled,
				Notification: item.entry,
				Confirmation: item.confirmation,

				Dependencies: lo.Map(item.deps, func(dep *pendingPromise, _ int) psi.PromiseHandle {
					return dep.key
				}),
			})
		}
	}()

	func() {
		sch.promisesMutex.RLock()
		defer sch.promisesMutex.RUnlock()

		for _, promise := range sch.promises {
			result.Promises[promise.key] = &PendingPromiseSnapshot{
				Key:        promise.key,
				Resolved:   promise.resolved,
				Count:      promise.count,
				Dependees:  promise.queues,
				References: promise.refs,
			}
		}
	}()

	return result
}
