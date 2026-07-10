package com.monitor.service;

import com.monitor.dto.LoginRequest;
import com.monitor.dto.RegisterRequest;
import com.monitor.entity.Company;
import com.monitor.repository.CompanyRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.security.authentication.BadCredentialsException;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.stereotype.Service;
import org.springframework.web.server.ResponseStatusException;

import java.time.Instant;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class AuthService {

    private final CompanyRepository companyRepository;
    private final JwtService jwtService;
    private final BCryptPasswordEncoder passwordEncoder;

    public Company register(RegisterRequest request) {
        if (companyRepository.findByEmail(request.getEmail()).isPresent()) {
            throw new ResponseStatusException(HttpStatus.CONFLICT, "Email already registered");
        }

        String apiToken = UUID.randomUUID().toString();

        Company company = Company.builder()
                .name(request.getName())
                .email(request.getEmail())
                .passwordHash(passwordEncoder.encode(request.getPassword()))
                .apiToken(apiToken)
                .createdAt(Instant.now())
                .build();

        return companyRepository.save(company);
    }

    public String login(LoginRequest request) {

        Company company = companyRepository.findByEmail(request.getEmail())
                .orElseThrow(() -> new BadCredentialsException("Invalid credentials"));

        if (!passwordEncoder.matches(request.getPassword(), company.getPasswordHash())) {
            throw new BadCredentialsException("Invalid credentials");
        }

        return jwtService.generateToken(company.getId());
    }

}
