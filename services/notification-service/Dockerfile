# ─── Build stage ─────────────────────────────────────────────
FROM node:20-alpine AS build
WORKDIR /app

COPY package.json ./
RUN npm install

COPY tsconfig.json nest-cli.json ./
COPY src ./src
RUN npm run build

# ─── Runtime stage ───────────────────────────────────────────
FROM node:20-alpine
WORKDIR /app
RUN apk add --no-cache wget

COPY package.json ./
RUN npm install --omit=dev && npm cache clean --force
COPY --from=build /app/dist ./dist

EXPOSE 3003
HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
  CMD wget -qO- http://localhost:3003/health || exit 1

ENTRYPOINT ["node", "dist/main.js"]
