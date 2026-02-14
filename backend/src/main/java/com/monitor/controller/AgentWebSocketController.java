package com.monitor.controller;

import com.monitor.dto.MetricRequest;
import com.monitor.entity.Device;
import com.monitor.repository.DeviceRepository;
import com.monitor.service.AgentService;
import lombok.RequiredArgsConstructor;
import org.springframework.messaging.handler.annotation.MessageMapping;
import org.springframework.security.core.Authentication;
import org.springframework.stereotype.Controller;

import java.security.Principal;
import java.util.UUID;

@Controller
@RequiredArgsConstructor
public class AgentWebSocketController {

    private final DeviceRepository deviceRepository;
    private final AgentService agentService;

    @MessageMapping("/agent/metrics")
    public void receiveMetrics(MetricRequest request, Principal principal) {
        if (request == null || request.getDeviceId() == null) {
            return;
        }

        UUID companyId = getCompanyId(principal);
        Device device = deviceRepository.findById(request.getDeviceId())
                .orElseThrow(() -> new IllegalArgumentException("Device not found"));

        if (!device.getCompany().getId().equals(companyId)) {
            return;
        }

        agentService.saveMetric(request);
    }

    private UUID getCompanyId(Principal principal) {
        if (principal instanceof Authentication authentication) {
            Object value = authentication.getPrincipal();
            if (value instanceof UUID uuid) {
                return uuid;
            }
            if (value instanceof String text) {
                return UUID.fromString(text);
            }
        }
        throw new IllegalArgumentException("Unauthorized");
    }
}
