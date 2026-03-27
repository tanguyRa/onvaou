<script lang="ts">
    import type { LayoutProps } from "./$types";
    import { signOut } from "$lib/auth-client";
    import { t } from "$lib/i18n/index.svelte";
    import { useUser, resetUserStore } from "$lib/stores/user.svelte";
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";

    let { children }: LayoutProps = $props();

    const user = useUser();

    let sidebarCollapsed = $state(false);

    // Redirect to login if not authenticated
    $effect(() => {
        if (!user.state.isPending && !user.state.isAuthenticated) {
            goto("/login");
        }
    });

    // Load collapsed state from localStorage on mount
    $effect(() => {
        if (typeof window !== "undefined") {
            const stored = localStorage.getItem("sidebar-collapsed");
            if (stored !== null) {
                sidebarCollapsed = stored === "true";
            }
        }
    });

    // Persist collapsed state to localStorage
    function toggleSidebar() {
        sidebarCollapsed = !sidebarCollapsed;
        if (typeof window !== "undefined") {
            localStorage.setItem("sidebar-collapsed", String(sidebarCollapsed));
        }
    }

    async function handleLogout() {
        try {
            await signOut();
        } catch (e) {
            console.error("Logout error:", e);
        } finally {
            // Always reset store and redirect, even if signOut fails
            resetUserStore();
            goto("/");
        }
    }

    function isCurrentPath(path: string) {
        const pathname = $page.url.pathname;
        return pathname === path || pathname.startsWith(`${path}/`);
    }
</script>

{#if user.state.isPending || !user.state.isAuthenticated}
    <!-- Show loading while checking auth or redirecting -->
    <div class="loading-container">
        <div class="spinner spinner-dark"></div>
    </div>
{:else}
    <div class="app-layout" class:collapsed={sidebarCollapsed}>
        <aside class="sidebar">
            <div class="sidebar-header">
                <a href="/app" class="logo" aria-label="OnVaOu home">
                    <span class="logo-mark"></span>
                    {#if !sidebarCollapsed}
                        <span class="logo-text">OnVaOu</span>
                    {/if}
                </a>
                <button
                    class="toggle-btn"
                    onclick={toggleSidebar}
                    aria-label={t("protected.sidebar.toggle")}
                >
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="20"
                        height="20"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                    >
                        {#if sidebarCollapsed}
                            <polyline points="9 18 15 12 9 6"></polyline>
                        {:else}
                            <polyline points="15 18 9 12 15 6"></polyline>
                        {/if}
                    </svg>
                </button>
            </div>

            <nav class="sidebar-nav">
                <div class="nav-section"></div>

                <div class="nav-section nav-section-bottom">
                    <span class="nav-section-label"
                        >{sidebarCollapsed
                            ? ""
                            : t("protected.sidebar.settings")}</span
                    >
                    <a
                        href="/settings/profile"
                        class="nav-link"
                        aria-current={isCurrentPath("/settings/profile")
                            ? "page"
                            : undefined}
                    >
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            width="20"
                            height="20"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                        >
                            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"
                            ></path>
                            <circle cx="12" cy="7" r="4"></circle>
                        </svg>
                        {#if !sidebarCollapsed}
                            <span>{t("protected.sidebar.profile")}</span>
                        {/if}
                    </a>
                    <a
                        href="/settings/billing"
                        class="nav-link"
                        aria-current={isCurrentPath("/settings/billing")
                            ? "page"
                            : undefined}
                    >
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            width="20"
                            height="20"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                        >
                            <rect
                                x="1"
                                y="4"
                                width="22"
                                height="16"
                                rx="2"
                                ry="2"
                            ></rect>
                            <line x1="1" y1="10" x2="23" y2="10"></line>
                        </svg>
                        {#if !sidebarCollapsed}
                            <span>{t("protected.sidebar.billing")}</span>
                        {/if}
                    </a>
                    <a
                        href="/settings/api-keys"
                        class="nav-link"
                        aria-current={isCurrentPath("/settings/api-keys")
                            ? "page"
                            : undefined}
                    >
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            width="20"
                            height="20"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                        >
                            <path
                                d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"
                            ></path>
                        </svg>
                        {#if !sidebarCollapsed}
                            <span>{t("protected.sidebar.apiKeys")}</span>
                        {/if}
                    </a>
                    <a
                        href="/"
                        class="nav-link logout-btn"
                        onclick={handleLogout}
                        aria-label={t("protected.sidebar.logout")}
                    >
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            width="20"
                            height="20"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                        >
                            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"
                            ></path>
                            <polyline points="16 17 21 12 16 7"></polyline>
                            <line x1="21" y1="12" x2="9" y2="12"></line>
                        </svg>
                        {#if !sidebarCollapsed}
                            <span>{t("protected.sidebar.logout")}</span>
                        {/if}
                    </a>
                </div>
            </nav>
        </aside>

        <main class="main-content">
            {@render children()}
        </main>
    </div>
{/if}

<style>
    /* Loading State */
    .loading-container {
        min-height: 100vh;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--color-surface-alt);
    }

    /* App Layout */
    .app-layout {
        display: flex;
        height: 100vh;
    }

    /* Sidebar */
    .sidebar {
        width: 240px;
        height: 100%;
        background: var(--color-bg);
        border-right: 1px solid var(--color-border);
        display: flex;
        flex-direction: column;
        transition: width 0.3s ease;
    }

    .app-layout.collapsed .sidebar {
        width: 72px;
    }

    .sidebar-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: var(--lg);
        border-bottom: 1px solid var(--color-border);
    }

    .sidebar-header .logo {
        display: flex;
        align-items: center;
        gap: 0.75rem;
        text-decoration: none;
        color: var(--color-text);
        overflow: hidden;
    }

    .sidebar-header .logo-mark {
        width: 26px;
        height: 26px;
        background: var(--color-text);
        clip-path: polygon(50% 0, 100% 50%, 50% 100%, 0 50%);
        position: relative;
        flex-shrink: 0;
    }

    .sidebar-header .logo-mark::after {
        content: "";
        position: absolute;
        inset: 6px;
        background: var(--color-primary);
        clip-path: polygon(50% 0, 100% 50%, 50% 100%, 0 50%);
    }

    .sidebar-header .logo-text {
        font-size: 1.05rem;
        font-weight: 700;
        letter-spacing: 0.02em;
        font-family: "Canela", "Iowan Old Style", "Palatino Linotype",
            "Book Antiqua", Palatino, serif;
        white-space: nowrap;
    }

    .toggle-btn {
        background: none;
        border: none;
        padding: var(--spacing-sm);
        border-radius: var(--radius-sm);
        cursor: pointer;
        color: var(--color-text-muted);
        transition: all 0.1s ease;
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
    }

    .toggle-btn:hover {
        background: var(--color-border);
        color: var(--color-text);
    }

    .app-layout.collapsed .toggle-btn {
        margin-left: auto;
        margin-right: auto;
    }

    .app-layout.collapsed .sidebar-header {
        flex-direction: column;
        gap: var(--sm);
    }

    /* Sidebar Navigation */
    .sidebar-nav {
        flex: 1;
        display: flex;
        flex-direction: column;
        padding: var(--md);
        overflow-y: auto;
    }

    .nav-section {
        display: flex;
        flex-direction: column;
        gap: var(--xs);
    }

    .nav-section-bottom {
        margin-top: auto;
        border-top: 1px solid var(--color-border);
        padding-top: var(--md);
    }

    .nav-section-label {
        font-size: var(--xs);
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 0.5px;
        color: var(--color-text-muted);
        padding: var(--sm) var(--sm);
        min-height: 24px;
    }

    .nav-link {
        display: flex;
        align-items: center;
        gap: var(--xs);
        /* padding: var(--xs) var(--sm); */
        border-radius: var(--md);
        color: var(--color-text-soft);
        text-decoration: none;
        font-weight: 500;
        transition: all 0.1s ease;
        background: none;
        border: none;
        width: 100%;
        cursor: pointer;
        font-size: var(--font-size-base);
        font-family: var(--font-family);
    }

    .nav-link[aria-current="page"] {
        background: var(--color-primary-100);
        color: var(--color-text);
        box-shadow: inset 3px 0 0 var(--color-primary);
    }

    .nav-link:hover,
    .nav-link[aria-current="page"]:hover {
        background: var(--color-border-light);
        color: var(--color-text);
    }

    .nav-link svg {
        flex-shrink: 0;
    }

    .nav-link span {
        white-space: nowrap;
        overflow: hidden;
    }

    .app-layout.collapsed .nav-link {
        justify-content: center;
        padding: var(--spacing-sm);
    }

    .logout-btn {
        color: var(--color-error);
    }

    .logout-btn:hover {
        background: var(--color-error-bg);
        color: var(--color-error);
    }

    /* Main Content */
    .main-content {
        flex: 1;
        background: var(--color-surface);
        height: 100vh;
        transition: margin-left var(--transition-slow);
        overflow-y: auto;
        padding: var(--xxl);
    }

    /* Responsive */
    @media (max-width: 768px) {
        .sidebar {
            transform: translateX(-100%);
            width: 240px !important;
        }

        .main-content {
            margin-left: 0 !important;
        }
    }
</style>
