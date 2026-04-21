package main

import (
	"bytes"
	"encoding/binary"
	"net"
	"testing"

	"github.com/Cistern/sflow"
	"github.com/stretchr/testify/require"
)

var (
	testRemote = &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 58267}

	put32 = func(b *bytes.Buffer, v uint32) { _ = binary.Write(b, binary.BigEndian, v) }
	put64 = func(b *bytes.Buffer, v uint64) { _ = binary.Write(b, binary.BigEndian, v) }
)

func TestParser_parseFlowSample(t *testing.T) {
	p := Parser{}
	remote := &net.UDPAddr{
		IP:   testRemote.IP,
		Port: testRemote.Port,
	}

	sample := &sflow.FlowSample{
		SamplingRate: 10,
		Records: []sflow.Record{
			sflow.RawPacketFlow{
				FrameLength: 1518,
			},
			sflow.RawPacketFlow{
				FrameLength: 78,
			},
		},
	}

	events := p.parseFlowSample(sample, remote)

	require.Len(t, events, 2)
	require.Equal(t, EventKindFlow, events[0].Kind)
	require.Equal(t, "127.0.0.1", events[0].AgentIP)
	require.Equal(t, uint32(1518), events[0].FrameLength)
	require.Equal(t, uint32(10), events[0].SampleRate)
	require.Equal(t, uint32(78), events[1].FrameLength)
}

func TestParser_parseCounterSample(t *testing.T) {
	p := Parser{}
	remote := &net.UDPAddr{
		IP:   testRemote.IP,
		Port: testRemote.Port,
	}

	sample := &sflow.CounterSample{

		Records: []sflow.Record{
			sflow.GenericInterfaceCounters{
				Index:             8,
				InOctets:          1000,
				OutOctets:         2000,
				InUnicastPackets:  10,
				OutUnicastPackets: 20,
			},
		},
	}

	events := p.parseCounterSample(sample, remote)

	require.Equal(t, EventKindCounter, events[0].Kind)
	require.Equal(t, "127.0.0.1", events[0].AgentIP)
	require.Equal(t, uint32(8), events[0].IfIndex)
	require.Equal(t, uint64(1000), events[0].InOctets)
	require.Equal(t, uint64(2000), events[0].OutOctets)
	require.Equal(t, uint32(10), events[0].InUcastPkts)
	require.Equal(t, uint32(20), events[0].OutUcastPkts)
}

func TestParser_DecodeFlowSample(t *testing.T) {
	p := Parser{}
	packet := sflowDatagram(flowSample)

	events, err := p.DecodePacket(packet, testRemote)
	require.NoError(t, err)
	require.Len(t, events.Flows, 1)
	require.Empty(t, events.Counters)

	f := events.Flows[0]
	require.Equal(t, uint32(128), f.FrameLength)
	require.Equal(t, uint32(512), f.SampleRate)
	require.Equal(t, "127.0.0.1", f.AgentIP)
	require.Equal(t, EventKindFlow, f.Kind)
}

func TestParser_DecodeCounterSample(t *testing.T) {
	p := Parser{}
	packet := sflowDatagram(counterSample)

	events, err := p.DecodePacket(packet, testRemote)
	require.NoError(t, err)
	require.Len(t, events.Counters, 1)
	require.Empty(t, events.Flows)

	c := events.Counters[0]
	require.Equal(t, uint32(1), c.IfIndex)
	require.Equal(t, uint64(100_000), c.InOctets)
	require.Equal(t, uint64(200_000), c.OutOctets)
	require.Equal(t, uint32(500), c.InUcastPkts)
	require.Equal(t, uint32(800), c.OutUcastPkts)
	require.Equal(t, "127.0.0.1", c.AgentIP)
	require.Equal(t, EventKindCounter, c.Kind)
}

func sflowDatagram(samples ...func() (sampleType uint32, body []byte)) []byte {
	pkt := &bytes.Buffer{}
	put32(pkt, 5)                    // sFlow version
	put32(pkt, 1)                    // address type: IPv4
	pkt.Write([]byte{127, 0, 0, 1})  // agent IP
	put32(pkt, 0)                    // sub-agent ID
	put32(pkt, 1)                    // sequence number
	put32(pkt, 1000)                 // uptime (ms)
	put32(pkt, uint32(len(samples))) // number of samples

	for _, s := range samples {
		typ, body := s()
		put32(pkt, typ)
		put32(pkt, uint32(len(body)))
		pkt.Write(body)
	}

	return pkt.Bytes()
}

func flowSample() (uint32, []byte) {
	header := []byte{0x00, 0x50, 0x56, 0xAB} // 4 bytes – already aligned

	rawRecord := &bytes.Buffer{}
	put32(rawRecord, 1)                   // protocol: Ethernet
	put32(rawRecord, 128)                 // frame length
	put32(rawRecord, 0)                   // stripped bytes
	put32(rawRecord, uint32(len(header))) // header length
	rawRecord.Write(header)

	body := &bytes.Buffer{}
	put32(body, 1)          // sequence number
	put32(body, 0x00000001) // source ID
	put32(body, 512)        // sampling rate
	put32(body, 1024)       // sample pool
	put32(body, 0)          // drops
	put32(body, 1)          // input interface
	put32(body, 2)          // output interface
	put32(body, 1)          // number of records

	put32(body, 1) // record type: RawPacketFlow
	put32(body, uint32(rawRecord.Len()))
	body.Write(rawRecord.Bytes())

	return 1, body.Bytes() // sample type 1 = FlowSample
}

func counterSample() (uint32, []byte) {
	record := &bytes.Buffer{}
	put32(record, 1)             // IfIndex
	put32(record, 6)             // type: ethernetCsmacd
	put64(record, 1_000_000_000) // speed
	put32(record, 1)             // direction: fullDuplex
	put32(record, 3)             // status
	put64(record, 100_000)       // InOctets
	put32(record, 500)           // InUnicastPackets
	put32(record, 10)            // InMulticastPackets
	put32(record, 5)             // InBroadcastPackets
	put32(record, 0)             // InDiscards
	put32(record, 0)             // InErrors
	put32(record, 0)             // InUnknownProtos
	put64(record, 200_000)       // OutOctets
	put32(record, 800)           // OutUnicastPackets
	put32(record, 20)            // OutMulticastPackets
	put32(record, 8)             // OutBroadcastPackets
	put32(record, 0)             // OutDiscards
	put32(record, 0)             // OutErrors
	put32(record, 0)             // PromiscuousMode

	body := &bytes.Buffer{}
	put32(body, 1)          // sequence number
	put32(body, 0x00000001) // source ID
	put32(body, 1)          // number of records

	put32(body, 1) // record type: GenericInterfaceCounters
	put32(body, uint32(record.Len()))
	body.Write(record.Bytes())

	return 2, body.Bytes() // sample type 2 = CounterSample
}
