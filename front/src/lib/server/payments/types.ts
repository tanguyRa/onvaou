export type PaymentProvider = "none" | "polar";

export interface PaymentProductPrice {
    id: string;
    type: string;
    amountType: string;
    priceAmount: number | null;
    priceCurrency: string | null;
    recurringInterval: string | null;
}

export interface PaymentProductBenefit {
    id: string;
    description: string;
    type: string;
}

export interface PaymentProduct {
    id: string;
    slug: string;
    name: string;
    description: string | null;
    prices: PaymentProductPrice[];
    benefits: PaymentProductBenefit[];
    isRecurring: boolean;
    isHighlighted: boolean;
}

export interface PaymentCheckoutRequest {
    slug: string;
    successUrl?: string;
    cancelUrl?: string;
    customerEmail?: string;
    externalCustomerId?: string;
}

export interface PaymentCheckoutResponse {
    url: string;
}

export interface PaymentClient {
    provider: PaymentProvider;
    listProducts(): Promise<PaymentProduct[]>;
    createCheckout(request: PaymentCheckoutRequest): Promise<PaymentCheckoutResponse>;
}
