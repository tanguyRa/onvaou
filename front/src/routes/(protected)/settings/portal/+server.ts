import { CustomerPortal } from "@polar-sh/sveltekit";
import { auth } from "$lib/auth";

export const GET = CustomerPortal({
    server: process.env.POLAR_SERVER === 'production' ? 'production' : 'sandbox',
    accessToken: process.env.POLAR_ACCESS_TOKEN!,
    returnUrl: `${process.env.ORIGIN}/settings/billing`,
    getExternalCustomerId: async (event) => {
        const session = await auth.api.getSession({
            headers: event.request.headers,
        });

        if (!session?.user?.id) {
            throw new Error("User not authenticated");
        }

        return session.user.id;
    },
});