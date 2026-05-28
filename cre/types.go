package main

import "math/big"

type PriceUpdate struct {
	Source    string
	Pair      string
	Price     *big.Int
	Timestamp uint64
}

type Report struct {
	WorkflowID [32]byte
	Timestamp  uint64
	Recipient  [20]byte
	Amount     *big.Int
	Nonce      [32]byte
}

type Event struct {
	ID          string
	BlockNumber uint64
	BlockHash   [32]byte
	TxHash      [32]byte
	LogIndex    uint32
	Payload     []byte
}

type UserParams struct {
	Recipient   string
	SlippageBps uint16
	Deadline    uint64
	MinReturn   *big.Int
	Pair        string
}

type FinalParams struct {
	Recipient   string
	SlippageBps uint16
	Deadline    uint64
	MinReturn   *big.Int
	Pair        string
}
