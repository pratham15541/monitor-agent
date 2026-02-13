package com.monitor.service;

import com.monitor.entity.Device;
import com.monitor.entity.DeviceStatus;
import com.monitor.repository.DeviceRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

import java.time.LocalDateTime;
import java.util.List;

@Component
@RequiredArgsConstructor
public class DeviceStatusScheduler {

    private final DeviceRepository deviceRepository;
    private final SimpMessagingTemplate messagingTemplate;

    // Runs every 30 seconds
    @Scheduled(fixedRate = 30000)
    public void checkOfflineDevices() {

        List<Device> devices = deviceRepository.findAll();

        for (Device device : devices) {

            if (device.getLastSeenAt() == null)
                continue;

            boolean shouldBeOffline = device.getLastSeenAt().isBefore(LocalDateTime.now().minusSeconds(30));

            if (shouldBeOffline && device.getStatus() != DeviceStatus.OFFLINE) {

                device.setStatus(DeviceStatus.OFFLINE);
                deviceRepository.save(device);

                System.out.println("Device marked OFFLINE: " + device.getHostname());

                // 🔴 Broadcast status change
                messagingTemplate.convertAndSend(
                        "/topic/device-status/" + device.getId(),
                        device.getStatus());
            }
        }
    }
}
