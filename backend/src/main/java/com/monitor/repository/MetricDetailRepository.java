package com.monitor.repository;

import com.monitor.entity.Device;
import com.monitor.entity.MetricDetail;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;

public interface MetricDetailRepository extends JpaRepository<MetricDetail, Long> {
    List<MetricDetail> findTop20ByDeviceOrderByCreatedAtDesc(Device device);
}
