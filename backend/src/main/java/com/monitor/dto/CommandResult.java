package com.monitor.dto;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.UUID;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CommandResult {
    private UUID deviceId;
    private String commandId;
    private String type;
    private String status;
    private String output;
    private String error;
    private String startedAt;
    private String finishedAt;
}
