export interface EventSummary {
    event_id: string;
    title: string;
    start_dt: string;
    address: string;
    lat: number;
    lon: number;
    source_tag: string;
    source_url: string;
}

export interface EventDetail extends EventSummary {
    description: string;
    end_dt: string | null;
    location_name: string;
}

export interface EventSearchResponse {
    total: number;
    page: number;
    results: EventSummary[];
}

export interface CitySuggestion {
    name: string;
    city: string;
    postcode: string;
    lat: number;
    lon: number;
}

export interface MapCenter {
    lat: number;
    lon: number;
}
