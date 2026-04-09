package main

import "time"

const (
	EventKindFlow    = "flow"
	EventKindCounter = "counter"
)

type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Kind      string    `json:"kind"`
	AgentIP   string    `json:"agent_ip"`
	Collector string    `json:"collector"`
}

type FlowEvent struct {
	Event
	FrameLength uint32 `json:"frame_length"`
	SampleRate  uint32 `json:"sample_rate"`
}

type CounterEvent struct {
	Event
	Index        uint32 `json:"index,omitempty"`
	InOctets     uint64 `json:"in_octets,omitempty"`
	OutOctets    uint64 `json:"out_octets,omitempty"`
	InUcastPkts  uint32 `json:"in_ucast_pkts,omitempty"`
	OutUcastPkts uint32 `json:"out_ucast_pkts,omitempty"`
}

type ParsedEvents struct {
	Flows    []FlowEvent
	Counters []CounterEvent
}
