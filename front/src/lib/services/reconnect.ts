export interface ReconnectHandler {
    scheduleReconnect(): boolean;
    reset(): void;
    cancel(): void;
    getAttempts(): number;
    getNextDelay(): number;
}

export interface ReconnectOptions {
    maxAttempts?: number;
    maxDelayMs?: number;
    baseDelayMs?: number;
    onAttempt?: (attempt: number, delayMs: number) => void;
    onGiveUp?: () => void;
}

export function createReconnectHandler(
    connect: () => void,
    options: ReconnectOptions = {}
): ReconnectHandler {
    const {
        maxAttempts = 5,
        maxDelayMs = 30000,
        baseDelayMs = 1000,
        onAttempt,
        onGiveUp,
    } = options;

    let attempts = 0;
    let timeoutId: ReturnType<typeof setTimeout> | null = null;

    function getDelay(): number {
        // Exponential backoff: 1s, 2s, 4s, 8s, 16s, max 30s
        return Math.min(baseDelayMs * Math.pow(2, attempts), maxDelayMs);
    }

    return {
        scheduleReconnect(): boolean {
            if (attempts >= maxAttempts) {
                onGiveUp?.();
                return false;
            }

            const delay = getDelay();
            attempts++;

            onAttempt?.(attempts, delay);

            timeoutId = setTimeout(() => {
                timeoutId = null;
                connect();
            }, delay);

            return true;
        },

        reset() {
            attempts = 0;
            this.cancel();
        },

        cancel() {
            if (timeoutId) {
                clearTimeout(timeoutId);
                timeoutId = null;
            }
        },

        getAttempts() {
            return attempts;
        },

        getNextDelay() {
            return getDelay();
        },
    };
}
