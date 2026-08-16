# ─── Build stage ─────────────────────────────────────────────
FROM maven:3.9-eclipse-temurin-21 AS build
WORKDIR /app

# Cache dependencies against the POM before copying sources.
COPY pom.xml ./
RUN mvn -q -B dependency:go-offline

COPY src ./src
RUN mvn -q -B clean package -DskipTests

# ─── Runtime stage ───────────────────────────────────────────
FROM eclipse-temurin:21-jre
RUN apt-get update && apt-get install -y --no-install-recommends curl \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app

# OpenTelemetry Java agent — zero-code instrumentation for spring-web, JDBC
# (db.postgresql spans), spring-kafka (auto W3C traceparent inject/extract on
# the async Kafka hop), and logback (trace_id/span_id into MDC + OTLP log
# export). Pinned and attached via JAVA_TOOL_OPTIONS, so neither pom.xml nor any
# application code changes — the agent is a pure runtime concern. Downloaded
# before the app jar so this large, version-stable layer stays cached.
ARG OTEL_AGENT_VERSION=2.30.0
RUN mkdir -p /otel \
    && curl -fsSL -o /otel/opentelemetry-javaagent.jar \
       "https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/v${OTEL_AGENT_VERSION}/opentelemetry-javaagent.jar"
ENV JAVA_TOOL_OPTIONS="-javaagent:/otel/opentelemetry-javaagent.jar"

COPY --from=build /app/target/inventory-service-*.jar app.jar

EXPOSE 3002
HEALTHCHECK --interval=10s --timeout=3s --retries=5 --start-period=40s \
  CMD curl -fsS http://localhost:3002/health || exit 1

ENTRYPOINT ["java", "-jar", "app.jar"]
