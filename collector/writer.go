package main

type EventWriter interface {
	Index(events ParsedEvents) error
}

type Writer struct {
	backend EventWriter
}

func (i *Writer) Index(events ParsedEvents) error {
	return i.backend.Index(events)
}
