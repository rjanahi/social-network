package user

import (
	"fmt"
	"time"
)

// ComputeAgeFromDOB parses a date string (several common layouts) and returns
// the computed age in years and a normalized date string in YYYY-MM-DD form.
// It returns an error if parsing fails.
func ComputeAgeFromDOB(dob string) (int, string, error) {
	if dob == "" {
		return 0, "", fmt.Errorf("empty dob")
	}

	// Try common layouts returned by HTML date input and possible DB formats
	layouts := []string{
		"2006-01-02",          // HTML date input
		"2006-01-02 15:04:05", // common SQL datetime
		time.RFC3339,          // ISO with timezone
		"2006-01-02T15:04:05", // ISO without zone
	}

	var t time.Time
	var err error
	for _, l := range layouts {
		t, err = time.Parse(l, dob)
		if err == nil {
			break
		}
	}
	if err != nil {
		return 0, "", fmt.Errorf("could not parse dob: %w", err)
	}

	now := time.Now()
	years := now.Year() - t.Year()
	// adjust if birthday hasn't occurred yet this year
	if now.Month() < t.Month() || (now.Month() == t.Month() && now.Day() < t.Day()) {
		years--
	}

	normalized := t.Format("2006-01-02")
	return years, normalized, nil
}
