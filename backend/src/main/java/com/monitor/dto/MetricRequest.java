package com.monitor.dto;

import lombok.Data;

import jakarta.validation.constraints.Max;
import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.PositiveOrZero;
import java.util.UUID;

@Data
public class MetricRequest {

    @NotNull
    private UUID deviceId;

    @Min(0)
    @Max(100)
    private double cpuUsage;

    @Min(0)
    @Max(100)
    private double memoryUsage;

    @Min(0)
    @Max(100)
    private double diskUsage;

    @PositiveOrZero
    private double networkIn;

    @PositiveOrZero
    private double networkOut;
}
