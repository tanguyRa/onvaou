<script lang="ts">
    import { useSession } from "$lib/auth-client";
    import LanguageSwitcher from "$lib/components/LanguageSwitcher.svelte";
    import { t } from "$lib/i18n/index.svelte";

    const session = useSession();
</script>

<div class="landing">
    <header class="topbar row">
        <a class="brand" href="/">SaaS Seed</a>
        <nav class="row muted">
            <a href="#features">{t("landing.nav.features")}</a>
            <a href="/pricing">{t("landing.nav.pricing")}</a>
        </nav>
        <div class="actions row">
            <LanguageSwitcher />
            {#if $session.data?.user}
                <a class="btn btn-ghost" href="/app">{t("landing.nav.openApp")}</a>
            {:else}
                <a class="btn btn-ghost" href="/login">{t("landing.nav.signIn")}</a>
            {/if}
            <a class="btn primary" href="/register">{t("landing.nav.getStarted")}</a>
        </div>
    </header>

    <article class="landing-main">
        <section class="hero">
            <p class="eyebrow primary">{t("landing.hero.eyebrow")}</p>
            <h1>{t("landing.hero.title")}</h1>
            <p class="muted">{t("landing.hero.subtitle")}</p>
            <div class="hero-actions row">
                <a class="btn primary" href="/register">{t("landing.hero.primaryCta")}</a>
                <a class="btn secondary" href="/pricing">{t("landing.hero.secondaryCta")}</a>
            </div>
        </section>

        <section id="features" class="features column">
            <h2>{t("landing.features.title")}</h2>
            <p class="muted">{t("landing.features.subtitle")}</p>
            <div class="cards">
                <article class="card feature-card">
                    <h3>{t("landing.features.cards.i18n.title")}</h3>
                    <p class="muted">{t("landing.features.cards.i18n.body")}</p>
                </article>
                <article class="card feature-card">
                    <h3>{t("landing.features.cards.auth.title")}</h3>
                    <p class="muted">{t("landing.features.cards.auth.body")}</p>
                </article>
                <article class="card feature-card">
                    <h3>{t("landing.features.cards.billing.title")}</h3>
                    <p class="muted">{t("landing.features.cards.billing.body")}</p>
                </article>
            </div>
        </section>
    </article>

    <footer class="footer row muted">
        <span>{t("landing.footer.text")}</span>
        <div class="row">
            <a href="/pricing">{t("landing.footer.pricing")}</a>
            <a href="/login">{t("landing.footer.login")}</a>
            <a href="/register">{t("landing.footer.register")}</a>
        </div>
    </footer>
</div>

<style>
    .landing {
        min-height: 100vh;
        display: grid;
        grid-template-rows: auto 1fr auto;
    }

    .topbar {
        max-width: var(--container-xl);
        margin: 0 auto;
        padding: var(--xl);
        flex-wrap: wrap;
    }

    .brand {
        font-weight: 700;
        font-size: var(--xl);
    }

    .landing-main {
        margin: 0 auto;
        width: min(100%, var(--container-xl));
        gap: var(--xxxl);
    }

    .hero {
        max-width: 56ch;
    }

    .eyebrow {
        display: inline-flex;
        font-weight: 700;
        border-radius: var(--sm);
        padding: 0.2rem 0.75rem;
        width: fit-content;
    }

    h1 {
        font-size: clamp(2.2rem, 6vw, 4rem);
    }

    .hero-actions {
        flex-wrap: wrap;
        justify-content: flex-start;
    }

    .cards {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
        gap: var(--lg);
    }

    .feature-card {
        gap: var(--md);
    }

    .footer {
        max-width: var(--container-xl);
        margin: 0 auto;
        padding: var(--xl);
        flex-wrap: wrap;
    }

    @media (max-width: 768px) {
        .topbar,
        .footer {
            justify-content: center;
        }

        .topbar .actions {
            width: 100%;
            justify-content: center;
        }
    }
</style>
