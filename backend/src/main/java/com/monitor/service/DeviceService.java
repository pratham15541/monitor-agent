package com.monitor.service;

import com.monitor.dto.DeviceResponse;
import com.monitor.dto.MetricDetailResponse;
import com.monitor.entity.Company;
import com.monitor.entity.Device;
import com.monitor.entity.MetricDetail;
import com.monitor.repository.CompanyRepository;
import com.monitor.repository.DeviceRepository;
import lombok.RequiredArgsConstructor;
import com.monitor.entity.Metric;
import com.monitor.repository.MetricDetailRepository;
import com.monitor.repository.MetricRepository;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
public class DeviceService {

        private static final long DETAIL_CACHE_TTL_MS = 5000;

        private final CompanyRepository companyRepository;
        private final DeviceRepository deviceRepository;
        private final MetricRepository metricRepository;
        private final MetricDetailRepository metricDetailRepository;
        private final Map<UUID, CachedDetails> detailedMetricsCache = new ConcurrentHashMap<>();

        public List<DeviceResponse> getDevices(UUID companyId) {

                Company company = companyRepository.findById(companyId)
                                .orElseThrow();

                return deviceRepository.findByCompany(company)
                                .stream()
                                .map(device -> DeviceResponse.builder()
                                                .id(device.getId())
                                                .hostname(device.getHostname())
                                                .ipAddress(device.getIpAddress())
                                                .os(device.getOs())
                                                .status(device.getStatus())
                                                .lastSeenAt(device.getLastSeenAt())
                                                .build())
                                .collect(Collectors.toList());
        }

        public List<Metric> getMetrics(UUID deviceId) {

                Device device = deviceRepository.findById(deviceId)
                                .orElseThrow();

                return metricRepository.findTop50ByDeviceOrderByCreatedAtDesc(device);
        }

        @Transactional(readOnly = true)
        public List<MetricDetailResponse> getDetailedMetrics(UUID companyId, UUID deviceId) {
                Company company = companyRepository.findById(companyId)
                                .orElseThrow(() -> new RuntimeException("Company not found"));

                Device device = deviceRepository.findById(deviceId)
                                .orElseThrow(() -> new RuntimeException("Device not found"));

                if (!device.getCompany().getId().equals(company.getId())) {
                        throw new RuntimeException("Unauthorized");
                }

                CachedDetails cached = detailedMetricsCache.get(deviceId);
                long now = System.currentTimeMillis();
                if (cached != null && now - cached.cachedAt < DETAIL_CACHE_TTL_MS) {
                        return cached.data;
                }

                List<MetricDetail> details = metricDetailRepository
                                .findTop20ByDeviceOrderByCreatedAtDesc(device);

                List<MetricDetailResponse> response = details.stream()
                                .map(detail -> MetricDetailResponse.builder()
                                                .id(detail.getId())
                                                .detailsJson(detail.getDetailsJson())
                                                .createdAt(detail.getCreatedAt())
                                                .build())
                                .collect(Collectors.toList());

                detailedMetricsCache.put(deviceId, new CachedDetails(now, response));
                return response;
        }

        private static final class CachedDetails {
                private final long cachedAt;
                private final List<MetricDetailResponse> data;

                private CachedDetails(long cachedAt, List<MetricDetailResponse> data) {
                        this.cachedAt = cachedAt;
                        this.data = data;
                }
        }
}
