<script lang="ts">
    interface Props {
        value: number;
        onInput?: (value: number) => void;
    }

    let { value, onInput = () => {} }: Props = $props();
</script>

<div class="radius-shell">
    <div class="radius-copy">
        <span class="radius-label">Radius</span>
        <strong>{value} km</strong>
    </div>
    <input
        class="radius-slider"
        type="range"
        min="1"
        max="50"
        value={value}
        oninput={(event) =>
            onInput(Number((event.currentTarget as HTMLInputElement).value))}
    />
</div>

<style>
    .radius-shell {
        min-width: 14rem;
        padding: 0.95rem 1rem;
        border: 1px solid rgba(17, 24, 39, 0.08);
        border-radius: 1.25rem;
        background: rgba(255, 255, 255, 0.86);
        box-shadow: var(--shadow-sm);
    }

    .radius-copy {
        display: flex;
        justify-content: space-between;
        align-items: baseline;
        margin-bottom: 0.55rem;
    }

    .radius-label {
        font-size: var(--font-size-xs);
        font-weight: 700;
        letter-spacing: 0.08em;
        text-transform: uppercase;
        color: var(--color-text-muted);
    }

    .radius-slider {
        width: 100%;
        padding: 0;
        accent-color: var(--color-primary);
        box-shadow: none;
        background: transparent;
    }
</style>
