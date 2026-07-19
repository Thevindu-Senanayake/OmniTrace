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

COPY --from=build /app/target/payment-service-*.jar app.jar

EXPOSE 3004
HEALTHCHECK --interval=10s --timeout=3s --retries=5 --start-period=40s \
  CMD curl -fsS http://localhost:3004/health || exit 1

ENTRYPOINT ["java", "-jar", "app.jar"]
