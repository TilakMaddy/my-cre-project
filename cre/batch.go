package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"golang.org/x/sync/errgroup"
)

func FetchPrices(ctx context.Context, sources []string, pair string) ([]PriceUpdate, error) {
	g, ctx := errgroup.WithContext(ctx)
	results := make([]PriceUpdate, len(sources))
	for i, src := range sources {
		i, src := i, src
		g.Go(func() error {
			p, err := fetchPrice(ctx, src, pair)
			if err != nil {
				return fmt.Errorf("source %s: %w", src, err)
			}
			results[i] = p
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

func fetchPrice(ctx context.Context, source, pair string) (PriceUpdate, error) {
	jitter := time.Duration(rand.Intn(50)) * time.Millisecond
	select {
	case <-ctx.Done():
		return PriceUpdate{}, ctx.Err()
	case <-time.After(jitter):
	}
	raw, err := RPCCall(source, "oracle_getPrice", []any{pair})
	if err != nil {
		return PriceUpdate{}, err
	}
	var priceStr string
	if err := json.Unmarshal(raw, &priceStr); err != nil {
		return PriceUpdate{}, err
	}
	price, ok := new(big.Int).SetString(priceStr, 10)
	if !ok {
		return PriceUpdate{}, fmt.Errorf("bad price %q", priceStr)
	}
	return PriceUpdate{
		Source:    source,
		Pair:      pair,
		Price:     price,
		Timestamp: uint64(time.Now().Unix()),
	}, nil
}

func MedianPrice(updates []PriceUpdate) *big.Int {
	if len(updates) == 0 {
		return big.NewInt(0)
	}
	sum := new(big.Int)
	for _, u := range updates {
		sum.Add(sum, u.Price)
	}
	return sum.Div(sum, big.NewInt(int64(len(updates))))
}
