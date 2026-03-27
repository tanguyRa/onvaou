-- name: CountGeoEvents :one
SELECT COUNT(*) FROM "events";

-- name: ListUpcomingGeoEvents :many
SELECT
    event_id,
    title,
    start_dt,
    location_name,
    source_tag,
    source_url
FROM "events"
WHERE start_dt >= NOW()
ORDER BY start_dt ASC
LIMIT $1;
