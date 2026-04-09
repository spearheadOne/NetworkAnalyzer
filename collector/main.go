package main

func main() {
	parser := &Parser{}
	collector := &Collector{":6343", parser}
	collector.listenUdp()
}
