package com.monitor.controller;

import com.monitor.dto.AuthResponse;
import com.monitor.dto.LoginRequest;
import com.monitor.dto.RegisterRequest;
import com.monitor.entity.Company;
import com.monitor.service.AuthService;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/auth")
@RequiredArgsConstructor
public class AuthController {

    private final AuthService authService;

   @PostMapping("/register")
public AuthResponse register(@RequestBody RegisterRequest request) {
    Company company = authService.register(request);

    return AuthResponse.builder()
            .id(company.getId())
            .name(company.getName())
            .email(company.getEmail())
            .apiToken(company.getApiToken())
            .build();
}


    @PostMapping("/login")
public String login(@RequestBody LoginRequest request) {
    return authService.login(request);
}

}
