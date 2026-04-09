package main

import (
	"log"
	"net"
	"time"
)

type Collector struct {
	addr   string
	parser *Parser
}

func (c *Collector) listenUdp() {

	udpAddr, err := net.ResolveUDPAddr("udp", c.addr)
	if err != nil {
		log.Fatalf("Resolving of udp addr failed: %v", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("ListenUDP failed: %v", err)
	}
	defer conn.Close()

	log.Printf("Listening on %s", c.addr)
	buf := make([]byte, 65535)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("ReadFromUDP failed: %v", err)
			continue
		}

		packet := make([]byte, n)
		copy(packet, buf[:n])

		previewLen := min(n, 32)
		log.Printf(
			"ts=%s remote=%s bytes=%d preview=% x",
			time.Now().UTC().Format(time.RFC3339),
			remoteAddr.String(),
			n,
			buf[:previewLen],
		)

		events, err := c.parser.decodePacket(packet, remoteAddr)
		if err != nil {
			log.Printf("decode packet failed from %s: %v", remoteAddr.String(), err)
			continue
		}

		for _, evt := range events.Flows {
			log.Printf("flow event: %+v", evt)
		}

		for _, evt := range events.Counters {
			log.Printf("counter event: %+v", evt)
		}
	}
}
