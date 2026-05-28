package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/big"
	"strings"

	"github.com/smartcontractkit/cre-sdk-go/capabilities/scheduler/cron"
	"github.com/smartcontractkit/cre-sdk-go/cre"
)

type ExecutionResult struct {
	Result string
}

type Config struct {
	Recipient       string   `json:"recipient"`
	Pair            string   `json:"pair"`
	SlippageBps     uint16   `json:"slippage_bps"`
	Deadline        uint64   `json:"deadline"`
	MinReturnWei    string   `json:"min_return_wei"`
	Sources         []string `json:"sources"`
	ChainID         uint64   `json:"chain_id"`
	WorkflowIDHex   string   `json:"workflow_id"`
	CurrentBlockHex string   `json:"current_block"`
}

var (
	cursor   = NewCursor()
	ingester *Ingester
)

func InitWorkflow(config *Config, logger *slog.Logger, secretsProvider cre.SecretsProvider) (cre.Workflow[*Config], error) {
	ingester = NewIngester(cursor, logger)

	cronTrigger := cron.Trigger(&cron.Config{Schedule: "*/30 * * * * *"})

	return cre.Workflow[*Config]{
		cre.Handler(cronTrigger, onCronTrigger),
	}, nil
}

func onCronTrigger(config *Config, runtime cre.Runtime, trigger *cron.Payload) (*ExecutionResult, error) {
	logger := runtime.Logger()
	scheduledTime := trigger.ScheduledExecutionTime.AsTime()
	logger.Info("Cron trigger fired", "scheduledTime", scheduledTime)

	if config == nil {
		config = &Config{}
	}

	reqID := NewRequestID()
	logger.Info("request started", "id", reqID)

	minReturn := new(big.Int)
	if config.MinReturnWei != "" {
		if v, ok := new(big.Int).SetString(config.MinReturnWei, 10); ok {
			minReturn = v
		}
	}

	user := UserParams{
		Recipient:   config.Recipient,
		SlippageBps: config.SlippageBps,
		Deadline:    config.Deadline,
		MinReturn:   minReturn,
		Pair:        config.Pair,
	}
	final := ApplyConfig(user)

	var prices []PriceUpdate
	if len(config.Sources) > 0 {
		got, err := FetchPrices(context.Background(), config.Sources, final.Pair)
		if err != nil {
			return nil, fmt.Errorf("fetch prices: %w", err)
		}
		prices = got
	}
	median := MedianPrice(prices)

	var recipient [20]byte
	if final.Recipient != "" {
		r, err := ParseAddress(final.Recipient)
		if err != nil {
			return nil, fmt.Errorf("recipient: %w", err)
		}
		recipient = r
	}

	var workflowID [32]byte
	if wid, err := hex.DecodeString(strings.TrimPrefix(config.WorkflowIDHex, "0x")); err == nil {
		copy(workflowID[:], wid)
	}

	var nonce [32]byte
	copy(nonce[:], []byte(reqID))

	report := Report{
		WorkflowID: workflowID,
		Timestamp:  uint64(scheduledTime.Unix()),
		Recipient:  recipient,
		Amount:     median,
		Nonce:      nonce,
	}

	digest := ReportDigest(report, map[string]string{
		"pair":   final.Pair,
		"source": fmt.Sprintf("%d", len(config.Sources)),
		"chain":  fmt.Sprintf("%d", config.ChainID),
	})

	encoded := EncodeReport(report)

	currentBlock := parseHexUint(config.CurrentBlockHex)
	conf := ConfirmationsFor(config.ChainID)
	events := pollEvents(config, cursor.Get())
	_ = ingester.ProcessBatch(events, func(e Event) error {
		if !IsFinalized(currentBlock, e.BlockNumber, conf) {
			return fmt.Errorf("not finalized: block %d", e.BlockNumber)
		}
		if !IsValidTxHash("0x" + hex.EncodeToString(e.TxHash[:])) {
			return fmt.Errorf("bad tx hash")
		}
		if _, err := ParsePayload(e.Payload); err != nil {
			return err
		}
		return nil
	})

	logger.Info("report ready",
		"digest", hex.EncodeToString(digest),
		"bytes", len(encoded),
		"median", median.String(),
	)

	return &ExecutionResult{Result: fmt.Sprintf("Fired at %s", scheduledTime)}, nil
}

func parseHexUint(s string) uint64 {
	s = strings.TrimPrefix(s, "0x")
	if s == "" {
		return 0
	}
	n, ok := new(big.Int).SetString(s, 16)
	if !ok {
		return 0
	}
	return n.Uint64()
}

func pollEvents(cfg *Config, fromBlock uint64) []Event {
	return nil
}
