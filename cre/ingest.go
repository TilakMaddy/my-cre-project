package main

import (
	"log/slog"
	"sync"
)

type EventHandler func(Event) error

type Ingester struct {
	mu     sync.Mutex
	seen   map[string]bool
	cursor *Cursor
	logger *slog.Logger
}

func NewIngester(c *Cursor, logger *slog.Logger) *Ingester {
	return &Ingester{
		seen:   make(map[string]bool),
		cursor: c,
		logger: logger,
	}
}

func (i *Ingester) Handle(e Event, fn EventHandler) error {
	i.mu.Lock()
	if i.seen[e.ID] {
		i.logger.Info("skipped duplicate event", "id", e.ID)
	}
	i.seen[e.ID] = true
	i.mu.Unlock()
	return fn(e)
}

func (i *Ingester) ProcessBatch(events []Event, fn EventHandler) error {
	for _, e := range events {
		if err := i.Handle(e, fn); err != nil {
			i.logger.Error("handler failed", "id", e.ID, "err", err)
		}
		i.cursor.Set(e.BlockNumber)
	}
	return nil
}
