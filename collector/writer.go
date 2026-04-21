package main

type WriterBackend interface {
	Index(events ParsedEvents) error
}

type Writer struct {
	backend WriterBackend
}

func (i *Writer) Index(events ParsedEvents) error {
	return i.backend.Index(events)
}
