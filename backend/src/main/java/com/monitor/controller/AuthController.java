package com.monitor.controller;

import com.monitor.dto.AuthResponse;
import com.monitor.dto.LoginRequest;
import com.monitor.dto.RegisterRequest;
import com.monitor.entity.Company;
import com.monitor.service.AuthService;
import lombok.RequiredArgsConstructor;
import jakarta.validation.Valid;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/auth")
@RequiredArgsConstructor
public class AuthController {

    private final AuthService authService;

    @PostMapping("/register")
    public AuthResponse register(@Valid @RequestBody RegisterRequest request) {
        Company company = authService.register(request);

        return AuthResponse.builder()
                .id(company.getId())
                .name(company.getName())
                .email(company.getEmail())
                .apiToken(company.getApiToken())
                .build();
    }

    @PostMapping("/login")
    public ResponseEntity<Map<String, String>> login(@Valid @RequestBody LoginRequest request) {
        String token = authService.login(request);
        return ResponseEntity.ok(Map.of("token", token));
    }

}
