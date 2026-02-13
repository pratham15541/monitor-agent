package com.monitor.service;

import com.monitor.dto.DeviceResponse;
import com.monitor.entity.Company;
import com.monitor.entity.Device;
import com.monitor.repository.CompanyRepository;
import com.monitor.repository.DeviceRepository;
import lombok.RequiredArgsConstructor;
import com.monitor.entity.Metric;
import com.monitor.repository.MetricRepository;

import org.springframework.stereotype.Service;

import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
public class DeviceService {

        private final CompanyRepository companyRepository;
        private final DeviceRepository deviceRepository;
        private final MetricRepository metricRepository;

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
}
