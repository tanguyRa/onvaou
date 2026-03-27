<script lang="ts">
    import { useUser } from "$lib/stores/user.svelte";
    import { onMount } from "svelte";

    interface ProductPrice {
        id: string;
        type: string;
        amountType: string;
        priceAmount: number | null;
        priceCurrency: string | null;
        recurringInterval: string | null;
    }

    interface ProductBenefit {
        id: string;
        description: string;
        type: string;
    }

    interface Product {
        id: string;
        slug: string;
        name: string;
        description: string | null;
        prices: ProductPrice[];
        benefits: ProductBenefit[];
        isRecurring: boolean;
        isHighlighted: boolean;
    }

    let products = $state<Product[]>([]);
    let paymentProvider = $state<"none" | "polar">("none");
    let loading = $state(true);
    let checkoutLoading = $state<string | null>(null);
    let portalLoading = $state(false);
    let error = $state("");

    const user = useUser();

    onMount(() => {
        const abortController = new AbortController();
        fetchProducts(abortController.signal);
        return () => abortController.abort();
    });

    async function fetchProducts(signal?: AbortSignal) {
        try {
            const response = await fetch("/api/payments/products", { signal });
            const data = await response.json();
            paymentProvider = data.provider ?? "none";
            if (data.products) {
                products = data.products;
            }
        } catch (e) {
            // Ignore abort errors
            if (e instanceof Error && e.name === "AbortError") return;
            error = "Failed to load plans";
        } finally {
            loading = false;
        }
    }

    function getCurrentProduct(): Product | null {
        const subscription = user.state.activeSubscription;
        if (!subscription) return null;
        return products.find((p) => p.id === subscription.productId) || null;
    }

    function formatPrice(product: Product): string {
        const price = product.prices[0];
        if (!price || price.priceAmount === null || price.priceAmount === 0) {
            return "Free";
        }
        return `$${price.priceAmount / 100}`;
    }

    function getBillingCycle(product: Product): string {
        const price = product.prices[0];
        if (!price || price.priceAmount === null || price.priceAmount === 0) {
            return "";
        }
        if (price.recurringInterval === "month") return "/month";
        if (price.recurringInterval === "year") return "/year";
        return "";
    }

    function formatDate(date: string | Date | null | undefined): string {
        if (!date) return "";
        const d = date instanceof Date ? date : new Date(date);
        return d.toLocaleDateString("en-US", {
            year: "numeric",
            month: "long",
            day: "numeric",
        });
    }

    function getStatusBadgeClass(status: string): string {
        switch (status) {
            case "active":
                return "success";
            case "past_due":
                return "warning";
            case "canceled":
                return "error";
            default:
                return "neutral";
        }
    }
</script>

<article>
    <header>
        <h1>Billing & Subscription</h1>
        <p>Manage your subscription and billing information</p>
    </header>

    {#if loading || user.state.subscriptionLoading}
        <div class="loading-state">
            <div class="spinner spinner-dark"></div>
        </div>
    {:else}
        {@const subscription = user.state.activeSubscription}
        {@const currentProduct = getCurrentProduct()}

        <!-- Current Plan Section -->
        <section>
            <header>
                <h2>Current Plan</h2>
                <p>Your active subscription</p>
            </header>

            {#if subscription && currentProduct}
                <div class="current-plan">
                    <div class="plan-info">
                        <div class="plan-name-row">
                            <h3>{currentProduct.name}</h3>
                            <span
                                class="badge {getStatusBadgeClass(
                                    subscription.status,
                                )}"
                            >
                                {subscription.status === "active"
                                    ? "Active"
                                    : subscription.status}
                            </span>
                        </div>
                        <p class="plan-price-info">
                            {formatPrice(currentProduct)}{getBillingCycle(
                                currentProduct,
                            )}
                        </p>
                        {#if subscription.currentPeriodEnd}
                            <p class="plan-renewal">
                                {#if subscription.cancelAtPeriodEnd}
                                    <span class="cancellation-notice">
                                        Cancels on {formatDate(
                                            subscription.currentPeriodEnd,
                                        )}
                                    </span>
                                {:else}
                                    Renews on {formatDate(
                                        subscription.currentPeriodEnd,
                                    )}
                                {/if}
                            </p>
                        {/if}
                    </div>
                </div>
            {:else}
                <div class="current-plan free-plan">
                    <div class="plan-info">
                        <div class="plan-name-row">
                            <h3>Free Plan</h3>
                            <span class="badge badge-neutral">Active</span>
                        </div>
                        <p class="plan-price-info">$0/month</p>
                        <p class="plan-description">
                            Basic features with limited access
                        </p>
                    </div>
                </div>
            {/if}
        </section>

        {#if subscription && currentProduct}
            <!-- Payment Method Section -->
            <section>
                <header>
                    <h2>Payment Method</h2>
                    <p>Manage your payment information</p>
                </header>

                <div class="payment-info">
                    {#if paymentProvider === "polar"}
                        <p class="payment-description">
                            Update your payment method, view invoices, or cancel
                            your subscription through the customer portal.
                        </p>
                        <a class="btn btn-secondary" href="/settings/portal">
                            {#if portalLoading}
                                <span class="spinner spinner-sm spinner-dark"
                                ></span>
                            {/if}
                            Open Customer Portal
                        </a>
                    {:else}
                        <p class="payment-description">
                            No payment provider configured.
                        </p>
                    {/if}
                </div>
            </section>
        {/if}
    {/if}
</article>

<style>
    .current-plan {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        gap: var(--spacing-lg);
    }

    .plan-info {
        flex: 1;
    }

    .plan-name-row {
        display: flex;
        align-items: center;
        gap: var(--spacing-md);
        margin-bottom: var(--spacing-sm);
    }

    .plan-name-row h3 {
        font-size: var(--font-size-xl);
        color: var(--color-text);
        margin: 0;
    }

    .plan-price-info {
        font-size: var(--font-size-lg);
        color: var(--color-text-secondary);
        margin-bottom: var(--spacing-xs);
    }

    .plan-renewal {
        font-size: var(--font-size-sm);
        color: var(--color-text-muted);
    }

    .plan-description {
        font-size: var(--font-size-sm);
        color: var(--color-text-muted);
    }

    .cancellation-notice {
        color: var(--color-warning);
    }

    .payment-info {
        display: flex;
        flex-direction: column;
        gap: var(--spacing-md);
    }

    .payment-description {
        color: var(--color-text-muted);
        font-size: var(--font-size-sm);
    }

    .payment-info .btn {
        align-self: flex-start;
    }

    @media (max-width: 768px) {
        .current-plan {
            flex-direction: column;
        }
    }
</style>
