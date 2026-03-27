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
</style>
