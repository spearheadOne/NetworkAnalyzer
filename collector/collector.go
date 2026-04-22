package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type Collector struct {
	addr       string
	parser     PacketParser
	writer     EventWriter
	eventsCh   chan ParsedEvents
	workersNum int

	conn *net.UDPConn
	wg   sync.WaitGroup
}

func NewCollector(addr string, parser PacketParser, writer EventWriter, queueSize int, workersNum int) *Collector {

	return &Collector{
		addr:       addr,
		parser:     parser,
		writer:     writer,
		eventsCh:   make(chan ParsedEvents, queueSize),
		workersNum: workersNum,
	}

}

func (c *Collector) Start(ctx context.Context) error {
	udpAddr, err := net.ResolveUDPAddr("udp", c.addr)
	if err != nil {
		return fmt.Errorf("resolve udp addr failed: %w", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("listen udp failed: %w", err)
	}
	c.conn = conn

	log.Printf("Listening on %s", c.addr)
	log.Printf("collector workers=%d queue=%d", c.workersNum, cap(c.eventsCh))

	c.wg.Add(c.workersNum)
	for i := 0; i < c.workersNum; i++ {
		go c.runIndexWorker(i + 1)
	}

	c.wg.Add(1)
	go c.runReader(ctx)

	go func() {
		<-ctx.Done()
		_ = c.conn.Close()
	}()

	return nil
}

func (c *Collector) Wait() {
	c.wg.Wait()
}

func (c *Collector) runReader(ctx context.Context) {
	defer c.wg.Done()
	defer close(c.eventsCh)

	buf := make([]byte, 65535)

	for {
		n, remoteAddr, err := c.conn.ReadFrom(buf)
		if err != nil {
			select {
			case <-ctx.Done():
				log.Println("shutdown signal received, closing udp socket")
				return
			default:
				log.Printf("ReadFromUDP failed: %v", err)
				continue
			}
		}

		udpRemote, ok := remoteAddr.(*net.UDPAddr)
		if !ok {
			log.Printf("unexpected remote addr type: %T", udpRemote)
			continue
		}

		packet := make([]byte, n)
		copy(packet, buf[:n])

		previewLen := min(len(packet), 32)
		log.Printf(
			"ts=%s remote=%s bytes=%d preview=% x",
			time.Now().UTC().Format(time.RFC3339),
			remoteAddr.String(),
			len(packet),
			packet[:previewLen],
		)

		events, err := c.parser.DecodePacket(packet, udpRemote)
		if err != nil {
			log.Printf("decode packet failed from %s: %v", udpRemote.String(), err)
			continue
		}

		if len(events.Flows) == 0 && len(events.Counters) == 0 {
			continue
		}

		select {
		case c.eventsCh <- events:
		case <-ctx.Done():
			return
		default:
			log.Printf("events queue is full, dropping parsed events from %s", udpRemote.String())
		}
	}

}

func (c *Collector) runIndexWorker(id int) {
	defer c.wg.Done()

	for events := range c.eventsCh {
		if err := c.writer.Index(events); err != nil {
			log.Printf("index worker=%d failed: %v", id, err)
		}
	}
}
