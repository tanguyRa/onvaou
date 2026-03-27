export type ConnectionState = 'idle' | 'connecting' | 'connected' | 'disconnected' | 'error';

export interface TerminalSocketEvents {
    onData: (data: string | Uint8Array) => void;
    onStateChange: (state: ConnectionState, error?: string) => void;
    onConnected?: (containerId: string) => void;
}

export interface TerminalSocket {
    connect(): void;
    disconnect(): void;
    send(data: string): void;
    resize(cols: number, rows: number): void;
    getState(): ConnectionState;
}

export interface TerminalSocketOptions {
    userId: string;
    events: TerminalSocketEvents;
}

export function createTerminalSocket(options: TerminalSocketOptions): TerminalSocket {
    const { userId, events } = options;

    let ws: WebSocket | null = null;
    let state: ConnectionState = 'idle';
    let pendingResize: { cols: number; rows: number } | null = null;

    function setState(newState: ConnectionState, error?: string) {
        state = newState;
        events.onStateChange(newState, error);
    }

    function getWebSocketUrl(): string {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        return `${protocol}//${window.location.host}/api/containers/${userId}/terminal`;
    }

    return {
        connect() {
            if (ws) {
                ws.close();
            }

            setState('connecting');

            const url = getWebSocketUrl();
            ws = new WebSocket(url);
            ws.binaryType = 'arraybuffer';

            ws.onopen = () => {
                setState('connected');
                // Send initial resize if we have one pending
                if (pendingResize) {
                    ws?.send(JSON.stringify({ type: 'resize', ...pendingResize }));
                }
            };

            ws.onmessage = (event) => {
                if (typeof event.data === 'string') {
                    // Try to parse as JSON control message
                    if (event.data.startsWith('{')) {
                        try {
                            const msg = JSON.parse(event.data);
                            if (msg.type === 'connected') {
                                events.onConnected?.(msg.containerId);
                                return;
                            }
                            if (msg.type === 'error') {
                                setState('error', msg.message || msg.code);
                                return;
                            }
                            if (msg.type === 'disconnected') {
                                setState('disconnected', msg.reason);
                                return;
                            }
                        } catch {
                            // Not JSON, treat as terminal data
                        }
                    }
                    events.onData(event.data);
                } else if (event.data instanceof ArrayBuffer) {
                    // Binary data from terminal
                    events.onData(new Uint8Array(event.data));
                }
            };

            ws.onclose = (event) => {
                if (state !== 'error') {
                    setState('disconnected', event.reason || undefined);
                }
                ws = null;
            };

            ws.onerror = () => {
                setState('error', 'WebSocket connection failed');
            };
        },

        disconnect() {
            if (ws) {
                ws.close(1000, 'User disconnect');
                ws = null;
            }
            setState('idle');
        },

        send(data: string) {
            if (ws?.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({ type: 'input', data }));
            }
        },

        resize(cols: number, rows: number) {
            pendingResize = { cols, rows };
            if (ws?.readyState === WebSocket.OPEN) {
                ws.send(JSON.stringify({ type: 'resize', cols, rows }));
            }
        },

        getState() {
            return state;
        },
    };
}
