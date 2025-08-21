package config_test

import (
	"testing"
	"time"

	"github.com/rptcloud/packer-provisioner-ansible-aap/pkgs/config"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  config.Config
		wantErr bool
	}{
		{
			name: "valid config with username/password",
			config: config.Config{
				TowerHost:      "https://aap.example.com",
				Username:       "admin",
				Password:       "secret",
				JobTemplateID:  42,
				OrganizationID: 1,
			},
			wantErr: false,
		},
		{
			name: "valid config with access token",
			config: config.Config{
				TowerHost:      "https://aap.example.com",
				AccessToken:    "token123",
				JobTemplateID:  42,
				OrganizationID: 1,
			},
			wantErr: false,
		},
		{
			name: "missing tower_host",
			config: config.Config{
				Username:       "admin",
				Password:       "secret",
				JobTemplateID:  42,
				OrganizationID: 1,
			},
			wantErr: true,
		},
		{
			name: "invalid tower_host scheme",
			config: config.Config{
				TowerHost:      "ftp://aap.example.com",
				Username:       "admin",
				Password:       "secret",
				JobTemplateID:  42,
				OrganizationID: 1,
			},
			wantErr: true,
		},
		{
			name: "missing authentication",
			config: config.Config{
				TowerHost:      "https://aap.example.com",
				JobTemplateID:  42,
				OrganizationID: 1,
			},
			wantErr: true,
		},
		{
			name: "missing username with password",
			config: config.Config{
				TowerHost:      "https://aap.example.com",
				Password:       "secret",
				JobTemplateID:  42,
				OrganizationID: 1,
			},
			wantErr: true,
		},
		{
			name: "missing password with username",
			config: config.Config{
				TowerHost:      "https://aap.example.com",
				Username:       "admin",
				JobTemplateID:  42,
				OrganizationID: 1,
			},
			wantErr: true,
		},
		{
			name: "missing job template and workflow template",
			config: config.Config{
				TowerHost:      "https://aap.example.com",
				Username:       "admin",
				Password:       "secret",
				OrganizationID: 1,
			},
			wantErr: true,
		},
		{
			name: "missing organization_id with dynamic inventory",
			config: config.Config{
				TowerHost:        "https://aap.example.com",
				Username:         "admin",
				Password:         "secret",
				JobTemplateID:    42,
				DynamicInventory: true,
			},
			wantErr: true,
		},
		{
			name: "valid workflow template",
			config: config.Config{
				TowerHost:          "https://aap.example.com",
				Username:           "admin",
				Password:           "secret",
				WorkflowTemplateID: 42,
				OrganizationID:     1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Validate_Defaults(t *testing.T) {
	config := config.Config{
		TowerHost:      "https://aap.example.com",
		Username:       "admin",
		Password:       "secret",
		JobTemplateID:  42,
		OrganizationID: 1,
	}

	err := config.Validate()
	if err != nil {
		t.Fatalf("Config.Validate() failed: %v", err)
	}

	// Check default values
	if config.Timeout != 15*time.Minute {
		t.Errorf("Expected default timeout to be 15m, got %v", config.Timeout)
	}

	if config.PollInterval != 10*time.Second {
		t.Errorf("Expected default poll interval to be 10s, got %v", config.PollInterval)
	}

	if config.ExtraVars == nil {
		t.Error("Expected ExtraVars to be initialized as empty map")
	}
}

func TestConfig_Validate_CustomDefaults(t *testing.T) {
	config := config.Config{
		TowerHost:      "https://aap.example.com",
		Username:       "admin",
		Password:       "secret",
		JobTemplateID:  42,
		OrganizationID: 1,
		Timeout:        30 * time.Minute,
		PollInterval:   5 * time.Second,
		ExtraVars: map[string]interface{}{
			"custom": "value",
		},
	}

	err := config.Validate()
	if err != nil {
		t.Fatalf("Config.Validate() failed: %v", err)
	}

	// Check that custom values are preserved
	if config.Timeout != 30*time.Minute {
		t.Errorf("Expected custom timeout to be 30m, got %v", config.Timeout)
	}

	if config.PollInterval != 5*time.Second {
		t.Errorf("Expected custom poll interval to be 5s, got %v", config.PollInterval)
	}

	if config.ExtraVars["custom"] != "value" {
		t.Errorf("Expected custom extra vars to be preserved")
	}
}
