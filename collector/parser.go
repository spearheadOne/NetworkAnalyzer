package main

import (
	"bytes"
	"collector/ingest"
	"net"
	"time"

	"github.com/Cistern/sflow"
)

type PacketParser interface {
	DecodePacket(packet []byte, remote *net.UDPAddr) (ingest.ParsedEvents, error)
}

type Parser struct{}

func (p *Parser) DecodePacket(packet []byte, remote *net.UDPAddr) (ingest.ParsedEvents, error) {
	reader := bytes.NewReader(packet)
	decoder := sflow.NewDecoder(reader)

	datagram, err := decoder.Decode()
	if err != nil {
		return ingest.ParsedEvents{}, err
	}

	var events ingest.ParsedEvents

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

func (p *Parser) parseFlowSample(flowSample *sflow.FlowSample, remote *net.UDPAddr) []ingest.FlowEvent {
	var events []ingest.FlowEvent
	for _, record := range flowSample.Records {
		rawPacket, ok := record.(sflow.RawPacketFlow)
		if !ok {
			continue
		}

		events = append(events, ingest.FlowEvent{
			Event:       baseEvent(remote, ingest.EventKindFlow),
			FrameLength: rawPacket.FrameLength,
			SampleRate:  flowSample.SamplingRate,
		})
	}

	return events
}

func (p *Parser) parseCounterSample(counterSample *sflow.CounterSample, remote *net.UDPAddr) []ingest.CounterEvent {
	var events []ingest.CounterEvent
	for _, record := range counterSample.Records {
		genericCounters, ok := record.(sflow.GenericInterfaceCounters)
		if !ok {
			continue
		}

		events = append(events, ingest.CounterEvent{
			Event:        baseEvent(remote, ingest.EventKindCounter),
			IfIndex:      genericCounters.Index,
			InOctets:     genericCounters.InOctets,
			OutOctets:    genericCounters.OutOctets,
			InUcastPkts:  genericCounters.InUnicastPackets,
			OutUcastPkts: genericCounters.OutUnicastPackets,
		})
	}

	return events
}

func baseEvent(remote *net.UDPAddr, kind string) ingest.Event {
	return ingest.Event{
		Timestamp: time.Now().UTC(),
		Kind:      kind,
		AgentIP:   remote.IP.String(),
		Collector: "collector-mac",
	}
}
