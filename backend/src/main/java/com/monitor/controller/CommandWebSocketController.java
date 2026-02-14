package com.monitor.controller;

import com.monitor.dto.CommandRequest;
import com.monitor.dto.CommandResult;
import com.monitor.entity.Device;
import com.monitor.repository.DeviceRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.messaging.handler.annotation.DestinationVariable;
import org.springframework.messaging.handler.annotation.MessageMapping;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.security.core.Authentication;
import org.springframework.stereotype.Controller;

import java.security.Principal;
import java.util.UUID;

@Controller
@RequiredArgsConstructor
public class CommandWebSocketController {

    private final DeviceRepository deviceRepository;
    private final SimpMessagingTemplate messagingTemplate;

    @MessageMapping("/command/{deviceId}")
    public void sendCommand(@DestinationVariable UUID deviceId,
            CommandRequest request,
            Principal principal) {
        UUID companyId = getCompanyId(principal);
        Device device = deviceRepository.findById(deviceId)
                .orElseThrow(() -> new IllegalArgumentException("Device not found"));

        if (!device.getCompany().getId().equals(companyId)) {
            return;
        }

        if (request.getDeviceId() == null) {
            request.setDeviceId(deviceId);
        }
        if (request.getCommandId() == null || request.getCommandId().isBlank()) {
            request.setCommandId(UUID.randomUUID().toString());
        }

        messagingTemplate.convertAndSend("/topic/agent/" + deviceId, request);
    }

    @MessageMapping("/command-result")
    public void sendCommandResult(CommandResult result, Principal principal) {
        UUID companyId = getCompanyId(principal);
        if (result == null || result.getDeviceId() == null) {
            return;
        }

        Device device = deviceRepository.findById(result.getDeviceId())
                .orElseThrow(() -> new IllegalArgumentException("Device not found"));

        if (!device.getCompany().getId().equals(companyId)) {
            return;
        }

        messagingTemplate.convertAndSend(
                "/topic/command-result/" + result.getDeviceId(),
                result);
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
