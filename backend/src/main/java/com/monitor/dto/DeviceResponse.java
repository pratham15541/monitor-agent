package com.monitor.dto;

import com.monitor.entity.DeviceStatus;
import lombok.Builder;
import lombok.Data;

import java.time.LocalDateTime;
import java.util.UUID;

@Data
@Builder
public class DeviceResponse {
    private UUID id;
    private String hostname;
    private String ipAddress;
    private String os;
    private DeviceStatus status;
    private LocalDateTime lastSeenAt;
}
