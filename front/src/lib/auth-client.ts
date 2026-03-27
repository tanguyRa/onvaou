import { createAuthClient } from "better-auth/svelte"
import { polarClient } from "@polar-sh/better-auth";

const paymentProvider = ((import.meta.env.PUBLIC_PAYMENT_PROVIDER as string | undefined) || "").trim().toLowerCase();
const plugins = paymentProvider === "polar" ? [polarClient()] : [];

export const authClient = createAuthClient({
    plugins
})


export const {
    signIn,
    signOut,
    signUp,
    useSession,
    requestPasswordReset,
    resetPassword,
    changeEmail,
    changePassword,
    updateUser,
    checkout, usage, customer,
} = authClient;
