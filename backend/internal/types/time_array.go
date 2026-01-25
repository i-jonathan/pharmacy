package types

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

type TimeArray []time.Time

// Scan implements sql.Scanner
func (t *TimeArray) Scan(src any) error {
	if src == nil {
		*t = []time.Time{}
		return nil
	}

	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("TimeArray: cannot scan %T", src)
	}

	// Postgres array format: {2025-07-20,2025-08-01,2025-08-15}
	s = strings.Trim(s, "{}")
	if s == "" {
		*t = []time.Time{}
		return nil
	}

	parts := strings.Split(s, ",")
	times := make([]time.Time, 0, len(parts))
	for _, p := range parts {
		tVal, err := time.Parse("2006-01-02", p)
		if err != nil {
			return fmt.Errorf("TimeArray: parse error %w", err)
		}
		times = append(times, tVal)
	}

	*t = times
	return nil
}

// Value implements driver.Valuer (optional if you ever insert/update)
func (t TimeArray) Value() (driver.Value, error) {
	strParts := make([]string, len(t))
	for i, tm := range t {
		strParts[i] = tm.Format("2006-01-02")
	}
	return "{" + strings.Join(strParts, ",") + "}", nil
}
