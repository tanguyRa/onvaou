import type { RequestHandler } from './$types';
import { auth } from '$lib/auth';

const GO_API_URL = process.env.GO_API_URL || 'http://localhost:8080';

// Headers that should not be forwarded
const EXCLUDED_RESPONSE_HEADERS = new Set([
    'transfer-encoding',
    'connection',
    'keep-alive',
    'content-length',
    'content-encoding'
]);

async function handleRequest(event: Parameters<RequestHandler>[0]): Promise<Response> {
    const { request, params } = event;

    // Build the target URL - the path includes everything after /api/
    const path = params.path;
    const targetUrl = `${GO_API_URL}/api/${path}${event.url.search}`;

    // Get request body for non-GET/HEAD requests
    let body: string | null = null;
    if (request.method !== 'GET' && request.method !== 'HEAD') {
        body = await request.text();
    }

    // Get JWT token from better-auth if session cookies are present
    let token: string | undefined;
    if (request.headers.has('cookie')) {
        try {
            const tokenResult = await auth.api.getToken({
                headers: request.headers
            });
            token = tokenResult?.token;
        } catch (err) {
            console.warn('[API Proxy] auth.api.getToken failed:', err);
        }
    }

    // Debug logging
    if (!token) {
        console.warn('[API Proxy] No token received from auth.api.getToken');
        console.warn('[API Proxy] Request path:', path);
        console.warn('[API Proxy] Cookie header present:', request.headers.has('cookie'));
    }

    // Prepare headers for the proxied request
    const headers = new Headers();

    // Forward relevant headers from the original request
    for (const [key, value] of request.headers.entries()) {
        // Skip hop-by-hop headers and host
        if (!['host', 'connection', 'keep-alive', 'transfer-encoding'].includes(key.toLowerCase())) {
            headers.set(key, value);
        }
    }

    // Set the Authorization header with JWT token
    if (token) {
        headers.set('Authorization', `Bearer ${token}`);
    } else {
        console.warn('[API Proxy] Forwarding request without Authorization header');
    }

    // Make the request to the Go API
    let response: Response;
    try {
        response = await fetch(targetUrl, {
            method: request.method,
            headers,
            body: body || undefined,
        });
    } catch (error) {
        console.error('Proxy error:', error);
        return new Response(JSON.stringify({
            error: 'Bad Gateway',
            message: 'Failed to proxy request to Go API'
        }), {
            status: 502,
            headers: { 'Content-Type': 'application/json' }
        });
    }

    // Build response headers, excluding certain ones
    const responseHeaders = new Headers();
    for (const [key, value] of response.headers.entries()) {
        if (!EXCLUDED_RESPONSE_HEADERS.has(key.toLowerCase())) {
            responseHeaders.set(key, value);
        }
    }

    // Return the proxied response
    return new Response(response.body, {
        status: response.status,
        statusText: response.statusText,
        headers: responseHeaders
    });
}

// Handle all HTTP methods
export const GET: RequestHandler = handleRequest;
export const POST: RequestHandler = handleRequest;
export const PUT: RequestHandler = handleRequest;
export const PATCH: RequestHandler = handleRequest;
export const DELETE: RequestHandler = handleRequest;
export const OPTIONS: RequestHandler = handleRequest;
