package util

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"time"
)

func NormalizeSpace(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func FirstString(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return NormalizeSpace(typed)
	case []interface{}:
		for _, item := range typed {
			if text := FirstString(item); text != "" {
				return text
			}
		}
	case map[string]interface{}:
		for _, key := range []string{"fr", "en", "text", "value", "name", "label"} {
			if item, ok := typed[key]; ok {
				if text := FirstString(item); text != "" {
					return text
				}
			}
		}
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if text := FirstString(typed[key]); text != "" {
				return text
			}
		}
	}
	return ""
}

func BuildAddress(parts ...string) string {
	clean := make([]string, 0, len(parts))
	for _, part := range parts {
		part = NormalizeSpace(part)
		if part != "" {
			clean = append(clean, part)
		}
	}
	return strings.Join(clean, ", ")
}

func ParseDateTime(value interface{}) *time.Time {
	switch typed := value.(type) {
	case time.Time:
		t := typed.UTC()
		return &t
	case string:
		candidate := strings.TrimSpace(typed)
		if candidate == "" {
			return nil
		}
		candidate = strings.ReplaceAll(candidate, "Z", "+00:00")

		if parsed, err := time.Parse(time.RFC3339, candidate); err == nil {
			t := parsed.UTC()
			return &t
		}

		for _, layout := range []string{
			"2006-01-02 15:04:05",
			"2006-01-02 15:04",
			"2006-01-02",
			"02/01/2006 15:04:05",
			"02/01/2006 15:04",
			"02/01/2006",
		} {
			if parsed, err := time.Parse(layout, candidate); err == nil {
				t := parsed.UTC()
				return &t
			}
		}
	}
	return nil
}

func PayloadHash(value interface{}) string {
	encoded, _ := json.Marshal(value)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}
