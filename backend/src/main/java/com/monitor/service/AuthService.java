package com.monitor.service;

import com.monitor.dto.LoginRequest;
import com.monitor.dto.RegisterRequest;
import com.monitor.entity.Company;
import com.monitor.repository.CompanyRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class AuthService {

    private final CompanyRepository companyRepository;
    private final JwtService jwtService;
    private final BCryptPasswordEncoder passwordEncoder;

    public Company register(RegisterRequest request) {

        String apiToken = UUID.randomUUID().toString();

        Company company = Company.builder()
                .name(request.getName())
                .email(request.getEmail())
                .passwordHash(passwordEncoder.encode(request.getPassword()))
                .apiToken(apiToken)
                .createdAt(LocalDateTime.now())
                .build();

        return companyRepository.save(company);
    }

    public String login(LoginRequest request) {

        Company company = companyRepository.findByEmail(request.getEmail())
                .orElseThrow(() -> new RuntimeException("Invalid credentials"));

        if (!passwordEncoder.matches(request.getPassword(), company.getPasswordHash())) {
            throw new RuntimeException("Invalid credentials");
        }

        return jwtService.generateToken(company.getId());
    }

}
