import type {
    CitySuggestion,
    EventDetail,
    EventSearchResponse,
    EventSummary,
    MapCenter,
} from "$lib/types/geo";

const DEFAULT_CENTER: MapCenter = {
    lat: 45.764,
    lon: 4.8357,
};

export class MapDiscoveryController {
    searchValue = $state("");
    selectedCity = $state<CitySuggestion | null>(null);
    suggestions = $state<CitySuggestion[]>([]);
    radiusKm = $state(10);
    center = $state<MapCenter>(DEFAULT_CENTER);
    events = $state<EventSummary[]>([]);
    total = $state(0);
    loadingEvents = $state(false);
    loadingSuggestions = $state(false);
    selectedEventId = $state<string | null>(null);
    selectedEvent = $state<EventDetail | null>(null);
    detailLoading = $state(false);
    error = $state("");

    #initialCity: string;
    #suggestionTimer: number | null = null;
    #eventTimer: number | null = null;

    constructor(initialCity = "Lyon") {
        this.#initialCity = initialCity;
        this.searchValue = initialCity;
    }

    get hasResults() {
        return this.events.length > 0;
    }

    get cityLabel() {
        return this.selectedCity?.city ?? this.searchValue.trim();
    }

    async initialize() {
        await this.#loadInitialCity();
    }

    destroy() {
        if (this.#suggestionTimer !== null) {
            window.clearTimeout(this.#suggestionTimer);
            this.#suggestionTimer = null;
        }

        if (this.#eventTimer !== null) {
            window.clearTimeout(this.#eventTimer);
            this.#eventTimer = null;
        }
    }

    setSearchValue(value: string) {
        this.searchValue = value;

        if (this.selectedCity && value.trim() !== this.selectedCity.city) {
            this.selectedCity = null;
        }

        const query = value.trim();
        if (query.length < 2) {
            this.suggestions = [];
            this.loadingSuggestions = false;
            if (this.#suggestionTimer !== null) {
                window.clearTimeout(this.#suggestionTimer);
                this.#suggestionTimer = null;
            }
            return;
        }

        if (
            this.selectedCity &&
            (query.toLowerCase() === this.selectedCity.city.toLowerCase() ||
                query === this.selectedCity.postcode)
        ) {
            this.suggestions = [];
            this.loadingSuggestions = false;
            return;
        }

        if (this.#suggestionTimer !== null) {
            window.clearTimeout(this.#suggestionTimer);
        }

        this.#suggestionTimer = window.setTimeout(() => {
            void this.loadSuggestions(query);
        }, 220);
    }

    setRadius(value: number) {
        this.radiusKm = value;
        if (!this.selectedCity) return;

        if (this.#eventTimer !== null) {
            window.clearTimeout(this.#eventTimer);
        }

        this.#eventTimer = window.setTimeout(() => {
            if (this.selectedCity) {
                void this.loadEvents(this.selectedCity, this.radiusKm);
            }
        }, 300);
    }

    async submitSearch() {
        const query = this.searchValue.trim();
        if (!query) return;

        if (this.suggestions.length === 0) {
            await this.loadSuggestions(query);
        }

        const exactMatch =
            this.suggestions.find(
                (suggestion) =>
                    suggestion.city.toLowerCase() === query.toLowerCase() ||
                    suggestion.postcode === query,
            ) || this.suggestions[0];

        if (exactMatch) {
            await this.selectCity(exactMatch);
        }
    }

    async selectCity(city: CitySuggestion) {
        this.selectedCity = city;
        this.searchValue = city.city;
        this.suggestions = [];
        this.center = { lat: city.lat, lon: city.lon };
        await this.loadEvents(city, this.radiusKm);
    }

    async selectEvent(event: EventSummary) {
        this.selectedEventId = event.event_id;
        await this.loadEventDetail(event.event_id);
    }

    closeEvent(eventId?: string) {
        if (eventId && this.selectedEventId !== eventId) return;
        this.selectedEventId = null;
        this.selectedEvent = null;
    }

    async loadSuggestions(query: string) {
        this.loadingSuggestions = true;

        try {
            const response = await fetch(
                `/api/cities/search?q=${encodeURIComponent(query)}`,
            );
            if (!response.ok) {
                throw new Error("Suggestions request failed");
            }

            this.suggestions = (await response.json()) as CitySuggestion[];
        } catch (fetchError) {
            console.error(fetchError);
            this.suggestions = [];
        } finally {
            this.loadingSuggestions = false;
        }
    }

    async loadEvents(city: CitySuggestion, radius: number) {
        this.loadingEvents = true;
        this.error = "";

        try {
            const params = new URLSearchParams({
                city: city.city,
                radius_km: String(radius),
            });

            const response = await fetch(`/api/events?${params.toString()}`);
            const payload = await response.json();

            if (!response.ok) {
                const message =
                    typeof payload === "object" &&
                    payload !== null &&
                    "detail" in payload &&
                    typeof payload.detail === "string"
                        ? payload.detail
                        : "Event search failed";
                throw new Error(message);
            }

            const results = payload as EventSearchResponse;

            this.events = results.results;
            this.total = results.total;
            this.center = { lat: city.lat, lon: city.lon };

            if (
                this.selectedEventId &&
                !results.results.some(
                    (event) => event.event_id === this.selectedEventId,
                )
            ) {
                this.closeEvent();
            }
        } catch (fetchError) {
            console.error(fetchError);
            this.error = "Unable to load nearby events right now.";
            this.events = [];
            this.total = 0;
        } finally {
            this.loadingEvents = false;
        }
    }

    async loadEventDetail(eventId: string) {
        this.detailLoading = true;

        try {
            const response = await fetch(`/api/events/${eventId}`);
            if (!response.ok) {
                throw new Error("Event detail request failed");
            }

            this.selectedEvent = (await response.json()) as EventDetail;
        } catch (fetchError) {
            console.error(fetchError);
            this.selectedEvent = null;
        } finally {
            this.detailLoading = false;
        }
    }

    async #loadInitialCity() {
        try {
            const response = await fetch(
                `/api/cities/search?q=${encodeURIComponent(this.#initialCity)}`,
            );
            if (!response.ok) return;

            const results = (await response.json()) as CitySuggestion[];
            const first = results[0];
            if (!first) return;

            await this.selectCity(first);
        } catch (fetchError) {
            console.error("Failed to initialize city:", fetchError);
        }
    }
}

export function createMapDiscovery(initialCity = "Lyon") {
    return new MapDiscoveryController(initialCity);
}
