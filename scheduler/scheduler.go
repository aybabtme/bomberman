package scheduler

import (
	"container/heap"
)

// Scheduler registers actions that will occur in the future.
type Scheduler struct {
	events  *eventHeap
	now     int
	current []*event
}

// NewScheduler creates a Scheduler starting at turn 0.
func NewScheduler() *Scheduler {
	sch := &Scheduler{
		events:  &eventHeap{},
		now:     0,
		current: make([]*event, 0),
	}
	heap.Init(sch.events)
	return sch
}

// Register will add an action that starts at some turn in the future. If
// the action is registered in the past, it will be dropped and ignored.
func (s *Scheduler) Register(act Action, startsIn int) {
	e := &event{
		Stamp:     s.now + startsIn,
		TurnsDone: 0,
		Action:    act,
	}
	heap.Push(s.events, e)
}

// HasNext is true as long as there are events registered to happen
func (s *Scheduler) HasNext() bool {
	return !s.events.Empty()
}

// NextTurn advances to the next turn. All the actions of the previous turn
// are removed.
func (s *Scheduler) NextTurn() {
	s.now++
	if len(s.current) != 0 {
		for i := range s.current {
			s.current[i] = nil
		}
		s.current = s.current[:0]
	}

	if s.events.Empty() {
		return
	}

	for !s.events.Empty() && s.events.Peek().Stamp <= s.now {
		e := heap.Pop(s.events).(*event)
		if e.Stamp < s.now {
			// Ignore it
			continue
		}
		s.current = append(s.current, e)
	}
}

// DoTurn will invoke the given lambda with all the actions occuring at this
// turn, along with the delta since the action has begun.  If this is the first
// turn of an action, delta will be 0.
func (s *Scheduler) DoTurn(eachAction func(a Action, delta int) error) error {
	var err error
	for i := range s.current {
		ev := (s.current)[i]
		err = eachAction(ev.Action, ev.TurnsDone)
		if err != nil {
			break
		}

		ev.TurnsDone++
		if ev.TurnsDone < ev.Action.Duration() {
			ev.Stamp = s.now + 1
			heap.Push(s.events, ev)
		}
	}
	return err
}

// Action takes place at a time for a duration
type Action interface {
	Duration() int
}

type event struct {
	Stamp     int    `json:"stamp"`
	TurnsDone int    `json:"turnsDone"`
	Action    Action `json:"action"`
}

type eventHeap []*event

func (ev eventHeap) Len() int           { return len(ev) }
func (ev eventHeap) Less(i, j int) bool { return ev[i].Stamp < ev[j].Stamp }
func (ev eventHeap) Swap(i, j int)      { ev[i], ev[j] = ev[j], ev[i] }

func (ev *eventHeap) Pop() interface{} {
	old := *ev
	n := len(old)
	item := old[n-1]
	*ev = old[0 : n-1]
	return item
}

func (ev *eventHeap) Push(x interface{}) {
	*ev = append(*ev, x.(*event))
}

func (ev *eventHeap) Peek() *event {
	if ev.Empty() {
		return nil
	}
	return (*ev)[0]
}

func (ev *eventHeap) Empty() bool {
	return len(*ev) == 0
}
