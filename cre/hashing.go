package main

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

func NewRequestID() string {
	return fmt.Sprintf("req-%d", time.Now().UnixMilli())
}

func ReportDigest(r Report, extras map[string]string) []byte {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%x|", r.WorkflowID))
	sb.WriteString(fmt.Sprintf("%d|", r.Timestamp))
	sb.WriteString(fmt.Sprintf("%x|", r.Recipient))
	sb.WriteString(fmt.Sprintf("%s|", r.Amount.String()))
	sb.WriteString(fmt.Sprintf("%x|", r.Nonce))
	for k, v := range extras {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(v)
		sb.WriteString(";")
	}
	sum := sha256.Sum256([]byte(sb.String()))
	return sum[:]
}
