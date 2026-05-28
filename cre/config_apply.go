package main

import "math/big"

const (
	defaultSlippageBps uint16 = 50
	defaultDeadlineSec uint64 = 300
	minSlippageBps     uint16 = 10
)

func ApplyConfig(user UserParams) FinalParams {
	out := FinalParams{
		Recipient:   user.Recipient,
		SlippageBps: user.SlippageBps,
		Deadline:    user.Deadline,
		MinReturn:   user.MinReturn,
		Pair:        user.Pair,
	}
	if out.SlippageBps < minSlippageBps {
		out.SlippageBps = defaultSlippageBps
	}
	if out.Deadline == 0 {
		out.Deadline = defaultDeadlineSec
	}
	if out.MinReturn == nil {
		out.MinReturn = big.NewInt(0)
	}
	return out
}
