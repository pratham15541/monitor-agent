package com.monitor.controller;

import com.monitor.dto.AuthResponse;
import com.monitor.entity.Company;
import com.monitor.repository.CompanyRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.UUID;

@RestController
@RequestMapping("/company")
@RequiredArgsConstructor
public class CompanyController {

    private final CompanyRepository companyRepository;

    @GetMapping("/me")
    public AuthResponse getProfile() {
        Object principal = SecurityContextHolder.getContext().getAuthentication().getPrincipal();
        UUID companyId = principal instanceof UUID ? (UUID) principal : UUID.fromString(principal.toString());

        Company company = companyRepository.findById(companyId)
                .orElseThrow(() -> new RuntimeException("Company not found"));

        return AuthResponse.builder()
                .id(company.getId())
                .name(company.getName())
                .email(company.getEmail())
                .apiToken(company.getApiToken())
                .build();
    }
}
