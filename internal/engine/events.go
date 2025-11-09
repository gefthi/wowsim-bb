package engine

import "time"

type scheduledEvent struct {
	executeAt time.Duration
	action    func()
	cancelled bool
}

func (e *scheduledEvent) Cancel() {
	if e == nil {
		return
	}
	e.cancelled = true
	e.action = nil
}

type eventQueue []*scheduledEvent

func (eq *eventQueue) add(ev *scheduledEvent) {
	if ev == nil {
		return
	}
	inserted := false
	for i, existing := range *eq {
		if ev.executeAt < existing.executeAt {
			*eq = append(*eq, nil)
			copy((*eq)[i+1:], (*eq)[i:])
			(*eq)[i] = ev
			inserted = true
			break
		}
	}
	if !inserted {
		*eq = append(*eq, ev)
	}
}

func (eq *eventQueue) cleanFront() {
	for len(*eq) > 0 {
		ev := (*eq)[0]
		if ev == nil || ev.cancelled {
			*eq = (*eq)[1:]
			continue
		}
		break
	}
}

func (eq *eventQueue) popReady(now time.Duration) *scheduledEvent {
	eq.cleanFront()
	if len(*eq) == 0 {
		return nil
	}
	ev := (*eq)[0]
	if ev.executeAt > now {
		return nil
	}
	*eq = (*eq)[1:]
	if ev.cancelled {
		return nil
	}
	return ev
}

func (eq *eventQueue) nextDelta(now time.Duration) (time.Duration, bool) {
	eq.cleanFront()
	if len(*eq) == 0 {
		return 0, false
	}
	ev := (*eq)[0]
	if ev.executeAt <= now {
		return 0, true
	}
	return ev.executeAt - now, true
}
