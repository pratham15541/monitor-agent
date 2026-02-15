package com.monitor.service;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.monitor.dto.AgentRegisterRequest;
import com.monitor.dto.MetricDetailRequest;
import com.monitor.dto.MetricDetailResponse;
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
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

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

        @Transactional
        public void saveMetricsBatch(List<MetricRequest> requests) {
                if (requests == null || requests.isEmpty()) {
                        return;
                }

                UUID deviceId = requests.get(0).getDeviceId();
                if (deviceId == null) {
                        return;
                }

                Device device = deviceRepository.findById(deviceId)
                                .orElseThrow(() -> new RuntimeException("Device not found"));

                List<Metric> metrics = buildMetrics(device, requests);
                if (metrics.isEmpty()) {
                        return;
                }

                metricRepository.saveAll(metrics);

                LocalDateTime now = metrics.get(metrics.size() - 1).getCreatedAt();
                device.setLastSeenAt(now);
                if (device.getStatus() != DeviceStatus.ONLINE) {
                        device.setStatus(DeviceStatus.ONLINE);

                        messagingTemplate.convertAndSend(
                                        "/topic/device-status/" + device.getId(),
                                        DeviceStatus.ONLINE);
                }
                deviceRepository.save(device);

                Metric latestMetric = metrics.get(metrics.size() - 1);
                messagingTemplate.convertAndSend(
                                "/topic/device/" + device.getId(),
                                latestMetric);
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

        @Transactional
        public void saveMetricsBatch(List<MetricRequest> requests, String agentToken) {
                if (requests == null || requests.isEmpty()) {
                        return;
                }

                Company company = getCompanyByAgentToken(agentToken);
                UUID deviceId = requests.get(0).getDeviceId();
                if (deviceId == null) {
                        return;
                }

                Device device = deviceRepository.findById(deviceId)
                                .orElseThrow(() -> new RuntimeException("Device not found"));

                if (!device.getCompany().getId().equals(company.getId())) {
                        throw new RuntimeException("Unauthorized device");
                }

                saveMetricsBatch(requests);
        }

        @Transactional
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

                MetricDetail saved = metricDetailRepository.save(detail);

                MetricDetailResponse response = MetricDetailResponse.builder()
                                .id(saved.getId())
                                .detailsJson(saved.getDetailsJson())
                                .createdAt(saved.getCreatedAt())
                                .build();

                messagingTemplate.convertAndSend(
                                "/topic/device-detail/" + device.getId(),
                                response);
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

        @Transactional
        public void saveMetricDetailsBatch(List<MetricDetailRequest> requests) {
                if (requests == null || requests.isEmpty()) {
                        return;
                }

                UUID deviceId = requests.get(0).getDeviceId();
                if (deviceId == null) {
                        return;
                }

                Device device = deviceRepository.findById(deviceId)
                                .orElseThrow(() -> new RuntimeException("Device not found"));

                List<MetricDetail> details = buildMetricDetails(device, requests);
                if (details.isEmpty()) {
                        return;
                }

                List<MetricDetail> saved = metricDetailRepository.saveAll(details);
                MetricDetail latest = saved.get(saved.size() - 1);
                MetricDetailResponse response = MetricDetailResponse.builder()
                                .id(latest.getId())
                                .detailsJson(latest.getDetailsJson())
                                .createdAt(latest.getCreatedAt())
                                .build();

                messagingTemplate.convertAndSend(
                                "/topic/device-detail/" + device.getId(),
                                response);
        }

        @Transactional
        public void saveMetricDetailsBatch(List<MetricDetailRequest> requests, String agentToken) {
                if (requests == null || requests.isEmpty()) {
                        return;
                }

                Company company = getCompanyByAgentToken(agentToken);
                UUID deviceId = requests.get(0).getDeviceId();
                if (deviceId == null) {
                        return;
                }

                Device device = deviceRepository.findById(deviceId)
                                .orElseThrow(() -> new RuntimeException("Device not found"));

                if (!device.getCompany().getId().equals(company.getId())) {
                        throw new RuntimeException("Unauthorized device");
                }

                saveMetricDetailsBatch(requests);
        }

        private Company getCompanyByAgentToken(String agentToken) {
                if (agentToken == null || agentToken.isBlank()) {
                        throw new RuntimeException("Missing agent token");
                }
                return companyRepository.findByApiToken(agentToken)
                                .orElseThrow(() -> new RuntimeException("Invalid agent token"));
        }

        private List<Metric> buildMetrics(Device device, List<MetricRequest> requests) {
                List<Metric> metrics = new ArrayList<>(requests.size());

                for (MetricRequest request : requests) {
                        if (request == null || request.getDeviceId() == null) {
                                continue;
                        }
                        if (!request.getDeviceId().equals(device.getId())) {
                                continue;
                        }

                        metrics.add(Metric.builder()
                                        .device(device)
                                        .cpuUsage(request.getCpuUsage())
                                        .memoryUsage(request.getMemoryUsage())
                                        .diskUsage(request.getDiskUsage())
                                        .networkIn(request.getNetworkIn())
                                        .networkOut(request.getNetworkOut())
                                        .createdAt(LocalDateTime.now())
                                        .build());
                }

                return metrics;
        }

        private List<MetricDetail> buildMetricDetails(Device device, List<MetricDetailRequest> requests) {
                List<MetricDetail> details = new ArrayList<>(requests.size());

                for (MetricDetailRequest request : requests) {
                        if (request == null || request.getDeviceId() == null) {
                                continue;
                        }
                        if (!request.getDeviceId().equals(device.getId())) {
                                continue;
                        }

                        String detailsJson = "{}";
                        if (request.getDetails() != null) {
                                try {
                                        detailsJson = objectMapper.writeValueAsString(request.getDetails());
                                } catch (JsonProcessingException e) {
                                        throw new RuntimeException("Failed to serialize details", e);
                                }
                        }

                        details.add(MetricDetail.builder()
                                        .device(device)
                                        .detailsJson(detailsJson)
                                        .createdAt(LocalDateTime.now())
                                        .build());
                }

                return details;
        }

}
