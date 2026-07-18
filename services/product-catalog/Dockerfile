# ─── Build stage ─────────────────────────────────────────────
FROM golang:1.24-alpine AS build
WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download || go mod tidy

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /product-catalog .

# ─── Runtime stage ───────────────────────────────────────────
FROM alpine:3.21
RUN apk add --no-cache wget
COPY --from=build /product-catalog /product-catalog

EXPOSE 3001
HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
  CMD wget -qO- http://localhost:3001/health || exit 1

ENTRYPOINT ["/product-catalog"]
