import { Polar } from "@polar-sh/sdk";
import type {
    PaymentCheckoutRequest,
    PaymentCheckoutResponse,
    PaymentClient,
    PaymentProduct,
    PaymentProvider,
} from "./types";

function slugify(value: string): string {
    return value
        .toLowerCase()
        .trim()
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/^-+|-+$/g, "");
}

function providerFromEnv(): PaymentProvider {
    const provider = (process.env.PAYMENT_PROVIDER || "").trim().toLowerCase();
    if (provider === "polar") return "polar";
    return "none";
}

function sortProducts(products: PaymentProduct[]): PaymentProduct[] {
    return products.sort((a, b) => {
        const getMinPrice = (product: PaymentProduct) => {
            const fixed = product.prices.find((price) => price.amountType === "fixed" && price.priceAmount !== null);
            return fixed?.priceAmount ?? 0;
        };
        return getMinPrice(a) - getMinPrice(b);
    });
}

class EmptyPaymentClient implements PaymentClient {
    provider: PaymentProvider = "none";

    async listProducts(): Promise<PaymentProduct[]> {
        return [];
    }

    async createCheckout(_request: PaymentCheckoutRequest): Promise<PaymentCheckoutResponse> {
        throw new Error("No payment provider configured");
    }
}

class PolarPaymentClient implements PaymentClient {
    provider: PaymentProvider = "polar";
    private client: Polar;

    constructor() {
        this.client = new Polar({
            accessToken: process.env.POLAR_ACCESS_TOKEN,
            server: process.env.POLAR_SERVER === "production" ? "production" : "sandbox"
        });
    }

    private toSlug(product: { metadata?: Record<string, unknown> | null; name: string }): string {
        const metadataSlug = typeof product.metadata?.slug === "string" ? product.metadata.slug : null;
        return metadataSlug || slugify(product.name);
    }

    async listProducts(): Promise<PaymentProduct[]> {
        const response = await this.client.products.list({
            isArchived: false,
            limit: 100,
        });

        const products = response.result.items.map((product) => {
            const highlighted = product.metadata?.highlighted;
            return {
                id: product.id,
                slug: this.toSlug(product),
                name: product.name,
                description: product.description,
                prices: product.prices.map((price) => ({
                    // Polar SDK models vary between versions; normalize shape defensively.
                    ...(() => {
                        const priceRecord = price as Record<string, unknown>;
                        const priceType = typeof priceRecord.type === "string" ? priceRecord.type : null;
                        const recurringInterval = typeof priceRecord.recurringInterval === "string"
                            ? priceRecord.recurringInterval
                            : null;
                        return {
                    id: price.id,
                    type: priceType ?? (recurringInterval ? "recurring" : "one_time"),
                    amountType: price.amountType,
                    priceAmount: price.amountType === "fixed" ? price.priceAmount : null,
                    priceCurrency: price.amountType === "fixed" ? price.priceCurrency : null,
                            recurringInterval,
                        };
                    })(),
                })),
                benefits: product.benefits.map((benefit) => ({
                    id: benefit.id,
                    description: benefit.description,
                    type: benefit.type,
                })),
                isRecurring: product.isRecurring,
                isHighlighted: String(highlighted).toLowerCase() === "true",
            } satisfies PaymentProduct;
        });

        return sortProducts(products);
    }

    async createCheckout(request: PaymentCheckoutRequest): Promise<PaymentCheckoutResponse> {
        const products = await this.listProducts();
        const selected = products.find((product) => product.slug === request.slug);
        if (!selected) {
            throw new Error(`Unknown product slug: ${request.slug}`);
        }

        const checkout = await this.client.checkouts.create({
            products: [selected.id],
            successUrl: request.successUrl,
            returnUrl: request.cancelUrl,
            customerEmail: request.customerEmail || undefined,
            externalCustomerId: request.externalCustomerId || undefined,
        });

        return { url: checkout.url };
    }
}

let cachedClient: PaymentClient | null = null;
let cachedProvider: PaymentProvider | null = null;

export function getPaymentClient(): PaymentClient {
    const provider = providerFromEnv();
    if (cachedClient && cachedProvider === provider) {
        return cachedClient;
    }

    if (provider === "polar") {
        cachedClient = new PolarPaymentClient();
    } else {
        cachedClient = new EmptyPaymentClient();
    }
    cachedProvider = provider;
    return cachedClient;
}
