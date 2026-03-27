<script lang="ts">
    import {
        changeLanguage,
        getLanguage,
        getLanguages,
    } from "$lib/i18n/index.svelte";

    let open = $state(false);
    const languages = getLanguages();

    function select(code: string) {
        console.log("Selecting language:", code);
        changeLanguage(code);
        open = false;
    }

    function handleClickOutside(e: MouseEvent) {
        const target = e.target as HTMLElement;
        if (!target.closest(".language-switcher")) {
            open = false;
        }
    }
</script>

<svelte:document onclick={handleClickOutside} />

<div class="language-switcher">
    <div class="language-switcher-content">
        <button class="current" onclick={() => (open = !open)}>
            {languages.find((l) => l.code === getLanguage())?.label ?? "FR"}
        </button>
        {#if open}
            <ul class="dropdown">
                {#each languages.filter((l) => l.code !== getLanguage()) as lang}
                    <li>
                        <button onclick={() => select(lang.code)}>
                            {lang.label}
                        </button>
                    </li>
                {/each}
            </ul>
        {/if}
    </div>
</div>

<style>
    .language-switcher {
        position: relative;
        z-index: 1;
        display: flex;
        align-items: center;
        justify-content: center;
    }

    .language-switcher-content {
        display: flex;
        flex-direction: row;
        background: var(--color-surface-90);
        border: 1px solid var(--color-border-light);
        box-shadow: var(--shadow-sm);
        border-radius: 999px;
        backdrop-filter: blur(10px);
    }

    button {
        display: flex;
        align-items: center;
        justify-content: center;
        background: none;
        border: none;
        cursor: pointer;
        font-size: 0.75rem;
        font-weight: 600;
        line-height: 1;
        opacity: 0.8;
        color: var(--color-ink-soft);
        transition: all 0.2s ease;
        padding: 0.3rem 0.55rem;
        letter-spacing: 0.08em;
    }
    button:hover {
        color: var(--color-primary);
        background: var(--color-primary-100);
    }

    .current {
        color: var(--color-primary);
    }

    .current:hover {
        opacity: 1;
    }

    .dropdown {
        list-style: none;
        margin: 0;
        animation: fadeIn 0.15s ease;
        display: flex;
    }

    @keyframes fadeIn {
        from {
            opacity: 0;
            transform: translateY(-4px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }

    @media (max-width: 768px) {
        button {
            padding: 0.2rem 0.4rem;
            font-size: 0.7rem;
        }
    }
</style>
