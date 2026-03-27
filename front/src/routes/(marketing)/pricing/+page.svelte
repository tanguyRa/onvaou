<script lang="ts">
    import { useSession } from "$lib/auth-client";
    import { goto } from "$app/navigation";
    import { onMount } from "svelte";
    import LanguageSwitcher from "$lib/components/LanguageSwitcher.svelte";
    import { t } from "$lib/i18n/index.svelte";

    interface ProductPrice {
        priceAmount: number | null;
        recurringInterval: string | null;
    }

    interface ProductBenefit {
        id: string;
        description: string;
    }

    interface Product {
        slug: string;
        name: string;
        description: string | null;
        prices: ProductPrice[];
        benefits: ProductBenefit[];
        isHighlighted?: boolean;
    }

    const session = useSession();

    let products = $state<Product[]>([]);
    let loading = $state(true);
    let checkoutLoading = $state<string | null>(null);

    const demoPlans: Product[] = [
        {
            slug: "starter",
            name: "Starter",
            description: "Example free tier",
            prices: [{ priceAmount: 0, recurringInterval: null }],
            benefits: [
                { id: "starter-1", description: "Authentication flow example" },
                { id: "starter-2", description: "Localization setup example" }
            ]
        },
        {
            slug: "pro",
            name: "Pro",
            description: "Example recurring tier",
            prices: [{ priceAmount: 1900, recurringInterval: "month" }],
            benefits: [
                { id: "pro-1", description: "Polar checkout integration" },
                { id: "pro-2", description: "Customer portal ready" }
            ],
            isHighlighted: true
        }
    ];

    onMount(async () => {
        try {
            const response = await fetch("/api/payments/products");
            const data = await response.json();
            if (data.products?.length) {
                products = data.products;
            }
        } catch {
            // Keep demo plans when Polar is not configured.
        } finally {
            loading = false;
        }
    });

    function formatPrice(product: Product): string {
        const price = product.prices[0];
        if (!price || price.priceAmount === null || price.priceAmount === 0) {
            return t("pricing.free");
        }
        return `$${price.priceAmount / 100}`;
    }

    function billingCycle(product: Product): string {
        const interval = product.prices[0]?.recurringInterval;
        if (interval === "month") return t("pricing.monthly");
        if (interval === "year") return t("pricing.yearly");
        return "";
    }

    async function handleCheckout(slug: string) {
        if (!$session.data?.user) {
            goto("/login");
            return;
        }

        checkoutLoading = slug;
        try {
            const response = await fetch("/api/payments/checkout", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ slug })
            });

            if (!response.ok) {
                console.error("Checkout failed", await response.text());
                return;
            }

            const data = await response.json();
            if (data.url) {
                window.location.href = data.url;
            }
        } finally {
            checkoutLoading = null;
        }
    }
</script>

<div class="pricing-page">
    <header>
        <a href="/">OnVaOu</a>
        <LanguageSwitcher />
    </header>

    <main>
        <h1>{t("pricing.title")}</h1>
        <p>{t("pricing.subtitle")}</p>

        {#if loading}
            <p class="muted">{t("pricing.loading")}</p>
        {:else}
            {@const plans = products.length ? products : demoPlans}
            <div class="plans">
                {#each plans as product}
                    <article
                        class="plan card"
                        class:highlighted={product.isHighlighted}
                    >
                        <h2>{product.name}</h2>
                        <p class="muted">{product.description}</p>

                        <div class="price-row">
                            <span class="price">{formatPrice(product)}</span>
                            <span class="muted">{billingCycle(product)}</span>
                        </div>

                        <ul>
                            {#each product.benefits as benefit}
                                <li>{benefit.description}</li>
                            {/each}
                        </ul>

                        <button
                            class="btn"
                            class:btn-primary={product.isHighlighted}
                            class:btn-secondary={!product.isHighlighted}
                            disabled={checkoutLoading !== null}
                            onclick={() => handleCheckout(product.slug)}
                        >
                            {checkoutLoading === product.slug
                                ? t("pricing.processing")
                                : t("pricing.cta")}
                        </button>
                    </article>
                {/each}
            </div>
        {/if}
    </main>
</div>

<style>
    .pricing-page {
        min-height: 100vh;
        background: var(--color-bg);
    }

    header {
        max-width: var(--container-lg);
        margin: 0 auto;
        padding: var(--space-5) var(--space-6);
        font-weight: 700;
        display: flex;
        align-items: center;
        justify-content: space-between;
    }

    main {
        max-width: var(--container-lg);
        margin: 0 auto;
        padding: var(--space-6);
    }

    h1 {
        margin-bottom: var(--space-2);
    }

    .plans {
        margin-top: var(--space-6);
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
        gap: var(--space-4);
    }

    .plan {
        padding: var(--space-5);
        display: grid;
        gap: var(--space-3);
    }

    .plan.highlighted {
        border-color: var(--color-primary);
    }

    .price-row {
        display: flex;
        align-items: baseline;
        gap: var(--space-2);
    }

    .price {
        font-size: 1.75rem;
        font-weight: 700;
    }

    ul {
        margin: 0;
        padding-left: 1.1rem;
        color: var(--color-text-muted);
        display: grid;
        gap: var(--space-2);
    }
</style>
