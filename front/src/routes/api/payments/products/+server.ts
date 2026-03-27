import { json } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";
import { getPaymentClient } from "$lib/server/payments/client";

export const GET: RequestHandler = async () => {
    try {
        const client = getPaymentClient();
        const products = await client.listProducts();
        return json({ provider: client.provider, products });
    } catch (error) {
        console.error("Failed to load payment products:", error);
        return json({ error: "Failed to load payment products" }, { status: 500 });
    }
};
