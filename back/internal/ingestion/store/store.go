package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tanguyRa/onvaou/internal/ingestion/artifacts"
	"github.com/tanguyRa/onvaou/internal/ingestion/dedup"
	"github.com/tanguyRa/onvaou/internal/ingestion/model"
)

type Store struct {
	pool      *pgxpool.Pool
	artifacts *artifacts.Store
}

func New(pool *pgxpool.Pool, artifactStore *artifacts.Store) *Store {
	return &Store{
		pool:      pool,
		artifacts: artifactStore,
	}
}

func (s *Store) UpsertEventBatch(ctx context.Context, sourceTag string, events []model.Event) (model.IngestionResult, error) {
	result := model.NewResult(sourceTag)
	result.Fetched = len(events)

	for _, event := range events {
		event = event.WithEventID()

		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return result, err
		}

		status, upsertErr := s.upsertEvent(ctx, tx, event)
		if upsertErr != nil {
			_ = tx.Rollback(ctx)
			artifactID, artifactErr := s.artifacts.Persist(
				event.SourceTag,
				"normalized_event",
				"upsert",
				event.SourceUID,
				event,
				map[string]interface{}{"error": upsertErr.Error()},
			)
			result.Failed++
			detail := fmt.Sprintf("%s: %v", event.SourceUID, upsertErr)
			if artifactErr == nil {
				detail += fmt.Sprintf(" [artifact_id=%s]", artifactID)
			}
			result.Details = append(result.Details, detail)
			continue
		}

		if err := tx.Commit(ctx); err != nil {
			return result, err
		}

		switch status {
		case "duplicate":
			result.Duplicates++
		case "updated":
			result.Updated++
		default:
			result.Inserted++
		}
	}

	return result, nil
}

func (s *Store) upsertEvent(ctx context.Context, tx pgx.Tx, event model.Event) (string, error) {
	decision, err := dedup.CheckDuplicate(ctx, tx, event)
	if err != nil {
		return "", err
	}

	if decision.Action == "exact" {
		if _, err := tx.Exec(
			ctx,
			`
			INSERT INTO source_hashes (source_tag, content_hash, event_id)
			VALUES ($1, $2, $3)
			ON CONFLICT (source_tag, content_hash) DO NOTHING
			`,
			event.SourceTag,
			event.ContentHash(),
			decision.EventID,
		); err != nil {
			return "", err
		}
		return "duplicate", nil
	}

	if decision.Action == "near" && decision.EventID != uuid.Nil {
		event.EventID = decision.EventID
	}

	if _, err := tx.Exec(
		ctx,
		`
		INSERT INTO events (
			event_id,
			title,
			description,
			start_dt,
			end_dt,
			location_name,
			address,
			geom,
			source_tag,
			source_url
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			ST_SetSRID(ST_MakePoint($8, $9), 4326)::geography,
			$10,
			$11
		)
		ON CONFLICT (event_id) DO UPDATE SET
			title = EXCLUDED.title,
			description = CASE
				WHEN EXCLUDED.description = '' THEN events.description
				ELSE EXCLUDED.description
			END,
			start_dt = EXCLUDED.start_dt,
			end_dt = COALESCE(EXCLUDED.end_dt, events.end_dt),
			location_name = CASE
				WHEN EXCLUDED.location_name = '' THEN events.location_name
				ELSE EXCLUDED.location_name
			END,
			address = EXCLUDED.address,
			geom = EXCLUDED.geom,
			source_tag = events.source_tag,
			source_url = CASE
				WHEN EXCLUDED.source_url = '' THEN events.source_url
				ELSE EXCLUDED.source_url
			END
		`,
		event.EventID,
		event.Title,
		event.Description,
		event.StartDT,
		event.EndDT,
		event.LocationName,
		event.Address,
		event.Longitude,
		event.Latitude,
		event.SourceTag,
		event.SourceURL,
	); err != nil {
		return "", err
	}

	if _, err := tx.Exec(
		ctx,
		`
		INSERT INTO source_hashes (source_tag, content_hash, event_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (source_tag, content_hash) DO UPDATE SET
			event_id = EXCLUDED.event_id
		`,
		event.SourceTag,
		event.ContentHash(),
		event.EventID,
	); err != nil {
		return "", err
	}

	if decision.Action == "near" {
		return "updated", nil
	}
	return "inserted", nil
}

func (s *Store) UpsertOpenAgendaAgendas(ctx context.Context, agendas []model.OpenAgendaAgenda) (int, int, error) {
	inserted := 0
	updated := 0

	for _, agenda := range agendas {
		var wasInserted bool
		err := s.pool.QueryRow(
			ctx,
			`
			INSERT INTO openagenda_agendas (
				uid,
				slug,
				title,
				description,
				official,
				updated_at,
				upcoming_events,
				recently_added_events,
				source_url,
				last_payload_hash,
				last_seen_at,
				last_fetch_success_at,
				last_fetch_error
			)
			VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
				COALESCE($11, CURRENT_TIMESTAMP), CURRENT_TIMESTAMP, ''
			)
			ON CONFLICT (uid) DO UPDATE SET
				slug = EXCLUDED.slug,
				title = EXCLUDED.title,
				description = EXCLUDED.description,
				official = EXCLUDED.official,
				updated_at = EXCLUDED.updated_at,
				upcoming_events = EXCLUDED.upcoming_events,
				recently_added_events = EXCLUDED.recently_added_events,
				source_url = EXCLUDED.source_url,
				last_payload_hash = EXCLUDED.last_payload_hash,
				last_seen_at = EXCLUDED.last_seen_at,
				last_fetch_success_at = CURRENT_TIMESTAMP,
				last_fetch_error = ''
			RETURNING (xmax = 0) AS inserted
			`,
			agenda.UID,
			agenda.Slug,
			agenda.Title,
			agenda.Description,
			agenda.Official,
			agenda.UpdatedAt,
			agenda.UpcomingEvents,
			agenda.RecentlyAddedEvents,
			agenda.SourceURL,
			agenda.LastPayloadHash,
			agenda.LastSeenAt,
		).Scan(&wasInserted)
		if err != nil {
			return 0, 0, err
		}

		if wasInserted {
			inserted++
		} else {
			updated++
		}
	}

	return inserted, updated, nil
}

func (s *Store) StoreOpenAgendaAgendaFetch(ctx context.Context, batchHash string, rawAgendas []map[string]interface{}) (bool, error) {
	payload, err := json.Marshal(rawAgendas)
	if err != nil {
		return false, err
	}

	var inserted int
	err = s.pool.QueryRow(
		ctx,
		`
		INSERT INTO openagenda_agenda_fetches (batch_hash, agenda_count, raw_payload)
		VALUES ($1, $2, $3::jsonb)
		ON CONFLICT (batch_hash) DO NOTHING
		RETURNING 1
		`,
		batchHash,
		len(rawAgendas),
		string(payload),
	).Scan(&inserted)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return inserted == 1, nil
}

func (s *Store) ListChangedOpenAgendaAgendas(ctx context.Context, agendas []model.OpenAgendaAgenda) ([]model.OpenAgendaAgenda, error) {
	if len(agendas) == 0 {
		return nil, nil
	}

	uids := make([]int64, 0, len(agendas))
	for _, agenda := range agendas {
		uids = append(uids, agenda.UID)
	}

	rows, err := s.pool.Query(
		ctx,
		`
		SELECT uid, last_payload_hash
		FROM openagenda_agendas
		WHERE uid = ANY($1::bigint[])
		`,
		uids,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	known := make(map[int64]string, len(agendas))
	for rows.Next() {
		var uid int64
		var lastPayloadHash string
		if err := rows.Scan(&uid, &lastPayloadHash); err != nil {
			return nil, err
		}
		known[uid] = lastPayloadHash
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	out := make([]model.OpenAgendaAgenda, 0, len(agendas))
	for _, agenda := range agendas {
		if known[agenda.UID] != agenda.LastPayloadHash {
			out = append(out, agenda)
		}
	}
	return out, nil
}

func (s *Store) StoreOpenAgendaAgendaPayloads(ctx context.Context, agendas []model.OpenAgendaAgenda) (int, error) {
	if len(agendas) == 0 {
		return 0, nil
	}

	stored := 0
	for _, agenda := range agendas {
		if _, err := s.pool.Exec(
			ctx,
			`
			INSERT INTO openagenda_agendas (uid, title, last_fetch_attempt_at)
			VALUES ($1, $2, CURRENT_TIMESTAMP)
			ON CONFLICT (uid) DO UPDATE SET
				last_fetch_attempt_at = CURRENT_TIMESTAMP
			`,
			agenda.UID,
			agenda.Title,
		); err != nil {
			return 0, err
		}

		payload, err := json.Marshal(agenda)
		if err != nil {
			return 0, err
		}

		var inserted int
		err = s.pool.QueryRow(
			ctx,
			`
			INSERT INTO openagenda_agenda_payloads (uid, payload_hash, raw_payload)
			VALUES ($1, $2, $3::jsonb)
			ON CONFLICT (uid, payload_hash) DO NOTHING
			RETURNING 1
			`,
			agenda.UID,
			agenda.LastPayloadHash,
			string(payload),
		).Scan(&inserted)
		if err == pgx.ErrNoRows {
			continue
		}
		if err != nil {
			return 0, err
		}
		stored++
	}

	return stored, nil
}

func (s *Store) ListRelevantOpenAgendaAgendaIDs(ctx context.Context, limit int) ([]int64, error) {
	query := `
		SELECT uid
		FROM openagenda_agendas
		WHERE upcoming_events > 0
		ORDER BY official DESC, upcoming_events DESC, updated_at DESC NULLS LAST
	`
	args := []interface{}{}
	if limit > 0 {
		query += " LIMIT $1"
		args = append(args, limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var uid int64
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		ids = append(ids, uid)
	}
	return ids, rows.Err()
}

func (s *Store) HasOpenAgendaEventBatchChanged(ctx context.Context, agendaUID int64, batchHash string) (bool, error) {
	var storedHash string
	err := s.pool.QueryRow(
		ctx,
		`
		SELECT last_event_batch_hash
		FROM openagenda_agendas
		WHERE uid = $1
		`,
		agendaUID,
	).Scan(&storedHash)
	if err == pgx.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return storedHash != batchHash, nil
}

func (s *Store) StoreOpenAgendaEventFetch(ctx context.Context, agendaUID int64, batchHash string, rawEvents []map[string]interface{}) (bool, error) {
	payload, err := json.Marshal(rawEvents)
	if err != nil {
		return false, err
	}

	var inserted int
	err = s.pool.QueryRow(
		ctx,
		`
		INSERT INTO openagenda_event_fetches (agenda_uid, batch_hash, event_count, raw_payload)
		VALUES ($1, $2, $3, $4::jsonb)
		ON CONFLICT (agenda_uid, batch_hash) DO NOTHING
		RETURNING 1
		`,
		agendaUID,
		batchHash,
		len(rawEvents),
		string(payload),
	).Scan(&inserted)
	if err != nil && err != pgx.ErrNoRows {
		return false, err
	}

	if _, err := s.pool.Exec(
		ctx,
		`
		UPDATE openagenda_agendas
		SET last_fetch_attempt_at = CURRENT_TIMESTAMP
		WHERE uid = $1
		`,
		agendaUID,
	); err != nil {
		return false, err
	}

	return inserted == 1, nil
}

func (s *Store) MarkOpenAgendaAgendaSynced(ctx context.Context, agendaUID int64) error {
	_, err := s.pool.Exec(
		ctx,
		`
		UPDATE openagenda_agendas
		SET last_event_sync_at = $2
		WHERE uid = $1
		`,
		agendaUID,
		time.Now().UTC(),
	)
	return err
}

func (s *Store) MarkOpenAgendaAgendaSyncResult(ctx context.Context, agendaUID int64, batchHash string, fetchError string) error {
	_, err := s.pool.Exec(
		ctx,
		`
		UPDATE openagenda_agendas
		SET last_fetch_success_at = CASE
				WHEN $2 = '' THEN CURRENT_TIMESTAMP
				ELSE last_fetch_success_at
			END,
			last_event_batch_hash = COALESCE(NULLIF($3, ''), last_event_batch_hash),
			last_fetch_error = $2
		WHERE uid = $1
		`,
		agendaUID,
		fetchError,
		batchHash,
	)
	return err
}

func (s *Store) VacuumEvents(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, "VACUUM ANALYZE events"); err != nil {
		return err
	}
	if _, err := s.pool.Exec(ctx, "VACUUM ANALYZE source_hashes"); err != nil {
		return err
	}
	return nil
}
