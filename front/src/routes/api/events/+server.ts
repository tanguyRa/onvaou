import type { RequestHandler } from "./$types";

import { proxyGeoGet } from "$lib/server/geo";

export const GET: RequestHandler = async ({ url }) => {
    return proxyGeoGet(url, "/events");
};
