<script lang="ts">
    import type { CitySuggestion } from "$lib/types/geo";

    interface Props {
        value: string;
        suggestions?: CitySuggestion[];
        loading?: boolean;
        onInput?: (value: string) => void;
        onSelect?: (suggestion: CitySuggestion) => void;
        onSubmit?: () => void;
    }

    let {
        value,
        suggestions = [],
        loading = false,
        onInput = () => {},
        onSelect = () => {},
        onSubmit = () => {},
    }: Props = $props();

    function handleSubmit(event: SubmitEvent) {
        event.preventDefault();
        onSubmit();
    }
</script>

<form class="search-shell" onsubmit={handleSubmit}>
    <label class="search-label" for="city-search">Search city or ZIP code</label
    >
    <div class="search-control">
        <input
            id="city-search"
            class="search-input"
            type="text"
            autocomplete="off"
            placeholder="Lyon, Bordeaux, 33000..."
            {value}
            oninput={(event) =>
                onInput((event.currentTarget as HTMLInputElement).value)}
        />
        <button class="search-button" type="submit">Search</button>
    </div>

    {#if loading || suggestions.length > 0}
        <div class="suggestions">
            {#if loading}
                <div class="suggestion-state">Looking up municipalities...</div>
            {:else}
                {#each suggestions as suggestion (suggestion.name)}
                    <button
                        class="suggestion"
                        type="button"
                        onclick={() => onSelect(suggestion)}
                    >
                        <span>{suggestion.city}</span>
                        <small>{suggestion.postcode}</small>
                    </button>
                {/each}
            {/if}
        </div>
    {/if}
</form>

<style>
    .search-shell {
        position: relative;
        min-width: min(26rem, 100%);
        flex: 1;
    }

    .search-label {
        display: block;
        margin-bottom: var(--spacing-xs);
        font-size: var(--font-size-xs);
        font-weight: 700;
        letter-spacing: 0.08em;
        text-transform: uppercase;
        color: var(--color-text-muted);
    }

    .search-control {
        display: flex;
        gap: var(--spacing-sm);
        align-items: center;
    }

    .search-input {
        flex: 1;
        min-width: 0;
        padding: 0.95rem 1rem;
        border: 1px solid rgba(17, 24, 39, 0.08);
        border-radius: 1.25rem;
        background: rgba(255, 255, 255, 0.86);
        box-shadow: var(--shadow-sm);
    }

    .search-button {
        padding-inline: 1.25rem;
        white-space: nowrap;
        background: var(--color-text);
        color: white;
        box-shadow: var(--shadow-sm);
    }

    .suggestions {
        position: absolute;
        z-index: 1200;
        top: calc(100% + 0.5rem);
        left: 0;
        right: 0;
        display: grid;
        gap: 0.35rem;
        padding: 0.5rem;
        border: 1px solid rgba(17, 24, 39, 0.08);
        border-radius: 1.25rem;
        background: rgba(255, 255, 255, 0.97);
        box-shadow: var(--shadow-lg);
        backdrop-filter: blur(12px);
    }

    .suggestion {
        justify-content: space-between;
        width: 100%;
        padding: 0.85rem 0.95rem;
        border-radius: 0.95rem;
        background: transparent;
        color: var(--color-text);
    }

    .suggestion:hover {
        background: rgba(255, 107, 53, 0.08);
        transform: none;
    }

    .suggestion small,
    .suggestion-state {
        color: var(--color-text-muted);
    }

    @media (max-width: 720px) {
        .search-control {
            flex-direction: column;
            align-items: stretch;
        }
    }
</style>
