# syntax=docker/dockerfile:1

# ============================================
# Stage 1: Dependencies
# ============================================
FROM oven/bun:1 AS deps

WORKDIR /app

COPY package.json bun.lock* ./
RUN bun install

# ============================================
# Stage 2: Development
# ============================================
FROM deps AS development

WORKDIR /app
COPY . .

EXPOSE 5173
ENTRYPOINT ["bun", "run", "dev", "--host", "0.0.0.0"]

# ============================================
# Stage 3: Builder
# ============================================
FROM deps AS builder

WORKDIR /app
COPY . .

RUN bun --bun run build

# ============================================
# Stage 4: Production
# ============================================
FROM oven/bun:1 AS production

WORKDIR /app

COPY --from=builder /app/build ./build
COPY --from=builder /app/node_modules ./node_modules

EXPOSE 3000
ENTRYPOINT ["bun", "./build/index.js"]
