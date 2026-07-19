package com.ieeeyp.ecommerce.inventory.config;

import java.net.URI;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.env.EnvironmentPostProcessor;
import org.springframework.core.env.ConfigurableEnvironment;
import org.springframework.core.env.MapPropertySource;

import java.util.HashMap;
import java.util.Map;

/**
 * Translates the shared {@code DATABASE_URL} convention
 * ({@code postgres://user:pass@host:port/db}) - used identically by every
 * service in this stack - into the JDBC URL, username and password that Spring
 * Data JPA expects. This keeps docker-compose wiring uniform across Go and Java
 * services instead of forcing Java-specific env vars.
 */
public class DatabaseUrlEnvironmentPostProcessor implements EnvironmentPostProcessor {

    @Override
    public void postProcessEnvironment(ConfigurableEnvironment environment, SpringApplication application) {
        String databaseUrl = environment.getProperty("DATABASE_URL");
        if (databaseUrl == null || databaseUrl.isBlank()) {
            return;
        }

        URI uri = URI.create(databaseUrl.replaceFirst("^postgres(ql)?://", "http://"));
        String host = uri.getHost();
        int port = uri.getPort() == -1 ? 5432 : uri.getPort();
        String path = uri.getPath();
        String query = uri.getQuery() == null ? "" : "?" + uri.getQuery();
        String jdbcUrl = "jdbc:postgresql://" + host + ":" + port + path + query;

        Map<String, Object> props = new HashMap<>();
        props.put("spring.datasource.url", jdbcUrl);

        String userInfo = uri.getUserInfo();
        if (userInfo != null) {
            String[] parts = userInfo.split(":", 2);
            props.put("spring.datasource.username", parts[0]);
            if (parts.length > 1) {
                props.put("spring.datasource.password", parts[1]);
            }
        }

        environment.getPropertySources()
                .addFirst(new MapPropertySource("databaseUrlFromEnv", props));
    }
}
