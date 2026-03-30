const GEO_API_URL = process.env.GEO_API_URL || "http://localhost:8000";

const EXCLUDED_RESPONSE_HEADERS = new Set([
    "transfer-encoding",
    "connection",
    "keep-alive",
    "content-length",
    "content-encoding",
]);

function buildResponseHeaders(source: Headers): Headers {
    const headers = new Headers();

    for (const [key, value] of source.entries()) {
        if (!EXCLUDED_RESPONSE_HEADERS.has(key.toLowerCase())) {
            headers.set(key, value);
        }
    }

    return headers;
}

export async function proxyGeoGet(
    requestUrl: URL,
    pathname: string,
): Promise<Response> {
    const targetUrl = new URL(pathname, GEO_API_URL);
    targetUrl.search = requestUrl.search;

    let response: Response;

    try {
        response = await fetch(targetUrl, {
            method: "GET",
            headers: {
                Accept: "application/json",
            },
        });
    } catch (error) {
        console.error("[Geo Proxy] GET failed:", error);
        return new Response(
            JSON.stringify({
                error: "Bad Gateway",
                message: "Failed to proxy request to Geo API",
            }),
            {
                status: 502,
                headers: { "Content-Type": "application/json" },
            },
        );
    }

    return new Response(response.body, {
        status: response.status,
        statusText: response.statusText,
        headers: buildResponseHeaders(response.headers),
    });
}
