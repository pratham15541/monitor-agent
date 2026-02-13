package com.monitor.repository;

import com.monitor.entity.Device;
import com.monitor.entity.Company;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

public interface DeviceRepository extends JpaRepository<Device, UUID> {
    List<Device> findByCompany(Company company);
    Optional<Device> findByHostnameAndCompany(String hostname, Company company);
}
