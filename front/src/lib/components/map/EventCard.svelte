<script lang="ts">
    import type { EventDetail, EventSummary } from "$lib/types/geo";

    interface Props {
        event: EventSummary;
        expanded?: boolean;
        detail?: EventDetail | null;
        detailLoading?: boolean;
        onSelect?: (event: EventSummary) => void;
        onClose?: (eventId?: string) => void;
    }

    let {
        event,
        expanded = false,
        detail = null,
        detailLoading = false,
        onSelect = () => {},
        onClose = () => {},
    }: Props = $props();

    function formatDate(value: string): string {
        return new Intl.DateTimeFormat("en-GB", {
            weekday: "short",
            day: "numeric",
            month: "short",
            hour: "2-digit",
            minute: "2-digit",
        }).format(new Date(value));
    }
</script>

<article class:expanded class="event-card">
    <header class="event-header">
        <div class="event-topline">
            <span class="source">{event.source_tag}</span>
            <time datetime={event.start_dt}>{formatDate(event.start_dt)}</time>
        </div>
        <h3>{event.title}</h3>
        <p>{event.address}</p>
    </header>

    <div class="event-actions">
        {#if expanded}
            <button class="toggle-button secondary" type="button" onclick={() => onClose(event.event_id)}>
                Close
            </button>
        {:else}
            <button class="toggle-button" type="button" onclick={() => onSelect(event)}>
                Open details
            </button>
        {/if}
    </div>

    {#if expanded}
        <section class="event-detail" aria-label={`Details for ${event.title}`}>
            {#if detailLoading}
                <p class="detail-state">Loading details…</p>
            {:else if detail}
                <p>{detail.description || "No description available."}</p>
                <a href={detail.source_url} target="_blank" rel="noreferrer">Open source</a>
            {:else}
                <p class="detail-state">No detail available for this event.</p>
            {/if}
        </section>
    {/if}
</article>

<style>
    .event-card {
        display: grid;
        gap: 0.75rem;
        padding: 0.9rem 0.95rem;
        border: 1px solid rgba(17, 24, 39, 0.08);
        border-radius: 1.1rem;
        background: white;
        box-shadow: var(--shadow-sm);
        text-align: left;
    }

    .event-card.expanded {
        border-color: rgba(255, 107, 53, 0.35);
        box-shadow: 0 14px 28px rgba(255, 107, 53, 0.14);
    }

    .event-header {
        display: grid;
        gap: 0.4rem;
    }

    .event-topline {
        display: flex;
        justify-content: space-between;
        gap: 0.75rem;
        font-size: var(--font-size-xs);
        color: var(--color-text-muted);
        text-transform: uppercase;
        letter-spacing: 0.06em;
    }

    .source {
        display: inline-flex;
        align-items: center;
        width: fit-content;
        padding: 0.22rem 0.55rem;
        border-radius: 999px;
        background: rgba(255, 107, 53, 0.12);
        color: var(--color-primary-dark);
        font-weight: 700;
    }

    h3 {
        font-size: 1rem;
        font-weight: 700;
        color: var(--color-text);
    }

    p {
        color: var(--color-text-secondary);
    }

    .event-actions {
        display: flex;
        justify-content: flex-end;
    }

    .toggle-button {
        padding: 0.55rem 0.8rem;
        border-radius: 999px;
        background: var(--color-text);
        color: white;
        font-size: var(--font-size-sm);
    }

    .toggle-button.secondary {
        background: rgba(17, 24, 39, 0.08);
        color: var(--color-text);
    }

    .event-detail {
        display: grid;
        gap: 0.75rem;
        padding-top: 0.2rem;
        border-top: 1px solid rgba(17, 24, 39, 0.08);
    }

    .event-detail a {
        font-weight: 700;
        color: var(--color-primary-dark);
    }

    .detail-state {
        color: var(--color-text-muted);
    }
</style>
