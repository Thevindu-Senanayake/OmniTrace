# ─── Build stage ─────────────────────────────────────────────
FROM golang:1.24-alpine AS build
WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download || go mod tidy

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /order-service .

# ─── Runtime stage ───────────────────────────────────────────
FROM alpine:3.21
RUN apk add --no-cache wget
COPY --from=build /order-service /order-service

EXPOSE 3000
HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
  CMD wget -qO- http://localhost:3000/health || exit 1

ENTRYPOINT ["/order-service"]
