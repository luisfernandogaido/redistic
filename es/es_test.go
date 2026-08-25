package es

import (
	"os"
	"testing"
	"time"
)

func TestBulkIndex(t *testing.T) {
	hostname, _ := os.Hostname()
	data := struct {
		Action   string    `json:"action"`
		Time     time.Time `json:"time"`
		Hostname string    `json:"hostname"`
	}{
		Action:   "acao",
		Time:     time.Now(),
		Hostname: hostname,
	}
	BulkIndex("cnc", []any{data})
}
