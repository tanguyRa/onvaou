<script lang="ts">
    import { onMount } from "svelte";

    import { useSession } from "$lib/auth-client";
    import EventPanel from "$lib/components/map/EventPanel.svelte";
    import LanguageSwitcher from "$lib/components/LanguageSwitcher.svelte";
    import MapView from "$lib/components/map/MapView.svelte";
    import RadiusSlider from "$lib/components/map/RadiusSlider.svelte";
    import SearchBar from "$lib/components/map/SearchBar.svelte";
    import { createMapDiscovery } from "$lib/components/map/discovery.svelte";

    const session = useSession();
    const discovery = createMapDiscovery("Lyon");

    onMount(() => {
        void discovery.initialize();
        return () => discovery.destroy();
    });
</script>

<main class="landing-map" id="map">
    <MapView
        events={discovery.events}
        center={discovery.center}
        radiusKm={discovery.radiusKm}
        selectedEventId={discovery.selectedEventId}
        onSelect={(event) => discovery.selectEvent(event)}
    />

    <header class="landing-header" aria-label="Site header">
        <div class="brand-block">
            <a class="brand" href="/">OnVaOu</a>
            <p class="brand-copy">
                Search nearby public events before creating an account.
            </p>
        </div>

        <div class="header-search">
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

        <nav class="header-actions" aria-label="Primary">
            <LanguageSwitcher />
            {#if $session.data?.user}
                <a class="action-link" href="/app">Open app</a>
            {:else}
                <a class="action-link" href="/login">Sign in</a>
            {/if}
            <a class="action-link primary" href="/register">Get started</a>
        </nav>
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
    {:else if discovery.loadingEvents && !discovery.hasResults}
        <output class="status-toast">Loading nearby events...</output>
    {:else if !discovery.loadingEvents && discovery.selectedCity && !discovery.hasResults}
        <output class="status-toast">No events found in this radius yet.</output>
    {/if}
</main>

<style>
    .landing-map {
        position: relative;
        width: 100vw;
        height: 100vh;
        overflow: hidden;
        background: #f5f7fb;
    }

    .landing-header {
        position: absolute;
        inset: 1rem 1rem auto 1rem;
        display: grid;
        grid-template-columns: minmax(0, 18rem) minmax(0, 1fr) auto;
        gap: 1rem;
        align-items: start;
        z-index: 1100;
    }

    .brand-block,
    .header-search,
    .header-actions,
    .status-toast {
        pointer-events: auto;
    }

    .brand-block {
        padding: 1rem 1.1rem;
        border-radius: 1.5rem;
        background: rgba(17, 24, 39, 0.68);
        color: white;
        backdrop-filter: blur(18px);
        box-shadow: var(--shadow-lg);
    }

    .brand {
        display: inline-block;
        margin-bottom: 0.45rem;
        font-family: var(--font-display);
        font-size: 1.6rem;
        font-weight: 700;
        color: white;
    }

    .brand-copy {
        color: rgba(255, 255, 255, 0.82);
    }

    .header-search {
        display: flex;
        gap: 1rem;
        align-items: flex-start;
    }

    .header-actions {
        display: flex;
        gap: 0.75rem;
        align-items: center;
        padding: 0.7rem 0.8rem;
        border-radius: 999px;
        background: rgba(255, 255, 255, 0.88);
        backdrop-filter: blur(18px);
        box-shadow: var(--shadow-md);
    }

    .action-link {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        min-height: 2.6rem;
        padding: 0.7rem 1rem;
        border-radius: 999px;
        font-weight: 600;
        color: var(--color-text);
    }

    .action-link.primary {
        background: var(--color-text);
        color: white;
    }

    .status-toast {
        position: absolute;
        left: 1rem;
        bottom: 1rem;
        max-width: min(28rem, calc(100vw - 2rem));
        padding: 0.95rem 1rem;
        border-radius: 1rem;
        background: rgba(255, 255, 255, 0.92);
        box-shadow: var(--shadow-lg);
        backdrop-filter: blur(16px);
        color: var(--color-text-muted);
        z-index: 1100;
    }

    .status-toast.error {
        background: var(--color-error-bg);
        color: var(--color-error);
        border: 1px solid var(--color-error-border);
    }

    @media (max-width: 1100px) {
        .landing-header {
            grid-template-columns: 1fr;
        }

        .header-search {
            flex-wrap: wrap;
        }

        .header-actions {
            justify-content: flex-start;
            width: fit-content;
        }
    }

    @media (max-width: 720px) {
        .landing-header {
            inset: 0.75rem 0.75rem auto 0.75rem;
        }

        .header-search {
            gap: 0.75rem;
        }

        .header-actions {
            width: 100%;
            flex-wrap: wrap;
            border-radius: 1.25rem;
        }

        .status-toast {
            left: 0.75rem;
            right: 0.75rem;
            bottom: 0.75rem;
            max-width: none;
        }
    }
</style>
