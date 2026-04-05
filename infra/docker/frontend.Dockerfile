# --- Dependencies stage ---
FROM node:24-bookworm-slim AS deps

WORKDIR /build
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --ignore-scripts

# --- Build stage ---
FROM node:24-bookworm-slim AS builder

WORKDIR /build
COPY --from=deps /build/node_modules ./node_modules
COPY frontend/ .
RUN npm run postinstall && npm run build

# --- Runtime stage ---
FROM node:24-bookworm-slim

WORKDIR /app
COPY --from=builder /build/.output ./.output

ENV HOST=0.0.0.0
ENV PORT=3000
ENV NODE_ENV=production

EXPOSE 3000
USER node
CMD ["node", ".output/server/index.mjs"]
