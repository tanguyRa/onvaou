<script lang="ts">
    import { onMount } from "svelte";

    import type { EventSummary, MapCenter } from "$lib/types/geo";

    interface Props {
        events?: EventSummary[];
        center: MapCenter;
        radiusKm: number;
        selectedEventId?: string | null;
        onSelect?: (event: EventSummary) => void;
    }

    let {
        events = [],
        center,
        radiusKm,
        selectedEventId = null,
        onSelect = () => {},
    }: Props = $props();

    let container: HTMLDivElement | undefined = $state();
    let ready = $state(false);

    let map: any = null;
    let markersLayer: any = null;
    let radiusCircle: any = null;

    async function ensureLeaflet() {
        if (typeof window === "undefined") return null;
        const existing = (window as Window & { L?: any }).L;
        if (existing) return existing;

        const cssId = "leaflet-cdn-css";
        if (!document.getElementById(cssId)) {
            const link = document.createElement("link");
            link.id = cssId;
            link.rel = "stylesheet";
            link.href = "https://unpkg.com/leaflet@1.9.4/dist/leaflet.css";
            document.head.appendChild(link);
        }

        await new Promise<void>((resolve, reject) => {
            const scriptId = "leaflet-cdn-js";
            const current = document.getElementById(scriptId) as HTMLScriptElement | null;
            if (current) {
                current.addEventListener("load", () => resolve(), { once: true });
                current.addEventListener("error", () => reject(new Error("Leaflet failed to load")), {
                    once: true,
                });
                if ((window as Window & { L?: any }).L) resolve();
                return;
            }

            const script = document.createElement("script");
            script.id = scriptId;
            script.src = "https://unpkg.com/leaflet@1.9.4/dist/leaflet.js";
            script.async = true;
            script.onload = () => resolve();
            script.onerror = () => reject(new Error("Leaflet failed to load"));
            document.head.appendChild(script);
        });

        return (window as Window & { L?: any }).L;
    }

    onMount(() => {
        let cancelled = false;

        void (async () => {
            const L = await ensureLeaflet();
            if (!L || !container || cancelled) return;

            map = L.map(container, {
                zoomControl: false,
            }).setView([center.lat, center.lon], 12);

            L.control.zoom({ position: "bottomright" }).addTo(map);

            L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
                attribution: "&copy; OpenStreetMap contributors",
            }).addTo(map);

            markersLayer = L.layerGroup().addTo(map);
            radiusCircle = L.circle([center.lat, center.lon], {
                radius: radiusKm * 1000,
                color: "#ff6b35",
                weight: 1.5,
                fillColor: "#ff6b35",
                fillOpacity: 0.08,
            }).addTo(map);

            ready = true;
        })();

        return () => {
            cancelled = true;
            if (map) {
                map.remove();
                map = null;
            }
        };
    });

    $effect(() => {
        if (!ready || !map || !radiusCircle) return;
        map.setView([center.lat, center.lon], map.getZoom());
        radiusCircle.setLatLng([center.lat, center.lon]);
        radiusCircle.setRadius(radiusKm * 1000);
    });

    $effect(() => {
        if (!ready || !map || !markersLayer) return;
        const L = (window as Window & { L?: any }).L;
        if (!L) return;

        markersLayer.clearLayers();

        const bounds: any[] = [[center.lat, center.lon]];

        for (const event of events) {
            const isSelected = event.event_id === selectedEventId;
            const icon = L.divIcon({
                className: "map-marker-shell",
                html: `<span class="map-marker${isSelected ? " active" : ""}"></span>`,
                iconSize: [22, 22],
                iconAnchor: [11, 11],
            });

            const marker = L.marker([event.lat, event.lon], { icon }).addTo(markersLayer);
            marker.on("click", () => onSelect(event));
            bounds.push([event.lat, event.lon]);
        }

        if (events.length > 0) {
            map.fitBounds(bounds, {
                padding: [40, 40],
                maxZoom: 13,
            });
        }
    });
</script>

<div class="map-shell">
    <div bind:this={container} class="map-canvas"></div>
    <div class="map-gradient"></div>
</div>

<style>
    .map-shell {
        position: relative;
        min-height: 30rem;
        height: 100%;
        border-radius: 1.75rem;
        overflow: hidden;
        isolation: isolate;
        background:
            radial-gradient(circle at top left, rgba(255, 154, 118, 0.45), transparent 28%),
            linear-gradient(135deg, #fff7f2 0%, #f5f7fb 100%);
        box-shadow: var(--shadow-lg);
    }

    .map-canvas {
        position: absolute;
        inset: 0;
    }

    .map-gradient {
        position: absolute;
        inset: 0;
        pointer-events: none;
        background: linear-gradient(180deg, rgba(255, 255, 255, 0.08), transparent 16%, transparent 84%, rgba(17, 24, 39, 0.08));
    }

    :global(.map-marker-shell) {
        background: transparent;
        border: 0;
    }

    :global(.map-marker) {
        display: block;
        width: 22px;
        height: 22px;
        border-radius: 999px;
        border: 3px solid white;
        background: var(--color-primary);
        box-shadow: 0 10px 24px rgba(255, 107, 53, 0.35);
        transition: transform 0.15s ease;
    }

    :global(.map-marker.active) {
        transform: scale(1.22);
        background: #111827;
    }
</style>
