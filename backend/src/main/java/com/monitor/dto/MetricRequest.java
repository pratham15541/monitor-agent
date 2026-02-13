package com.monitor.dto;

import lombok.Data;

import java.util.UUID;

@Data
public class MetricRequest {

    private UUID deviceId;

    private double cpuUsage;
    private double memoryUsage;
    private double diskUsage;
    private double networkIn;
    private double networkOut;
}
