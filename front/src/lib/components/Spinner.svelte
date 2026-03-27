<script lang="ts">
    interface Props {
        size?: 'tiny' | 'small' | 'medium' | 'large';
    }

    let { size = 'medium' }: Props = $props();

    const sizes = {
        tiny: 12,
        small: 16,
        medium: 24,
        large: 32,
    };
</script>

<span
    class="spinner"
    style="width: {sizes[size]}px; height: {sizes[size]}px;"
></span>

<style>
    .spinner {
        display: inline-block;
        border: 2px solid var(--color-border-light);
        border-top-color: currentColor;
        border-radius: 50%;
        animation: spin 0.8s linear infinite;
    }

    @keyframes spin {
        to {
            transform: rotate(360deg);
        }
    }
</style>
