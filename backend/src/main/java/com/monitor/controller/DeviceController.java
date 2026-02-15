package com.monitor.controller;

import com.monitor.dto.DeviceResponse;
import com.monitor.dto.MetricDetailResponse;
import com.monitor.service.DeviceService;
import lombok.RequiredArgsConstructor;
import com.monitor.entity.Metric;
import org.springframework.security.core.context.SecurityContextHolder;

import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.UUID;

@RestController
@RequestMapping("/devices")
@RequiredArgsConstructor
public class DeviceController {

    private final DeviceService deviceService;

    @GetMapping
    public List<DeviceResponse> getDevices() {
        UUID companyId = (UUID) SecurityContextHolder
                .getContext()
                .getAuthentication()
                .getPrincipal();

        return deviceService.getDevices(companyId);
    }

    @GetMapping("/{deviceId}/metrics")
    public List<Metric> getMetrics(@PathVariable UUID deviceId) {
        return deviceService.getMetrics(deviceId);
    }

    @GetMapping("/{deviceId}/metrics-detail")
    public List<MetricDetailResponse> getDetailedMetrics(@PathVariable UUID deviceId) {
        UUID companyId = (UUID) SecurityContextHolder
                .getContext()
                .getAuthentication()
                .getPrincipal();

        return deviceService.getDetailedMetrics(companyId, deviceId);
    }

}
