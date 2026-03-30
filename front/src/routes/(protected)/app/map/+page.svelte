<script lang="ts">
    import { onMount } from "svelte";

    import EventPanel from "$lib/components/map/EventPanel.svelte";
    import MapView from "$lib/components/map/MapView.svelte";
    import RadiusSlider from "$lib/components/map/RadiusSlider.svelte";
    import SearchBar from "$lib/components/map/SearchBar.svelte";
    import { createMapDiscovery } from "$lib/components/map/discovery.svelte";

    const discovery = createMapDiscovery("Lyon");

    onMount(() => {
        void discovery.initialize();
        return () => discovery.destroy();
    });
</script>

<section class="app-map" aria-label="Event discovery map">
    <MapView
        events={discovery.events}
        center={discovery.center}
        radiusKm={discovery.radiusKm}
        selectedEventId={discovery.selectedEventId}
        onSelect={(event) => discovery.selectEvent(event)}
    />

    <header class="map-header">
        <div class="map-intro">
            <p class="kicker">Discovery</p>
            <h1>Local events on the map</h1>
        </div>

        <div class="map-controls">
            <SearchBar
                value={discovery.searchValue}
                suggestions={discovery.suggestions}
                loading={discovery.loadingSuggestions}
                onInput={(value) => discovery.setSearchValue(value)}
                onSelect={(city) => discovery.selectCity(city)}
                onSubmit={() => discovery.submitSearch()}
            />
            <RadiusSlider
                value={discovery.radiusKm}
                onInput={(value) => discovery.setRadius(value)}
            />
        </div>
    </header>

    <EventPanel
        events={discovery.events}
        total={discovery.total}
        cityLabel={discovery.cityLabel}
        radiusKm={discovery.radiusKm}
        loading={discovery.loadingEvents}
        selectedEventId={discovery.selectedEventId}
        selectedEvent={discovery.selectedEvent}
        detailLoading={discovery.detailLoading}
        onSelect={(event) => discovery.selectEvent(event)}
        onClose={(eventId) => discovery.closeEvent(eventId)}
    />

    {#if discovery.error}
        <output class="status-toast error">{discovery.error}</output>
    {/if}
</section>

<style>
    .app-map {
        position: relative;
        width: 100%;
        height: 100%;
        overflow: hidden;
    }

    .map-header {
        position: absolute;
        inset: 1rem 1rem auto 1rem;
        display: flex;
        justify-content: space-between;
        gap: 1rem;
        align-items: start;
        z-index: 1100;
    }

    .map-intro {
        padding: 1rem 1.1rem;
        border-radius: 1.5rem;
        background: rgba(17, 24, 39, 0.68);
        color: white;
        box-shadow: var(--shadow-lg);
        backdrop-filter: blur(18px);
    }

    .kicker {
        margin-bottom: 0.35rem;
        font-size: var(--font-size-xs);
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        color: #ffd0bf;
    }

    .map-intro h1 {
        margin: 0;
        font-family: var(--font-display);
        font-size: clamp(1.6rem, 3vw, 2.4rem);
        line-height: 0.95;
    }

    .map-controls {
        display: flex;
        gap: 1rem;
        align-items: flex-start;
    }

    .status-toast {
        position: absolute;
        left: 1rem;
        bottom: 1rem;
        padding: 0.95rem 1rem;
        border-radius: 1rem;
        background: var(--color-error-bg);
        color: var(--color-error);
        border: 1px solid var(--color-error-border);
        z-index: 1100;
    }

    @media (max-width: 1100px) {
        .map-header {
            flex-direction: column;
        }

        .map-controls {
            width: 100%;
            flex-wrap: wrap;
        }
    }

    @media (max-width: 720px) {
        .map-header {
            inset: 0.75rem 0.75rem auto 0.75rem;
        }
    }
</style>
