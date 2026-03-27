import { json } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";
import { getPaymentClient } from "$lib/server/payments/client";
import { auth } from "$lib/auth";

export const POST: RequestHandler = async ({ request }) => {
    try {
        const session = await auth.api.getSession({
            headers: request.headers
        });
        if (!session?.user) {
            return json({ error: "Not authenticated" }, { status: 401 });
        }

        const body = await request.json().catch(() => null) as { slug?: string } | null;
        const slug = body?.slug?.trim();
        if (!slug) {
            return json({ error: "Missing checkout slug" }, { status: 400 });
        }

        const client = getPaymentClient();
        if (client.provider === "none") {
            return json({ error: "No payment provider configured" }, { status: 400 });
        }

        const successUrl = process.env.POLAR_SUCCESS_URL || `${process.env.ORIGIN}/success`;
        const cancelUrl = `${process.env.ORIGIN}/pricing`;

        const checkout = await client.createCheckout({
            slug,
            successUrl,
            cancelUrl,
            customerEmail: session.user.email,
            externalCustomerId: session.user.id,
        });

        return json({ url: checkout.url });
    } catch (error) {
        console.error("Failed to create checkout:", error);
        return json({ error: "Failed to create checkout" }, { status: 500 });
    }
};
