//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package main

import (
	"errors"
	"strings"
	"time"
)

type Config struct {
	TowerHost         string                 `mapstructure:"tower_host"`
	Username          string                 `mapstructure:"username"`
	Password          string                 `mapstructure:"password"`
	JobTemplateID     int                    `mapstructure:"job_template_id"`
	InventoryID       int                    `mapstructure:"inventory_id"`
	OrganizationID    int                    `mapstructure:"organization_id"`
	DynamicInventory  bool                   `mapstructure:"dynamic_inventory"`
	KeepTempInventory bool                   `mapstructure:"keep_temp_inventory"`
	ExtraVars         map[string]interface{} `mapstructure:"extra_vars"`
	Timeout           time.Duration          `mapstructure:"timeout"`
	PollInterval      time.Duration          `mapstructure:"poll_interval"`
	WorkflowTemplateID int                   `mapstructure:"workflow_template_id"`
}

func (c *Config) Validate() error {
	if c.TowerHost == "" {
		return errors.New("tower_host must be set")
	}
	if !strings.HasPrefix(c.TowerHost, "http://") && !strings.HasPrefix(c.TowerHost, "https://") {
		return errors.New("tower_host must start with http:// or https://")
	}
	if c.Username == "" {
		return errors.New("username must be set")
	}
	if c.Password == "" {
		return errors.New("password must be set")
	}
	if c.JobTemplateID == 0 && c.WorkflowTemplateID == 0 {
        return errors.New("either job_template_id or workflow_template_id must be set")
    }
	if c.DynamicInventory && c.OrganizationID == 0 {
		return errors.New("organization_id must be set when dynamic_inventory is true")
	}
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Minute
	}
	if c.PollInterval == 0 {
		c.PollInterval = 5 * time.Second
	}
	if c.ExtraVars == nil {
		c.ExtraVars = make(map[string]interface{})
	}
	return nil
}
