<script lang="ts">
    import { t } from "$lib/i18n/index.svelte";
    import { useUser } from "$lib/stores/user.svelte";

    const user = useUser();
    const isPremium = $derived(user.state.hasActiveSubscription);
</script>

<div class="dashboard">
    <header class="dashboard-header">
        <div class="header-content">
            <h1>
                {t("dashboard.header.welcome")}
                {user.state.user?.name || t("dashboard.header.userFallback")}
            </h1>
            {#if !isPremium}
                <span class="badge tier-badge free"
                    >{t("dashboard.header.freePlan")}</span
                >
            {:else}
                <span class="badge tier-badge premium"
                    >{t("dashboard.header.premiumPlan")}</span
                >
            {/if}
        </div>
    </header>

    <section class="map-entry">
        <div>
            <p class="map-entry-label">Discovery</p>
            <h2>Open the live event map</h2>
            <p class="map-entry-copy">
                Search cities, inspect nearby public events, and browse results
                on an interactive map.
            </p>
        </div>
        <a class="btn btn-primary" href="/app/map">Open map</a>
    </section>
</div>

<style>
    .dashboard {
        padding: var(--spacing-2xl);
        max-width: 1280px;
        margin: 0 auto;
    }

    .dashboard-header {
        display: flex;
        align-items: center;
        margin-bottom: var(--spacing-2xl);
    }

    .header-content {
        display: flex;
        align-items: center;
        gap: var(--spacing-md);
    }

    .dashboard-header h1 {
        font-size: var(--font-size-2xl);
        font-weight: 600;
        color: var(--color-text);
    }

    .tier-badge {
        text-transform: uppercase;
        letter-spacing: 0.5px;
    }

    .tier-badge.free {
        background: var(--color-bg-tertiary);
        color: var(--color-text-muted);
    }

    .tier-badge.premium {
        background: var(--gradient-primary);
        color: white;
    }

    .map-entry {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: var(--spacing-lg);
        padding: var(--spacing-xl);
        border-radius: var(--radius-lg);
        background:
            radial-gradient(circle at top right, rgba(255, 154, 118, 0.25), transparent 32%),
            linear-gradient(135deg, #fff7f2 0%, #ffffff 55%, #f5f7fb 100%);
        box-shadow: var(--shadow-md);
    }

    .map-entry-label {
        margin-bottom: var(--spacing-xs);
        font-size: var(--font-size-xs);
        font-weight: 700;
        letter-spacing: 0.08em;
        text-transform: uppercase;
        color: var(--color-primary-dark);
    }

    .map-entry h2 {
        margin-bottom: var(--spacing-sm);
        text-align: left;
    }

    .map-entry-copy {
        max-width: 56ch;
        color: var(--color-text-muted);
    }

    @media (max-width: 720px) {
        .map-entry {
            flex-direction: column;
            align-items: stretch;
        }
    }
</style>
