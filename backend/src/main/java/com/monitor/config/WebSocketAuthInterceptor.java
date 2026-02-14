package com.monitor.config;

import com.monitor.entity.Company;
import com.monitor.repository.CompanyRepository;
import com.monitor.service.JwtService;
import lombok.RequiredArgsConstructor;
import org.springframework.messaging.Message;
import org.springframework.messaging.MessageChannel;
import org.springframework.messaging.simp.stomp.StompCommand;
import org.springframework.messaging.simp.stomp.StompHeaderAccessor;
import org.springframework.messaging.support.ChannelInterceptor;
import org.springframework.messaging.support.MessageHeaderAccessor;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.UUID;

@Component
@RequiredArgsConstructor
public class WebSocketAuthInterceptor implements ChannelInterceptor {

    private final JwtService jwtService;
    private final CompanyRepository companyRepository;

    @Override
    public Message<?> preSend(Message<?> message, MessageChannel channel) {
        StompHeaderAccessor accessor = MessageHeaderAccessor.getAccessor(message, StompHeaderAccessor.class);
        if (accessor == null) {
            return message;
        }

        if (StompCommand.CONNECT.equals(accessor.getCommand())) {
            UUID companyId = resolveCompanyId(accessor);
            if (companyId == null) {
                throw new IllegalArgumentException("Unauthorized");
            }

            UsernamePasswordAuthenticationToken authentication = new UsernamePasswordAuthenticationToken(companyId,
                    null, List.of());
            accessor.setUser(authentication);
        }

        return message;
    }

    private UUID resolveCompanyId(StompHeaderAccessor accessor) {
        String authHeader = accessor.getFirstNativeHeader("Authorization");
        if (authHeader != null && authHeader.startsWith("Bearer ")) {
            String token = authHeader.substring(7);
            try {
                return jwtService.extractCompanyId(token);
            } catch (Exception ignored) {
                return null;
            }
        }

        String agentToken = accessor.getFirstNativeHeader("x-agent-token");
        if (agentToken == null || agentToken.isBlank()) {
            return null;
        }

        return companyRepository.findByApiToken(agentToken)
                .map(Company::getId)
                .orElse(null);
    }
}
