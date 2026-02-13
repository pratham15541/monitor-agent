package com.monitor.service;

import com.monitor.dto.AgentRegisterRequest;
import com.monitor.dto.MetricRequest;
import com.monitor.entity.*;
import com.monitor.repository.*;
import lombok.RequiredArgsConstructor;
import com.monitor.entity.Company;
import com.monitor.entity.Device;
import com.monitor.entity.DeviceStatus;
import com.monitor.entity.Metric;
import com.monitor.repository.CompanyRepository;
import com.monitor.repository.DeviceRepository;
import com.monitor.repository.MetricRepository;
import org.springframework.messaging.simp.SimpMessagingTemplate;

import org.springframework.stereotype.Service;

import java.time.LocalDateTime;

@Service
@RequiredArgsConstructor
public class AgentService {

        private final CompanyRepository companyRepository;
        private final DeviceRepository deviceRepository;
        private final MetricRepository metricRepository;
        private final SimpMessagingTemplate messagingTemplate;

        public Device registerDevice(AgentRegisterRequest request) {

                Company company = companyRepository.findByApiToken(request.getToken())
                                .orElseThrow(() -> new RuntimeException("Invalid token"));

                Device device = deviceRepository
                                .findByHostnameAndCompany(request.getHostname(), company)
                                .orElse(
                                                Device.builder()
                                                                .hostname(request.getHostname())
                                                                .ipAddress(request.getIpAddress())
                                                                .os(request.getOs())
                                                                .company(company)
                                                                .createdAt(LocalDateTime.now())
                                                                .build());

                device.setLastSeenAt(LocalDateTime.now());
                device.setStatus(DeviceStatus.ONLINE);

                return deviceRepository.save(device);
        }

        public void saveMetric(MetricRequest request) {

                Device device = deviceRepository.findById(request.getDeviceId())
                                .orElseThrow(() -> new RuntimeException("Device not found"));

                Metric metric = Metric.builder()
                                .device(device)
                                .cpuUsage(request.getCpuUsage())
                                .memoryUsage(request.getMemoryUsage())
                                .diskUsage(request.getDiskUsage())
                                .networkIn(request.getNetworkIn())
                                .networkOut(request.getNetworkOut())
                                .createdAt(LocalDateTime.now())
                                .build();

                metricRepository.save(metric);

                device.setLastSeenAt(LocalDateTime.now());
                if (device.getStatus() != DeviceStatus.ONLINE) {
                        device.setStatus(DeviceStatus.ONLINE);

                        messagingTemplate.convertAndSend(
                                        "/topic/device-status/" + device.getId(),
                                        DeviceStatus.ONLINE);
                }
                deviceRepository.save(device);

                // 🔴 ADD THIS FOR LIVE STREAMING
                messagingTemplate.convertAndSend(
                                "/topic/device/" + device.getId(),
                                metric);
        }

}
