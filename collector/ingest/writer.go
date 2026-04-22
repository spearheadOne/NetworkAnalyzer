package ingest

type EventWriter interface {
	Index(events ParsedEvents) error
}
