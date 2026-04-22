package main

import (
	"bytes"
	"net"
	"time"

	"github.com/Cistern/sflow"
)

type PacketParser interface {
	DecodePacket(packet []byte, remote *net.UDPAddr) (ParsedEvents, error)
}

type Parser struct{}

func (p *Parser) DecodePacket(packet []byte, remote *net.UDPAddr) (ParsedEvents, error) {
	reader := bytes.NewReader(packet)
	decoder := sflow.NewDecoder(reader)

	datagram, err := decoder.Decode()
	if err != nil {
		return ParsedEvents{}, err
	}

	var events ParsedEvents

	for _, sample := range datagram.Samples {

		if flowSample, ok := sample.(*sflow.FlowSample); ok {
			events.Flows = append(events.Flows, p.parseFlowSample(flowSample, remote)...)
		}

		if counterSample, ok := sample.(*sflow.CounterSample); ok {
			events.Counters = append(events.Counters, p.parseCounterSample(counterSample, remote)...)
		}
	}

	return events, nil
}

func (p *Parser) parseFlowSample(flowSample *sflow.FlowSample, remote *net.UDPAddr) []FlowEvent {
	var events []FlowEvent
	for _, record := range flowSample.Records {
		rawPacket, ok := record.(sflow.RawPacketFlow)
		if !ok {
			continue
		}

		events = append(events, FlowEvent{
			Event:       baseEvent(remote, EventKindFlow),
			FrameLength: rawPacket.FrameLength,
			SampleRate:  flowSample.SamplingRate,
		})
	}

	return events
}

func (p *Parser) parseCounterSample(counterSample *sflow.CounterSample, remote *net.UDPAddr) []CounterEvent {
	var events []CounterEvent
	for _, record := range counterSample.Records {
		genericCounters, ok := record.(sflow.GenericInterfaceCounters)
		if !ok {
			continue
		}

		events = append(events, CounterEvent{
			Event:        baseEvent(remote, EventKindCounter),
			IfIndex:      genericCounters.Index,
			InOctets:     genericCounters.InOctets,
			OutOctets:    genericCounters.OutOctets,
			InUcastPkts:  genericCounters.InUnicastPackets,
			OutUcastPkts: genericCounters.OutUnicastPackets,
		})
	}

	return events
}

func baseEvent(remote *net.UDPAddr, kind string) Event {
	return Event{
		Timestamp: time.Now().UTC(),
		Kind:      kind,
		AgentIP:   remote.IP.String(),
		Collector: "collector-mac",
	}
}
