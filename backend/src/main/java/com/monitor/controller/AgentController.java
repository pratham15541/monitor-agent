package com.monitor.controller;

import com.monitor.dto.AgentRegisterRequest;
import com.monitor.dto.DeviceResponse;
import com.monitor.dto.MetricDetailRequest;
import com.monitor.dto.MetricRequest;
import com.monitor.entity.Device;
import com.monitor.service.AgentService;
import lombok.RequiredArgsConstructor;
import jakarta.validation.Valid;
import org.springframework.web.bind.annotation.*;
import org.springframework.validation.annotation.Validated;

import java.util.List;

@RestController
@RequestMapping("/agent")
@RequiredArgsConstructor
@Validated
public class AgentController {

    private final AgentService agentService;

    @PostMapping("/register")
    public DeviceResponse register(@Valid @RequestBody AgentRegisterRequest request) {

        Device device = agentService.registerDevice(request);

        return DeviceResponse.builder()
                .id(device.getId())
                .hostname(device.getHostname())
                .ipAddress(device.getIpAddress())
                .os(device.getOs())
                .status(device.getStatus())
                .lastSeenAt(device.getLastSeenAt())
                .build();
    }

    @PostMapping("/metrics")
    public void sendMetrics(@RequestHeader("x-agent-token") String agentToken,
            @Valid @RequestBody MetricRequest request) {
        agentService.saveMetric(request, agentToken);
    }

    @PostMapping("/metrics/batch")
    public void sendMetricsBatch(@RequestHeader("x-agent-token") String agentToken,
            @Valid @RequestBody List<MetricRequest> requests) {
        agentService.saveMetricsBatch(requests, agentToken);
    }

    @PostMapping("/metrics-detail")
    public void sendMetricDetails(@RequestHeader("x-agent-token") String agentToken,
            @Valid @RequestBody MetricDetailRequest request) {
        agentService.saveMetricDetail(request, agentToken);
    }

    @PostMapping("/metrics-detail/batch")
    public void sendMetricDetailsBatch(@RequestHeader("x-agent-token") String agentToken,
            @Valid @RequestBody List<MetricDetailRequest> requests) {
        agentService.saveMetricDetailsBatch(requests, agentToken);
    }
}
