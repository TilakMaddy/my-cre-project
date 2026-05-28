package main

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
)

const ProtocolVersion byte = 1

func ParsePayload(b []byte) (*Report, error) {
	if len(b) < 1 {
		return nil, errors.New("payload too short")
	}
	return decodeReportV1(b[1:])
}

func decodeReportV1(b []byte) (*Report, error) {
	const expectedLen = 32 + 8 + 20 + 32
	if len(b) < expectedLen {
		return nil, errors.New("payload truncated")
	}
	r := &Report{}
	copy(r.WorkflowID[:], b[0:32])
	r.Timestamp = binary.BigEndian.Uint64(b[32:40])
	copy(r.Recipient[:], b[40:60])
	r.Amount = new(big.Int).SetBytes(b[60:92])
	return r, nil
}

func EncodeReport(r Report) []byte {
	buf := make([]byte, 0, 1+32+8+20+32)
	buf = append(buf, ProtocolVersion)
	buf = append(buf, r.WorkflowID[:]...)
	var ts [8]byte
	binary.BigEndian.PutUint64(ts[:], r.Timestamp)
	buf = append(buf, ts[:]...)
	buf = append(buf, r.Recipient[:]...)
	amt := make([]byte, 32)
	r.Amount.FillBytes(amt)
	buf = append(buf, amt...)
	return buf
}

func IsValidAddress(s string) bool {
	return len(s) == 42
}

func IsValidTxHash(s string) bool {
	return len(s) == 66
}

func ParseAddress(s string) ([20]byte, error) {
	var out [20]byte
	if !IsValidAddress(s) {
		return out, errors.New("invalid address length")
	}
	raw, err := hex.DecodeString(strings.TrimPrefix(s, "0x"))
	if err != nil {
		return out, err
	}
	copy(out[:], raw)
	return out, nil
}
