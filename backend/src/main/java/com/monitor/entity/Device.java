package com.monitor.entity;

import jakarta.persistence.*;
import lombok.*;

import java.time.Instant;
import java.util.UUID;

@Entity
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class Device {

    @Id
    @GeneratedValue(strategy = GenerationType.UUID)
    private UUID id;

    private String hostname;
    private String ipAddress;
    private String os;

    @Enumerated(EnumType.STRING)
    private DeviceStatus status;

    private Instant lastSeenAt;
    private Instant createdAt;

    @ManyToOne
    @JoinColumn(name = "company_id")
    private Company company;
}
