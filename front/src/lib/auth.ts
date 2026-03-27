import { Pool } from "pg";

import { sveltekitCookies } from "better-auth/svelte-kit";
import { betterAuth } from "better-auth";
import { jwt } from "better-auth/plugins";
import { polar, checkout, portal, usage } from "@polar-sh/better-auth";
import { Polar } from "@polar-sh/sdk";

import { getRequestEvent } from "$app/server";


const rawDatabaseUrl = process.env.DATABASE_URL || "";
let sanitizedDatabaseUrl = rawDatabaseUrl;
try {
    const parsed = new URL(rawDatabaseUrl);
    parsed.searchParams.delete("channel_binding");
    sanitizedDatabaseUrl = parsed.toString();
} catch {
    sanitizedDatabaseUrl = rawDatabaseUrl.replace(/([?&])channel_binding=require(&|$)/, "$1");
    sanitizedDatabaseUrl = sanitizedDatabaseUrl.replace(/[?&]$/, "");
}

const pool = new Pool({
    connectionString: sanitizedDatabaseUrl || rawDatabaseUrl,
    // Pool configuration
    max: 20, // Maximum 20 concurrent connections
    idleTimeoutMillis: 30000, // Close idle connections after 30s
    maxLifetimeSeconds: 3600, // Max connection lifetime 1 hour
    // connectionTimeout: 10, // Connection timeout 10s
});

const paymentProvider = (process.env.PAYMENT_PROVIDER || "").trim().toLowerCase();
const usePolar = paymentProvider === "polar";

export const polarClient = usePolar
    ? new Polar({
        accessToken: process.env.POLAR_ACCESS_TOKEN,
        server: process.env.POLAR_SERVER === "production" ? "production" : "sandbox"
    })
    : null;
const polarPlugin = usePolar && polarClient
    ? polar({
        client: polarClient,
        createCustomerOnSignUp: true,
        use: [
            checkout({
                products: [
                    {
                        productId: "e54c3dec-3fa6-4a6d-b359-35fafdfe4b30",
                        slug: "Premium-Annual" // Custom slug for easy reference in Checkout URL, e.g. /checkout/Premium-Annual
                    },
                    {
                        productId: "a741f0a8-929d-4420-8329-2e880fa2ecf8",
                        slug: "Premium" // Custom slug for easy reference in Checkout URL, e.g. /checkout/Premium
                    },
                    {
                        productId: "015ddd64-2330-4fc7-a59d-c8cfcd9751ed",
                        slug: "Free" // Custom slug for easy reference in Checkout URL, e.g. /checkout/Free
                    }
                ],
                successUrl: process.env.PAYMENT_SUCCESS_URL,
                authenticatedUsersOnly: true
            }),
            portal({
                returnUrl: `${process.env.ORIGIN}/settings/billing`
            }),
            usage(),
        ],
    })
    : null;

export const auth = betterAuth({
    baseURL: process.env.ORIGIN,
    database: pool,
    advanced: {
        database: {
            generateId: false, // "serial" for auto-incrementing numeric IDs
        },
    },
    trustedOrigins: [process.env.PAYMENT_SUCCESS_URL || 'http://127.0.0.1:3000'],
    emailAndPassword: {
        enabled: true,
        async sendResetPassword(url, user) {
            console.log("Reset password url:", url);
        },
    },
    plugins: [
        jwt(),
        ...(polarPlugin ? [polarPlugin] : []),
        sveltekitCookies(getRequestEvent),
    ]
})
