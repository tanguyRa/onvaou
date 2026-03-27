<script lang="ts">
    import type { LayoutProps } from "./$types";
    import { useSession } from "$lib/auth-client";
    import LanguageSwitcher from "$lib/components/LanguageSwitcher.svelte";
    import { goto } from "$app/navigation";

    let { children }: LayoutProps = $props();

    const session = useSession();

    $effect(() => {
        if ($session.data) {
            goto("/app");
        }
    });
</script>

{#if !$session.data}
    <div class="auth-container">
        <article class="card auth">
            <div class="auth-language">
                <LanguageSwitcher />
            </div>
            {@render children()}
        </article>
    </div>
{/if}

<style>
    .auth-container {
        min-height: 100vh;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: var(--space-6) var(--space-4);
        background: radial-gradient(
            circle at top,
            #fff3ed 0%,
            #f8f7f4 55%,
            #ffffff 100%
        );
    }
    .auth-language {
        align-self: flex-end;
    }

    .auth {
        width: min(420px, 100%);
        gap: 0;
    }
</style>
