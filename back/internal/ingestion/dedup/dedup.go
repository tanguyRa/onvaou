package dedup

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tanguyRa/onvaou/internal/ingestion/model"
)

type Decision struct {
	EventID uuid.UUID
	Action  string
}

func CheckDuplicate(ctx context.Context, tx pgx.Tx, event model.Event) (Decision, error) {
	var existingID uuid.UUID
	err := tx.QueryRow(
		ctx,
		`
		SELECT event_id
		FROM source_hashes
		WHERE source_tag = $1 AND content_hash = $2
		`,
		event.SourceTag,
		event.ContentHash(),
	).Scan(&existingID)
	if err == nil {
		return Decision{EventID: existingID, Action: "exact"}, nil
	}
	if err != pgx.ErrNoRows {
		return Decision{}, err
	}

	rows, err := tx.Query(
		ctx,
		`
		SELECT event_id, title, address
		FROM events
		WHERE start_dt BETWEEN $1 AND $2
		LIMIT 50
		`,
		event.StartDT.Add(-12*time.Hour),
		event.StartDT.Add(12*time.Hour),
	)
	if err != nil {
		return Decision{}, err
	}
	defer rows.Close()

	bestScore := 0
	bestID := uuid.Nil

	for rows.Next() {
		var candidateID uuid.UUID
		var title string
		var address string
		if err := rows.Scan(&candidateID, &title, &address); err != nil {
			return Decision{}, err
		}

		haystack := strings.TrimSpace(strings.ToLower(title) + " " + strings.ToLower(address))
		if haystack == "" {
			continue
		}

		score := tokenSortRatio(event.DedupText(), haystack)
		if score >= 85 && score > bestScore {
			bestScore = score
			bestID = candidateID
		}
	}
	if rows.Err() != nil {
		return Decision{}, rows.Err()
	}

	if bestID != uuid.Nil {
		return Decision{EventID: bestID, Action: "near"}, nil
	}

	return Decision{Action: "new"}, nil
}

func tokenSortRatio(left, right string) int {
	left = sortTokens(left)
	right = sortTokens(right)
	if left == "" && right == "" {
		return 100
	}
	maxLen := len([]rune(left))
	if other := len([]rune(right)); other > maxLen {
		maxLen = other
	}
	if maxLen == 0 {
		return 100
	}
	dist := levenshtein([]rune(left), []rune(right))
	score := (1.0 - float64(dist)/float64(maxLen)) * 100
	if score < 0 {
		return 0
	}
	return int(score + 0.5)
}

func sortTokens(value string) string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(value)))
	sort.Strings(fields)
	return strings.Join(fields, " ")
}

func levenshtein(left, right []rune) int {
	if len(left) == 0 {
		return len(right)
	}
	if len(right) == 0 {
		return len(left)
	}

	prev := make([]int, len(right)+1)
	curr := make([]int, len(right)+1)
	for j := 0; j <= len(right); j++ {
		prev[j] = j
	}
	for i := 1; i <= len(left); i++ {
		curr[0] = i
		for j := 1; j <= len(right); j++ {
			cost := 0
			if left[i-1] != right[j-1] {
				cost = 1
			}
			curr[j] = min(
				curr[j-1]+1,
				prev[j]+1,
				prev[j-1]+cost,
			)
		}
		prev, curr = curr, prev
	}
	return prev[len(right)]
}

func min(values ...int) int {
	out := values[0]
	for _, value := range values[1:] {
		if value < out {
			out = value
		}
	}
	return out
}
