package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

type AAPClient struct {
	client *resty.Client
}

type HostDetails struct {
	Host     string
	Port     int
	Username string
	KeyFile  string
	Password string
}

func NewAAPClient(cfg Config) *AAPClient {
	client := resty.New().
		SetHostURL(cfg.TowerHost).
		SetBasicAuth(cfg.Username, cfg.Password).
		SetHeader("Content-Type", "application/json").
		SetTimeout(cfg.Timeout)
	return &AAPClient{client: client}
}

func getHostDetails() (HostDetails, error) {
	host := os.Getenv("PACKER_SSH_HOST")
	if host == "" {
		return HostDetails{}, errors.New("PACKER_SSH_HOST environment variable is missing")
	}
	portStr := os.Getenv("PACKER_SSH_PORT")
	if portStr == "" {
		return HostDetails{}, errors.New("PACKER_SSH_PORT environment variable is missing")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return HostDetails{}, fmt.Errorf("invalid PACKER_SSH_PORT: %s", err)
	}
	user := os.Getenv("PACKER_SSH_USERNAME")
	if user == "" {
		return HostDetails{}, errors.New("PACKER_SSH_USERNAME environment variable is missing")
	}
	keyFile := os.Getenv("PACKER_SSH_KEY_FILE")
	password := os.Getenv("PACKER_SSH_PASSWORD")
	if keyFile == "" && password == "" {
		return HostDetails{}, errors.New("either PACKER_SSH_KEY_FILE or PACKER_SSH_PASSWORD must be set")
	}
	return HostDetails{
		Host:     host,
		Port:     port,
		Username: user,
		KeyFile:  keyFile,
		Password: password,
	}, nil
}

func (c *AAPClient) CreateInventory(ctx context.Context, orgID int) (int, error) {
	invBody := map[string]interface{}{
		"name":         fmt.Sprintf("packer-inv-%d", time.Now().Unix()),
		"organization": orgID,
	}
	resp, err := c.client.R().SetContext(ctx).SetBody(invBody).Post("/api/v2/inventories/")
	if err != nil {
		return 0, fmt.Errorf("failed to create inventory: %s", err)
	}
	if resp.IsError() {
		return 0, fmt.Errorf("failed to create inventory: %s (status: %d)", resp.String(), resp.StatusCode())
	}
	var result struct{ ID int `json:"id"` }
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return 0, fmt.Errorf("failed to parse inventory response: %s", err)
	}
	return result.ID, nil
}

func (c *AAPClient) CreateHost(ctx context.Context, invID int, details HostDetails) (int, error) {
	vars := fmt.Sprintf(
		"ansible_host: %s\nansible_port: %d\nansible_user: %s\nansible_ssh_private_key_file: %s\n",
		details.Host, details.Port, details.Username, details.KeyFile,
	)
	if details.Password != "" {
		vars += fmt.Sprintf("ansible_password: %s\n", details.Password)
	}
	hostBody := map[string]interface{}{
		"name":      details.Host,
		"inventory": invID,
		"variables": vars,
	}
	resp, err := c.client.R().SetContext(ctx).SetBody(hostBody).Post("/api/v2/hosts/")
	if err != nil {
		return 0, fmt.Errorf("failed to create host: %s", err)
	}
	if resp.IsError() {
		return 0, fmt.Errorf("failed to create host: %s (status: %d)", resp.String(), resp.StatusCode())
	}
	var result struct{ ID int `json:"id"` }
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return 0, fmt.Errorf("failed to parse host response: %s", err)
	}
	return result.ID, nil
}

func (c *AAPClient) DeleteHost(ctx context.Context, hostID int) error {
	resp, err := c.client.R().SetContext(ctx).Delete(fmt.Sprintf("/api/v2/hosts/%d/", hostID))
	if err != nil {
		return fmt.Errorf("failed to delete host: %s", err)
	}
	if resp.IsError() {
		return fmt.Errorf("failed to delete host: %s (status: %d)", resp.String(), resp.StatusCode())
	}
	return nil
}

func (c *AAPClient) DeleteInventory(ctx context.Context, invID int) error {
	resp, err := c.client.R().SetContext(ctx).Delete(fmt.Sprintf("/api/v2/inventories/%d/", invID))
	if err != nil {
		return fmt.Errorf("failed to delete inventory: %s", err)
	}
	if resp.IsError() {
		return fmt.Errorf("failed to delete inventory: %s (status: %d)", resp.String(), resp.StatusCode())
	}
	return nil
}

func (c *AAPClient) LaunchJob(ctx context.Context, invID, jobTemplateID, workflowTemplateID int, extraVars map[string]interface{}) (int, error) {
	launch := map[string]interface{}{
		"inventory":  invID,
		"extra_vars": extraVars,
	}
	var endpoint string
	if workflowTemplateID != 0 {
		launch["workflow_template"] = workflowTemplateID
		endpoint = fmt.Sprintf("/api/v2/workflow_job_templates/%d/launch/", workflowTemplateID)
	} else {
		launch["job_template"] = jobTemplateID
		endpoint = fmt.Sprintf("/api/v2/job_templates/%d/launch/", jobTemplateID)
	}
	resp, err := c.client.R().SetContext(ctx).SetBody(launch).Post(endpoint)
		return 0, fmt.Errorf("failed to launch job: %s", err)
	}
	if resp.IsError() {
		return 0, fmt.Errorf("failed to launch job: %s (status: %d)", resp.String(), resp.StatusCode())
	}
	var result struct{ Job int `json:"job"` }
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return 0, fmt.Errorf("failed to parse job launch response: %s", err)
	}
	return result.Job, nil
}

func (c *AAPClient) PollJob(ctx context.Context, jobID int, timeout, pollInterval time.Duration) error {
	start := time.Now()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollInterval):
			var result struct {
				Status string `json:"status"`
				Failed bool   `json:"failed"`
				Stdout string `json:"stdout"`
			}
			resp, err := c.client.R().SetContext(ctx).SetResult(&result).Get(fmt.Sprintf("/api/v2/jobs/%d/", jobID))
			if err != nil || resp.IsError() {
				return fmt.Errorf("failed to poll job %d: %s", jobID, err)
			}
			if result.Status == "successful" {
				return nil
			}
			if result.Status == "failed" || result.Failed {
				return fmt.Errorf("job %d failed: %s", jobID, result.Stdout)
			}
		}
		if time.Since(start) > timeout {
			return errors.New("timeout waiting for job completion")
		}
	}
}
