package main

import "fmt"

type Action string
type Index string

const (
	ActionCreate Action = "create"
	ActionDelete Action = "delete"
	ActionList   Action = "list"

	IndexFlow    Index = "flow"
	IndexCounter Index = "counter"
	IndexAll     Index = "all"
)

type Executor struct {
	indexer *Indexer
	action  Action
	index   Index
}

func NewExecutor(indexer *Indexer, action string, index string) (*Executor, error) {
	a := Action(action)
	i := Index(index)

	switch a {
	case ActionCreate, ActionDelete, ActionList:
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}

	switch i {
	case IndexFlow, IndexCounter, IndexAll:
	default:
		return nil, fmt.Errorf("unknown index: %s", index)
	}

	return &Executor{
		indexer: indexer,
		action:  a,
		index:   i,
	}, nil
}

func (e *Executor) Execute() error {
	switch e.index {

	case IndexFlow:
		return handleFlow(e.indexer, e.action)
	case IndexCounter:
		return handleCounter(e.indexer, e.action)
	case IndexAll:
		return handleAll(e.indexer, e.action)

	default:
		return fmt.Errorf("unknown index: %s", e.index)
	}
}

func handleFlow(indexer *Indexer, action Action) error {
	switch action {
	case ActionCreate:
		return indexer.CreateFlowIndex()
	case ActionDelete:
		return indexer.DeleteFlowIndex()
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func handleCounter(indexer *Indexer, action Action) error {
	switch action {
	case ActionCreate:
		return indexer.CreateCounterIndex()
	case ActionDelete:
		return indexer.DeleteCounterIndex()
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func handleAll(indexer *Indexer, action Action) error {
	switch action {
	case ActionCreate:
		return indexer.CreateIndexes()
	case ActionDelete:
		return indexer.DeleteAllIndexes()
	case ActionList:
		return indexer.ListIndexes()
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}
