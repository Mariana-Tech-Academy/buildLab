package service

import (
	"testing"
	"time"
)

// TestFormatOrderNumber checks that order numbers are built in the format ORD-YYYYMMDD-NNN.
func TestFormatOrderNumber(t *testing.T) {
	date := time.Date(2026, time.October, 29, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		sequence int
		want     string
	}{
		{sequence: 1, want: "ORD-20261029-001"},
		{sequence: 15, want: "ORD-20261029-015"},
		{sequence: 999, want: "ORD-20261029-999"},
	}

	for _, tt := range tests {
		got := FormatOrderNumber(date, tt.sequence)
		if got != tt.want {
			t.Errorf("FormatOrderNumber(%d) = %s, want %s", tt.sequence, got, tt.want)
		}
	}
}
