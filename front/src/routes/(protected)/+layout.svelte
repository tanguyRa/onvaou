<script lang="ts">
    import { onMount } from "svelte";
    import type { LayoutProps } from "./$types";
    import { signOut } from "$lib/auth-client";
    import { t } from "$lib/i18n/index.svelte";
    import { useUser, resetUserStore } from "$lib/stores/user.svelte";
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import Spinner from "$lib/components/Spinner.svelte";

    let { children }: LayoutProps = $props();

    const user = useUser();

    let sidebarCollapsed = $state(false);
    let mobileNavOpen = $state(false);
    let isMobileViewport = $state(false);

    $effect(() => {
        if (!user.state.isPending && !user.state.isAuthenticated) {
            goto("/login");
        }
    });

    $effect(() => {
        if (typeof window === "undefined") {
            return;
        }

        const mediaQuery = window.matchMedia("(max-width: 900px)");
        const handleChange = () => {
            isMobileViewport = mediaQuery.matches;
            if (!mediaQuery.matches) {
                mobileNavOpen = false;
            }
        };

        handleChange();
        mediaQuery.addEventListener("change", handleChange);

        return () => {
            mediaQuery.removeEventListener("change", handleChange);
        };
    });

    $effect(() => {
        if (typeof window !== "undefined") {
            const stored = localStorage.getItem("sidebar-collapsed");
            if (stored !== null) {
                sidebarCollapsed = stored === "true";
            }
        }
    });

    $effect(() => {
        $page.url.pathname;
        mobileNavOpen = false;
    });

    function toggleSidebar() {
        sidebarCollapsed = !sidebarCollapsed;
        if (typeof window !== "undefined") {
            localStorage.setItem("sidebar-collapsed", String(sidebarCollapsed));
        }
    }

    function toggleMobileNav() {
        mobileNavOpen = !mobileNavOpen;
    }

    async function handleLogout(event: MouseEvent) {
        event.preventDefault();

        try {
            await signOut();
        } catch (e) {
            console.error("Logout error:", e);
        } finally {
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
    <div class="loading-container">
        <div class="spinner spinner-dark"></div>
    </div>
{:else}
    <div
        class="app-layout"
        class:collapsed={sidebarCollapsed && !isMobileViewport}
        class:mobile-nav-open={mobileNavOpen}
    >
        {#if isMobileViewport}
            <button
                class="sidebar-backdrop"
                class:visible={mobileNavOpen}
                onclick={() => (mobileNavOpen = false)}
                aria-label="Close sidebar"
            ></button>
        {/if}

        <aside
            id="protected-sidebar"
            class="sidebar"
            class:open={mobileNavOpen}
        >
            <div class="sidebar-header">
                <a href="/app" class="logo" aria-label="SaaS Seed home">
                    <span class="logo-mark"></span>
                    {#if !sidebarCollapsed || isMobileViewport}
                        <span class="logo-text">OnVaOu</span>
                    {/if}
                </a>

                {#if isMobileViewport}
                    <button
                        class="toggle-btn"
                        onclick={() => (mobileNavOpen = false)}
                        aria-label="Close menu"
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
                            <line x1="18" y1="6" x2="6" y2="18"></line>
                            <line x1="6" y1="6" x2="18" y2="18"></line>
                        </svg>
                    </button>
                {:else}
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
                {/if}
            </div>

            <nav class="sidebar-nav">
                <div class="nav-section">
                    <a
                        href="/app"
                        class="nav-link"
                        aria-current={isCurrentPath("/app")
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
                            <rect x="3" y="3" width="7" height="7"></rect>
                            <rect x="14" y="3" width="7" height="7"></rect>
                            <rect x="14" y="14" width="7" height="7"></rect>
                            <rect x="3" y="14" width="7" height="7"></rect>
                        </svg>
                        {#if !sidebarCollapsed || isMobileViewport}
                            <span>Dashboard</span>
                        {/if}
                    </a>

                    <a
                        href="/app/map"
                        class="nav-link"
                        aria-current={isCurrentPath("/app/map")
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
                            <polygon points="3 6 9 3 15 6 21 3 21 18 15 21 9 18 3 21"></polygon>
                            <line x1="9" y1="3" x2="9" y2="18"></line>
                            <line x1="15" y1="6" x2="15" y2="21"></line>
                        </svg>
                        {#if !sidebarCollapsed || isMobileViewport}
                            <span>Map</span>
                        {/if}
                    </a>
                </div>

                <div class="nav-section nav-section-bottom">
                    <span class="nav-section-label"
                        >{sidebarCollapsed && !isMobileViewport
                            ? ""
                            : t("protected.sidebar.settings")}</span
                    >

                    <a
                        href="/settings/rules"
                        class="nav-link"
                        aria-current={isCurrentPath("/settings/rules")
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
                            <path d="M12 3l1.912 5.813L20 10.5l-4.95 3.6L16.962 20 12 16.9 7.038 20l1.912-5.9L4 10.5l6.088-1.687z"></path>
                        </svg>
                        {#if !sidebarCollapsed || isMobileViewport}
                            <span>Auto-tag rules</span>
                        {/if}
                    </a>

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
                        {#if !sidebarCollapsed || isMobileViewport}
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
                        {#if !sidebarCollapsed || isMobileViewport}
                            <span>{t("protected.sidebar.billing")}</span>
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
                        {#if !sidebarCollapsed || isMobileViewport}
                            <span>{t("protected.sidebar.logout")}</span>
                        {/if}
                    </a>
                </div>
            </nav>
        </aside>

        <main class="main-content">
            <div class="mobile-topbar">
                <button
                    class="mobile-menu-btn"
                    onclick={toggleMobileNav}
                    aria-expanded={mobileNavOpen}
                    aria-controls="protected-sidebar"
                    aria-label="Open menu"
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
                        <line x1="3" y1="12" x2="21" y2="12"></line>
                        <line x1="3" y1="6" x2="21" y2="6"></line>
                        <line x1="3" y1="18" x2="21" y2="18"></line>
                    </svg>
                </button>
                <a href="/app" class="mobile-title">OnVaOu</a>
            </div>

            <div class="main-shell" class:map-shell={isCurrentPath("/app/map")}>
                {@render children()}
            </div>
        </main>
    </div>
{/if}

<style>
    .loading-container {
        min-height: 100vh;
        display: flex;
        align-items: center;
        justify-content: center;
        background: var(--color-surface-alt);
    }

    .app-layout {
        display: flex;
        height: 100dvh;
        background: var(--color-surface);
        position: relative;
        overflow: hidden;
    }

    .sidebar {
        width: 248px;
        height: 100dvh;
        background: linear-gradient(180deg, var(--color-bg) 0%, #f3f2ef 100%);
        border-right: 1px solid var(--color-border);
        display: flex;
        flex-direction: column;
        transition:
            width var(--duration-base) var(--ease-standard),
            transform var(--duration-base) var(--ease-standard);
        z-index: 20;
        flex-shrink: 0;
    }

    .app-layout.collapsed .sidebar {
        width: 76px;
    }

    .sidebar-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: var(--space-3);
        padding: var(--space-4);
        border-bottom: 1px solid var(--color-border);
    }

    .logo {
        display: flex;
        align-items: center;
        gap: var(--space-3);
        color: var(--color-text);
        overflow: hidden;
        min-width: 0;
    }

    .logo-mark {
        width: 26px;
        height: 26px;
        background: var(--color-text);
        clip-path: polygon(50% 0, 100% 50%, 50% 100%, 0 50%);
        position: relative;
        flex-shrink: 0;
    }

    .logo-mark::after {
        content: "";
        position: absolute;
        inset: 6px;
        background: var(--color-primary);
        clip-path: polygon(50% 0, 100% 50%, 50% 100%, 0 50%);
    }

    .logo-text {
        font-size: var(--text-lg);
        font-weight: 700;
        letter-spacing: 0.02em;
        font-family: var(--font-display);
        white-space: nowrap;
    }

    .toggle-btn,
    .mobile-menu-btn {
        background: transparent;
        border: 1px solid transparent;
        padding: var(--space-2);
        border-radius: var(--radius-md);
        cursor: pointer;
        color: var(--color-text-muted);
        transition:
            background var(--duration-fast) var(--ease-standard),
            color var(--duration-fast) var(--ease-standard),
            border-color var(--duration-fast) var(--ease-standard);
        display: inline-flex;
        align-items: center;
        justify-content: center;
    }

    .toggle-btn:hover,
    .mobile-menu-btn:hover {
        background: var(--color-surface);
        border-color: var(--color-border);
        color: var(--color-text);
    }

    .app-layout.collapsed .sidebar-header {
        flex-direction: column;
        gap: var(--space-2);
    }

    .sidebar-nav {
        flex: 1;
        display: flex;
        flex-direction: column;
        padding: var(--space-3);
        overflow-y: auto;
    }

    .nav-section {
        display: flex;
        flex-direction: column;
        gap: var(--space-2);
    }

    .nav-subsection,
    .quick-actions {
        display: flex;
        flex-direction: column;
        gap: var(--space-2);
        margin-top: var(--space-2);
        padding-left: var(--space-4);
    }

    .nav-section-bottom {
        margin-top: auto;
        border-top: 1px solid var(--color-border);
        padding-top: var(--space-4);
    }

    .nav-section-label {
        font-size: var(--text-sm);
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        color: var(--color-text-muted);
        padding: 0 var(--space-2);
        min-height: 20px;
    }

    .nav-link {
        display: flex;
        align-items: center;
        gap: var(--space-3);
        padding: var(--space-2) var(--space-3);
        border-radius: var(--radius-md);
        color: var(--color-text-muted);
        font-weight: 600;
        font-size: var(--text-md);
        transition:
            background var(--duration-fast) var(--ease-standard),
            color var(--duration-fast) var(--ease-standard),
        box-shadow var(--duration-fast) var(--ease-standard);
    }

    .nav-sublink,
    .quick-action-btn {
        display: flex;
        align-items: center;
        gap: var(--space-2);
        justify-content: space-between;
        padding: 0.5rem 0.75rem;
        border-radius: var(--radius-md);
        color: var(--color-text-muted);
        font-size: var(--text-sm);
        font-weight: 600;
        border: 1px solid transparent;
        background: transparent;
        text-align: left;
    }

    .quick-action-btn {
        cursor: pointer;
        color: var(--color-text);
        background: var(--color-surface);
        border-color: var(--color-border);
    }

    .quick-action-btn:disabled {
        opacity: 0.7;
        cursor: wait;
    }

    .nav-sublink[aria-current="page"] {
        color: var(--color-text);
        background: var(--color-surface);
        border-color: var(--color-border);
    }

    .nav-sublink:hover,
    .quick-action-btn:hover {
        background: var(--color-surface);
        color: var(--color-text);
    }

    .nav-sublink-tag {
        justify-content: flex-start;
    }

    .tag-swatch {
        width: 0.6rem;
        height: 0.6rem;
        border-radius: 999px;
        background: var(--tag-color);
        flex: 0 0 auto;
    }

    .nav-count {
        padding: 0.15rem 0.5rem;
        font-size: 0.72rem;
    }

    .nav-link[aria-current="page"] {
        background: var(--color-primary-soft);
        color: var(--color-text);
        box-shadow: inset 3px 0 0 var(--color-primary);
    }

    .nav-link:hover,
    .nav-link[aria-current="page"]:hover {
        background: var(--color-surface);
        color: var(--color-text);
    }

    .nav-link svg {
        flex-shrink: 0;
    }

    .nav-link span {
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    .app-layout.collapsed .nav-link {
        justify-content: center;
        padding: var(--space-2);
    }

    .logout-btn,
    .logout-btn:hover {
        color: var(--color-danger);
    }

    .logout-btn:hover {
        background: var(--color-danger-bg);
    }

    .sidebar-backdrop {
        display: none;
    }

    .main-content {
        flex: 1;
        min-width: 0;
        height: 100dvh;
        background: var(--color-surface);
        overflow: hidden;
    }

    .main-shell {
        height: 100%;
        overflow-y: auto;
        padding: var(--space-4);
    }

    .main-shell.map-shell {
        padding: 0;
        overflow: hidden;
    }

    .mobile-topbar {
        display: none;
    }

    @media (max-width: 900px) {
        .app-layout {
            height: 100dvh;
        }

        .sidebar {
            position: fixed;
            left: 0;
            top: 0;
            bottom: 0;
            width: min(86vw, 290px);
            transform: translateX(-100%);
            box-shadow: var(--shadow-lg);
        }

        .sidebar.open {
            transform: translateX(0);
        }

        .sidebar-backdrop {
            display: block;
            position: fixed;
            inset: 0;
            border: 0;
            background: rgba(17, 24, 39, 0.42);
            opacity: 0;
            pointer-events: none;
            transition: opacity var(--duration-base) var(--ease-standard);
            z-index: 10;
        }

        .sidebar-backdrop.visible {
            opacity: 1;
            pointer-events: auto;
        }

        .main-content {
            height: 100dvh;
        }

        .main-shell {
            height: calc(100% - 61px);
            padding: var(--space-2);
        }

        .main-shell.map-shell {
            padding: 0;
        }

        .mobile-topbar {
            display: flex;
            align-items: center;
            justify-content: space-between;
            gap: var(--space-3);
            position: sticky;
            top: 0;
            z-index: 5;
            padding: var(--space-3) var(--space-4);
            background: color-mix(
                in srgb,
                var(--color-surface) 92%,
                transparent
            );
            backdrop-filter: blur(6px);
            border-bottom: 1px solid var(--color-border);
        }

        .mobile-title {
            font-family: var(--font-display);
            font-size: var(--text-lg);
            font-weight: 700;
            color: var(--color-text);
        }
    }
</style>
