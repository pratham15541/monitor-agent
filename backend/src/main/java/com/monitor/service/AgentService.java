package com.monitor.service;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.monitor.dto.AgentRegisterRequest;
import com.monitor.dto.MetricDetailRequest;
import com.monitor.dto.MetricRequest;
import com.monitor.entity.Company;
import com.monitor.entity.Device;
import com.monitor.entity.DeviceStatus;
import com.monitor.entity.Metric;
import com.monitor.entity.MetricDetail;
import com.monitor.repository.CompanyRepository;
import com.monitor.repository.DeviceRepository;
import com.monitor.repository.MetricDetailRepository;
import com.monitor.repository.MetricRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;

@Service
@RequiredArgsConstructor
public class AgentService {

        private final CompanyRepository companyRepository;
        private final DeviceRepository deviceRepository;
        private final MetricRepository metricRepository;
        private final MetricDetailRepository metricDetailRepository;
        private final ObjectMapper objectMapper;
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

        public void saveMetric(MetricRequest request, String agentToken) {
                Company company = getCompanyByAgentToken(agentToken);

                Device device = deviceRepository.findById(request.getDeviceId())
                                .orElseThrow(() -> new RuntimeException("Device not found"));

                if (!device.getCompany().getId().equals(company.getId())) {
                        throw new RuntimeException("Unauthorized device");
                }

                saveMetric(request);
        }

        public void saveMetricDetail(MetricDetailRequest request) {

                Device device = deviceRepository.findById(request.getDeviceId())
                                .orElseThrow(() -> new RuntimeException("Device not found"));

                String detailsJson = "{}";
                if (request.getDetails() != null) {
                        try {
                                detailsJson = objectMapper.writeValueAsString(request.getDetails());
                        } catch (JsonProcessingException e) {
                                throw new RuntimeException("Failed to serialize details", e);
                        }
                }

                MetricDetail detail = MetricDetail.builder()
                                .device(device)
                                .detailsJson(detailsJson)
                                .createdAt(LocalDateTime.now())
                                .build();

                metricDetailRepository.save(detail);
        }

        public void saveMetricDetail(MetricDetailRequest request, String agentToken) {
                Company company = getCompanyByAgentToken(agentToken);
                Device device = deviceRepository.findById(request.getDeviceId())
                                .orElseThrow(() -> new RuntimeException("Device not found"));

                if (!device.getCompany().getId().equals(company.getId())) {
                        throw new RuntimeException("Unauthorized device");
                }

                saveMetricDetail(request);
        }

        private Company getCompanyByAgentToken(String agentToken) {
                if (agentToken == null || agentToken.isBlank()) {
                        throw new RuntimeException("Missing agent token");
                }
                return companyRepository.findByApiToken(agentToken)
                                .orElseThrow(() -> new RuntimeException("Invalid agent token"));
        }

}
