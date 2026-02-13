package com.monitor.dto;

import lombok.Data;

@Data
public class AgentRegisterRequest {
    private String token;
    private String hostname;
    private String ipAddress;
    private String os;
}
