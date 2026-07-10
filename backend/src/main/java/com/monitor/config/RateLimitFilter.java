package com.monitor.config;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Component
public class RateLimitFilter extends OncePerRequestFilter {

    private static final String LOAD_TESTER_HEADER = "X-Load-Tester";
    private static final String LOAD_TESTER_USER_AGENT_PREFIX = "monitor-loadtester/";

    private final Map<String, Counter> counters = new ConcurrentHashMap<>();

    @Value("${app.rateLimit.windowSeconds:60}")
    private long windowSeconds;

    @Value("${app.rateLimit.maxRequests:120}")
    private int maxRequests;

    @Override
    protected boolean shouldNotFilter(HttpServletRequest request) {
        String path = request.getRequestURI();
        if (path.equals("/agent") || path.startsWith("/agent/") || path.equals("/ws") || path.startsWith("/ws/")) {
            return true;
        }

        String loadTester = request.getHeader(LOAD_TESTER_HEADER);
        if ("true".equalsIgnoreCase(loadTester)) {
            return true;
        }

        String userAgent = request.getHeader("User-Agent");
        return userAgent != null && userAgent.startsWith(LOAD_TESTER_USER_AGENT_PREFIX);
    }

    @Override
    protected void doFilterInternal(HttpServletRequest request,
            HttpServletResponse response,
            FilterChain filterChain) throws ServletException, IOException {
        String key = buildKey(request);
        long now = System.currentTimeMillis();

        Counter counter = counters.compute(key, (k, existing) -> {
            if (existing == null || now - existing.windowStartMs >= windowSeconds * 1000) {
                return new Counter(now, 1);
            }
            existing.count++;
            return existing;
        });

        if (counter.count > maxRequests) {
            response.setStatus(HttpStatus.TOO_MANY_REQUESTS.value());
            return;
        }

        filterChain.doFilter(request, response);
    }

    private String buildKey(HttpServletRequest request) {
        String ip = request.getHeader("X-Forwarded-For");
        if (ip == null || ip.isBlank()) {
            ip = request.getRemoteAddr();
        } else {
            ip = ip.split(",")[0].trim();
        }

        return ip + "|" + request.getRequestURI();
    }

    private static class Counter {
        private final long windowStartMs;
        private int count;

        private Counter(long windowStartMs, int count) {
            this.windowStartMs = windowStartMs;
            this.count = count;
        }
    }
}
