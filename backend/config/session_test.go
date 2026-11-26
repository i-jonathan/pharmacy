package config

import (
	"testing"
	"time"
)

func TestSessionExpiryIsTheNext3AM(t *testing.T) {
	lagos := time.FixedZone("WAT", 3600)
	cases := []struct{ now, want time.Time }{
		{
			time.Date(2026, 11, 26, 2, 59, 0, 0, lagos),
			time.Date(2026, 11, 26, 3, 0, 0, 0, lagos),
		},
		{
			time.Date(2026, 11, 26, 3, 0, 0, 0, lagos),
			time.Date(2026, 11, 27, 3, 0, 0, 0, lagos),
		},
		{
			time.Date(2026, 12, 31, 22, 0, 0, 0, lagos),
			time.Date(2027, 1, 1, 3, 0, 0, 0, lagos),
		},
	}
	for _, c := range cases {
		if got := SessionExpiry(c.now); !got.Equal(c.want) {
			t.Errorf("SessionExpiry(%v) = %v, want %v", c.now, got, c.want)
		}
	}
}
