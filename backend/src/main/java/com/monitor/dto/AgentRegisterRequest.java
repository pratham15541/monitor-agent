package com.monitor.dto;

import lombok.Data;

import jakarta.validation.constraints.NotBlank;

@Data
public class AgentRegisterRequest {
    @NotBlank
    private String token;

    @NotBlank
    private String hostname;

    @NotBlank
    private String ipAddress;

    @NotBlank
    private String os;
}
