package company

import (
	"context"
	"fmt"
	"strings"

	"loadtester/internal/api"
	"loadtester/internal/config"
)

type Record struct {
	ID        string `json:"companyId"`
	Name      string `json:"companyName"`
	AdminName string `json:"adminName"`
	Email     string `json:"adminEmail"`
	Password  string `json:"password"`
	APIToken  string `json:"apiToken"`
	JWT       string `json:"jwt,omitempty"`
	Systems   int    `json:"systems"`
	Connected int    `json:"connected"`
}

type Manager struct {
	client *api.Client
}

func NewManager(client *api.Client) *Manager {
	return &Manager{client: client}
}

func (m *Manager) CreateAll(ctx context.Context, specs []config.CompanyConfig) ([]Record, error) {
	records := make([]Record, 0, len(specs))
	for _, spec := range specs {
		record, err := m.Create(ctx, spec)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (m *Manager) Create(ctx context.Context, spec config.CompanyConfig) (Record, error) {
	created, _, registerErr := m.client.RegisterCompany(ctx, api.AuthRequest{Name: spec.Name, Email: spec.AdminEmail, Password: spec.Password})
	reused := false
	if registerErr != nil {
		reused = true
	}

	token, _, err := m.client.LoginCompany(ctx, api.LoginRequest{Email: spec.AdminEmail, Password: spec.Password})
	if err != nil {
		if registerErr != nil {
			return Record{}, fmt.Errorf("register company %q: %w; login existing company also failed: %w", spec.Name, registerErr, err)
		}
		return Record{}, fmt.Errorf("login company %q: %w", spec.Name, err)
	}
	profile, _, err := m.client.GetCompanyMe(ctx, token)
	if err != nil {
		return Record{}, fmt.Errorf("profile company %q: %w", spec.Name, err)
	}
	if created.ID == "" {
		created = profile
	}

	fmt.Println()
	fmt.Println("==========================================")
	fmt.Println("GENERATED TEST DATA")
	fmt.Println("==========================================")
	fmt.Printf("Company ID: %s\n", created.ID)
	fmt.Printf("Company: %s\n", profile.Name)
	if reused {
		fmt.Println("Mode: reused existing company")
	}
	fmt.Printf("Login Email: %s\n", profile.Email)
	fmt.Printf("Password: %s\n", spec.Password)
	fmt.Printf("API Token: %s\n", profile.APIToken)
	fmt.Printf("JWT Token: %s\n", token)
	fmt.Println("Use these credentials in the dashboard while the load test is running.")
	fmt.Println("----------------------------------------")

	return Record{
		ID:        created.ID,
		Name:      strings.TrimSpace(profile.Name),
		AdminName: spec.AdminName,
		Email:     profile.Email,
		Password:  spec.Password,
		APIToken:  profile.APIToken,
		JWT:       token,
		Systems:   spec.Systems,
	}, nil
}
