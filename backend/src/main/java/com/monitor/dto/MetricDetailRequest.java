package com.monitor.dto;

import lombok.Data;

import jakarta.validation.constraints.NotNull;
import java.util.Map;
import java.util.UUID;

@Data
public class MetricDetailRequest {

    @NotNull
    private UUID deviceId;

    @NotNull
    private Map<String, Object> details;
}
