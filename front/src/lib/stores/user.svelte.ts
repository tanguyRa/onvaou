import { browser } from "$app/environment";
import { useSession, customer } from "$lib/auth-client";
import type { CustomerStateSubscription } from "@polar-sh/sdk/models/components/customerstatesubscription.js";

export interface UserData {
    id: string;
    email: string;
    name: string;
    image?: string | null;
    emailVerified: boolean;
    createdAt: Date;
    updatedAt: Date;
}

export interface UserState {
    // Session state
    isPending: boolean;
    isAuthenticated: boolean;

    // User data
    user: UserData | null;

    // CustomerStateSubscription data
    subscriptions: CustomerStateSubscription[];
    subscriptionLoading: boolean;
    subscriptionError: string | null;

    // Computed helpers
    activeSubscription: CustomerStateSubscription | null;
    hasActiveSubscription: boolean;
    subscriptionStatus: string | null;
}

interface SessionData {
    user: {
        id: string;
        email: string;
        name: string;
        image?: string | null;
        emailVerified: boolean;
        createdAt: Date;
        updatedAt: Date;
    };
}

interface SessionState {
    data: SessionData | null;
    isPending: boolean;
    error: unknown;
}

// SSR-safe stub that returns static pending state without creating subscriptions
const SSR_STUB_STATE: UserState = {
    isPending: true,
    isAuthenticated: false,
    user: null,
    subscriptions: [],
    subscriptionLoading: true,
    subscriptionError: null,
    activeSubscription: null,
    hasActiveSubscription: false,
    subscriptionStatus: null,
};

type UserStore = {
    readonly state: UserState;
    readonly session: ReturnType<typeof useSession> | null;
    refreshSubscriptions: () => Promise<void>;
    destroy: () => void;
};

const ssrStub: UserStore = {
    get state(): UserState {
        return SSR_STUB_STATE;
    },
    get session() {
        return null;
    },
    refreshSubscriptions: async () => {},
    destroy: () => {},
};

function createUserStore() {
    const sessionAtom = useSession();

    // Reactive state that mirrors the session atom
    let sessionState = $state<SessionState>({ data: null, isPending: true, error: null });
    let subscriptions = $state<CustomerStateSubscription[]>([]);
    let subscriptionLoading = $state(true);
    let subscriptionError = $state<string | null>(null);
    let subscriptionFetched = false;
    let isFetchingSubscriptions = false;

    // Subscribe to session atom changes and store cleanup function
    const unsubscribe = sessionAtom.subscribe((value) => {
        const typedValue = value as SessionState;
        const wasAuthenticated = !!sessionState.data;
        const isAuthenticated = !!typedValue.data;

        sessionState = typedValue;

        // Reset subscription state on logout
        if (wasAuthenticated && !isAuthenticated) {
            subscriptions = [];
            subscriptionFetched = false;
            isFetchingSubscriptions = false;
            subscriptionLoading = true;
            subscriptionError = null;
        }
    });

    async function fetchSubscriptions() {
        // Guard against concurrent fetches and redundant fetches
        if (subscriptionFetched || isFetchingSubscriptions) return;

        isFetchingSubscriptions = true;
        subscriptionLoading = true;
        subscriptionError = null;
        try {
            if (!customer || typeof customer.state !== "function") {
                subscriptions = [];
                subscriptionFetched = true;
                return;
            }
            const { data: customerState } = await customer.state();
            if (customerState) {
                subscriptions = customerState.activeSubscriptions;
            }
            subscriptionFetched = true;
        } catch (e) {
            subscriptionError = "Failed to load subscription data";
            console.error("Failed to fetch subscriptions:", e);
        } finally {
            subscriptionLoading = false;
            isFetchingSubscriptions = false;
        }
    }

    async function refreshSubscriptions() {
        subscriptionFetched = false;
        isFetchingSubscriptions = false;
        await fetchSubscriptions();
    }

    // Watch for session changes and fetch subscriptions
    const effectCleanup = $effect.root(() => {
        $effect(() => {
            if (sessionState.data && !subscriptionFetched && !isFetchingSubscriptions) {
                fetchSubscriptions();
            }
        });
    });

    // Reactive state derived from session and subscriptions (pure, no side effects)
    const state = $derived.by(() => {
        const activeSubscription = subscriptions.find(s => s.status === "active") ?? null;

        return {
            isPending: sessionState.isPending,
            isAuthenticated: !!sessionState.data,
            user: sessionState.data?.user ? {
                id: sessionState.data.user.id,
                email: sessionState.data.user.email,
                name: sessionState.data.user.name,
                image: sessionState.data.user.image,
                emailVerified: sessionState.data.user.emailVerified,
                createdAt: sessionState.data.user.createdAt,
                updatedAt: sessionState.data.user.updatedAt,
            } : null,
            subscriptions,
            subscriptionLoading,
            subscriptionError,
            activeSubscription,
            hasActiveSubscription: !!activeSubscription,
            subscriptionStatus: activeSubscription?.status ?? null,
        } satisfies UserState;
    });

    return {
        get state() {
            return state;
        },
        get session() {
            return sessionAtom;
        },
        refreshSubscriptions,
        destroy() {
            unsubscribe();
            effectCleanup();
        },
    };
}

// Store instance - only created in browser to avoid SSR issues
let userStoreInstance: UserStore | null = null;

export function useUser(): UserStore {
    // Return SSR stub on server to avoid memory leaks from subscriptions
    if (!browser) {
        return ssrStub;
    }

    if (!userStoreInstance) {
        userStoreInstance = createUserStore();
    }
    return userStoreInstance;
}

// Reset store (useful for testing or full logout)
export function resetUserStore() {
    if (userStoreInstance) {
        userStoreInstance.destroy();
        userStoreInstance = null;
    }
}
