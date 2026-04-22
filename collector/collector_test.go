package main

import (
	"collector/ingest"
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockParser struct {
	mu     sync.Mutex
	calls  [][]byte
	result ingest.ParsedEvents
	err    error
}

func (m *mockParser) DecodePacket(packet []byte, _ *net.UDPAddr) (ingest.ParsedEvents, error) {
	m.mu.Lock()
	m.calls = append(m.calls, packet)
	m.mu.Unlock()
	return m.result, m.err
}

func (m *mockParser) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

type mockWriter struct {
	mu     sync.Mutex
	events []ingest.ParsedEvents
	err    error
}

func (m *mockWriter) Index(events ingest.ParsedEvents) error {
	m.mu.Lock()
	m.events = append(m.events, events)
	m.mu.Unlock()
	return m.err
}

func (m *mockWriter) indexed() []ingest.ParsedEvents {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]ingest.ParsedEvents(nil), m.events...)
}

func newTestCollector(t *testing.T, parser PacketParser, writer ingest.EventWriter) (*Collector, *net.UDPConn, context.CancelFunc) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())

	c := NewCollector("127.0.0.1:0", parser, writer, 64, 2)
	require.NoError(t, c.Start(ctx))

	// Dial back to the port the collector actually bound.
	conn, err := net.DialUDP("udp", nil, c.conn.LocalAddr().(*net.UDPAddr))
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	return c, conn, cancel
}

func sendAndWait(t *testing.T, conn *net.UDPConn, payload []byte, condition func() bool) {
	t.Helper()
	_, err := conn.Write(payload)
	require.NoError(t, err)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out waiting for condition")
}

func TestCollector_FlowEventsAreIndexed(t *testing.T) {
	parser := &mockParser{
		result: ingest.ParsedEvents{
			Flows: []ingest.FlowEvent{{FrameLength: 128, SampleRate: 512}},
		},
	}
	writer := &mockWriter{}

	c, conn, cancel := newTestCollector(t, parser, writer)

	sendAndWait(t, conn, []byte("dummy"), func() bool {
		return len(writer.indexed()) == 1
	})

	cancel()
	c.Wait()

	got := writer.indexed()
	require.Len(t, got, 1)
	require.Len(t, got[0].Flows, 1)
	require.Empty(t, got[0].Counters)
	require.Equal(t, uint32(128), got[0].Flows[0].FrameLength)
}

func TestCollector_CounterEventsAreIndexed(t *testing.T) {
	parser := &mockParser{
		result: ingest.ParsedEvents{
			Counters: []ingest.CounterEvent{{IfIndex: 1, InOctets: 100_000, OutOctets: 200_000}},
		},
	}
	writer := &mockWriter{}

	c, conn, cancel := newTestCollector(t, parser, writer)

	sendAndWait(t, conn, []byte("dummy"), func() bool {
		return len(writer.indexed()) == 1
	})

	cancel()
	c.Wait()

	got := writer.indexed()
	require.Len(t, got, 1)
	require.Len(t, got[0].Counters, 1)
	require.Empty(t, got[0].Flows)
	require.Equal(t, uint32(1), got[0].Counters[0].IfIndex)
}

func TestCollector_EmptyEventsAreDropped(t *testing.T) {
	parser := &mockParser{result: ingest.ParsedEvents{}}
	writer := &mockWriter{}

	c, conn, cancel := newTestCollector(t, parser, writer)

	_, err := conn.Write([]byte("dummy"))
	require.NoError(t, err)

	// Give the reader time to process.
	time.Sleep(100 * time.Millisecond)

	cancel()
	c.Wait()

	require.Empty(t, writer.indexed())
	require.Equal(t, 1, parser.callCount())
}

func TestCollector_ParseErrorIsSkipped(t *testing.T) {
	parser := &mockParser{err: errors.New("bad packet")}
	writer := &mockWriter{}

	c, conn, cancel := newTestCollector(t, parser, writer)

	_, err := conn.Write([]byte("garbage"))
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	cancel()
	c.Wait()

	require.Empty(t, writer.indexed())
	require.Equal(t, 1, parser.callCount())
}

func TestCollector_WriterErrorDoesNotStopWorker(t *testing.T) {
	// Writer always fails – worker should log and keep draining, not crash.
	parser := &mockParser{
		result: ingest.ParsedEvents{Flows: []ingest.FlowEvent{{FrameLength: 64}}},
	}
	writer := &mockWriter{err: errors.New("index unavailable")}

	c, conn, cancel := newTestCollector(t, parser, writer)

	for i := 0; i < 3; i++ {
		_, err := conn.Write([]byte("dummy"))
		require.NoError(t, err)
	}

	// All three should have been attempted despite errors.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if len(writer.indexed()) == 3 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	cancel()
	c.Wait()

	require.Len(t, writer.indexed(), 3)
}

func TestCollector_GracefulShutdown(t *testing.T) {
	parser := &mockParser{result: ingest.ParsedEvents{}}
	writer := &mockWriter{}

	c, _, cancel := newTestCollector(t, parser, writer)

	cancel()

	done := make(chan struct{})
	go func() {
		c.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("collector did not shut down in time")
	}
}
