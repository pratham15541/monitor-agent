package com.monitor.repository;

import com.monitor.entity.Metric;
import com.monitor.entity.Device;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;

public interface MetricRepository extends JpaRepository<Metric, Long> {
    List<Metric> findTop50ByDeviceOrderByCreatedAtDesc(Device device);
}
