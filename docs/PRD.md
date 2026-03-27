# PRD: L'Événement Local (FR-Aggregator)

## 1. Project Overview
**L'Événement Local** is a high-precision data discovery tool designed to bridge the gap between fragmented French public data and its citizens. It provides a real-time, radius-based search (e.g., 5km) for cultural, municipal, and social events occurring within a 4-week window across France.

## 2. Target Audience
* **Local Residents:** Users looking for hyper-local activities (festivals, markets, workshops).
* **Regional Planners/Tourists:** Users exploring specific French "Communes" or "Départements."

---

## 3. Functional Requirements

### 3.1 Geospatial Search Engine
* **City-to-Coord Resolution:** Must integrate the **Base Adresse Nationale (BAN) API** to convert French city names or ZIP codes into precise $WGS84$ coordinates.
* **Dynamic Radius Filtering:** Users can toggle a search radius between **1km and 50km**.
* **Temporal Window:** The engine strictly filters for events with a `start_date` between `Today` and `Today + 28 days`.

### 3.2 Data Ingestion Pipeline (The Harvester)
* **Official OpenData:** Scheduled fetching from `data.gouv.fr` (national) and regional portals.
* **Cultural Hubs:** Deep integration with the **OpenAgenda API** (the primary source for French municipal agendas).
* **Automated Scraping:** Use **Playwright (Python)** to monitor official "Mairie" (Town Hall) agenda pages that lack public APIs.
* **Social Export:** Periodic ingestion of public event data via Meta Graph API or third-party scrapers (e.g., Apify) targeting regional French handles.

### 3.3 Intelligent Normalization
* **Deduplication:** A Python-based matching logic (using `Levenshtein` distance or `FuzzyWuzzy`) to identify duplicate listings across multiple sources.
* **Standardization:** All addresses must be normalized to the French national standard to ensure spatial accuracy.

---

## 4. Technical Specifications

### 4.1 The Backend Stack (Python)
* **Framework:** **FastAPI** for high-concurrency asynchronous API requests.
* **Processing Engine:** **GeoPandas** for all spatial transformations and data cleaning.
* **Geometry Logic:** Use **Shapely** to handle the creation of 5km "Buffer" zones for point-in-polygon queries.

### 4.2 The Storage Layer (SpatiaLite)
* **Database:** **SQLite 3** with the **SpatiaLite** extension.
* **Spatial Index:** Mandatory `R-Tree` index on the `geometry` column for $O(\log N)$ search performance.
* **Coordinate Systems:**
    * **Storage:** $EPSG:4326$ (Standard GPS).
    * **Calculation:** Project to **$EPSG:2154$ (Lambert-93)** for high-precision metric distance calculations within France.

### 4.3 Database Schema (Key Entities)
| Field | Type | Description |
| :--- | :--- | :--- |
| `event_id` | UUID | Primary Key |
| `title` | String | Event name (Sanitized) |
| `description` | Text | Full event details |
| `start_dt` | DateTime | Start timestamp |
| `geom` | Point | SpatiaLite Geometry (Point, 4326) |
| `source_tag` | Enum | e.g., 'OpenAgenda', 'Mairie_Lyon', 'FB' |

---

## 5. Logic & Mathematics

### Distance Query Logic
To ensure the "5km radius" is accurate despite the Earth's curvature, the system uses the SpatiaLite `ST_Distance` function on projected coordinates:

$$Dist(P_1, P_2) = \text{ST\_Distance}(\text{ST\_Transform}(P_1, 2154), \text{ST\_Transform}(P_2, 2154)) < 5000$$

### Data Lifecycle
1.  **Daily Sync (03:00 CET):** Clear expired events and poll new data from all sources.
2.  **Hourly Refresh:** Update commercial ticket availability for major events.
3.  **Vacuum:** Run `VACUUM` on the SQLite file weekly to optimize disk space.

---

## 6. Compliance & Constraints
* **RGPD:** No storage of user PII (Personally Identifiable Information). All event data stored is strictly public domain.
* **Rate Limiting:** Python scrapers must implement `random_sleep` intervals to respect French government server limitations and avoid IP blacklisting.
* **Memory:** Max 2GB RAM usage for the GeoPandas processing layer to keep hosting costs low on a single VPS.