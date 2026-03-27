package utils

import (
	"fmt"
	"time"
)

var timeLayouts = []string{
	time.RFC3339,
	"2006-01-02 15:04:05Z",
	"2006-01-02T15:04:05Z",
	"2006-01-02 15:04:05 Z",
	"2006-01-02T15:04:05.999999",
	"2006-01-02T15:04:05.999999Z",
	"2006-01-02T15:04:05.000Z",
	"2006-01-02",
}

func ParseTime(s string) (time.Time, error) {
	for _, layout := range timeLayouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("failed to parse time %s", s)
}
