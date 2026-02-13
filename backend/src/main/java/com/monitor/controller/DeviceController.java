package com.monitor.controller;

import com.monitor.dto.DeviceResponse;
import com.monitor.entity.Company;
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
System.out.println("JWT FILTER EXECUTED");
        UUID companyId = (UUID) SecurityContextHolder
                .getContext()
                .getAuthentication()
                .getPrincipal();
        
        System.out.println("Authenticated companyId: " + companyId);

        return deviceService.getDevices(companyId);
    }

    @GetMapping("/{deviceId}/metrics")
    public List<Metric> getMetrics(@PathVariable UUID deviceId) {
        return deviceService.getMetrics(deviceId);
    }

}
