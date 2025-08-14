package common

// enum event for EvictedFunc
const (
	ExpirationEvent = iota
	EvictionEvent
	DeleteEvent
	ClearEvent
)

type EvictedFunc = func(key, value any, event int)
