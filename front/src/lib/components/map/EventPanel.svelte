<script lang="ts">
    import type { EventDetail, EventSummary } from "$lib/types/geo";
    import EventCard from "./EventCard.svelte";

    interface Props {
        events?: EventSummary[];
        total?: number;
        cityLabel?: string;
        radiusKm?: number;
        loading?: boolean;
        selectedEventId?: string | null;
        selectedEvent?: EventDetail | null;
        detailLoading?: boolean;
        onSelect?: (event: EventSummary) => void;
        onClose?: (eventId?: string) => void;
    }

    let {
        events = [],
        total = 0,
        cityLabel = "",
        radiusKm = 10,
        loading = false,
        selectedEventId = null,
        selectedEvent = null,
        detailLoading = false,
        onSelect = () => {},
        onClose = () => {},
    }: Props = $props();
</script>

{#if events.length > 0}
    <aside class="event-panel" aria-label="Nearby events">
        <header class="panel-header">
            <div>
                <p class="panel-kicker">Nearby events</p>
                <h2>{loading ? "Refreshing events..." : `${total} event${total === 1 ? "" : "s"}`}</h2>
            </div>
            <p class="panel-meta">{cityLabel} · {radiusKm} km</p>
        </header>

        <ol class="event-list">
            {#each events as event (event.event_id)}
                <li>
                    <EventCard
                        {event}
                        expanded={event.event_id === selectedEventId}
                        detail={event.event_id === selectedEventId
                            ? selectedEvent
                            : null}
                        detailLoading={event.event_id === selectedEventId
                            ? detailLoading
                            : false}
                        onSelect={onSelect}
                        onClose={onClose}
                    />
                </li>
            {/each}
        </ol>
    </aside>
{/if}

<style>
    .event-panel {
        position: absolute;
        top: 1rem;
        right: 1rem;
        bottom: 1rem;
        width: min(24rem, calc(100vw - 2rem));
        display: grid;
        grid-template-rows: auto 1fr;
        gap: 1rem;
        padding: 1rem;
        border: 1px solid rgba(255, 255, 255, 0.45);
        border-radius: 1.5rem;
        background: rgba(255, 255, 255, 0.92);
        box-shadow: var(--shadow-xl);
        backdrop-filter: blur(18px);
        z-index: 1100;
    }

    .panel-header {
        display: flex;
        justify-content: space-between;
        gap: 1rem;
        align-items: end;
        padding-bottom: 0.75rem;
        border-bottom: 1px solid rgba(17, 24, 39, 0.08);
    }

    .panel-kicker {
        margin-bottom: 0.25rem;
        font-size: var(--font-size-xs);
        font-weight: 700;
        letter-spacing: 0.08em;
        text-transform: uppercase;
        color: var(--color-primary-dark);
    }

    .panel-header h2 {
        margin: 0;
        font-size: 1.35rem;
        line-height: 1.1;
    }

    .panel-meta {
        color: var(--color-text-muted);
        white-space: nowrap;
    }

    .event-list {
        display: grid;
        gap: 0.75rem;
        overflow-y: auto;
        list-style: none;
    }

    @media (max-width: 900px) {
        .event-panel {
            top: auto;
            left: 0.75rem;
            right: 0.75rem;
            bottom: 0.75rem;
            width: auto;
            max-height: 48vh;
        }

        .panel-header {
            align-items: start;
            flex-direction: column;
        }
    }
</style>
