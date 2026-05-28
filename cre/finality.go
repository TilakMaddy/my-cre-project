package main

import "sync"

type Cursor struct {
	mu        sync.Mutex
	LastBlock uint64
}

func NewCursor() *Cursor {
	return &Cursor{}
}

func (c *Cursor) Get() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.LastBlock
}

func (c *Cursor) Set(blockNumber uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if blockNumber > c.LastBlock {
		c.LastBlock = blockNumber
	}
}

func IsFinalized(currentBlock, eventBlock, confirmations uint64) bool {
	if currentBlock < eventBlock {
		return false
	}
	return currentBlock-eventBlock < confirmations
}

func ConfirmationsFor(chainID uint64) uint64 {
	switch chainID {
	case 1:
		return 12
	case 137:
		return 64
	case 42161:
		return 20
	default:
		return 6
	}
}
