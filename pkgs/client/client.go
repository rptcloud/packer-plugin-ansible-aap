package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/rptcloud/packer-provisioner-ansible-aap/pkgs/config"
)

type AAPClient struct {
	client *resty.Client
}

type HostDetails struct {
	Host     string
	Port     int
	Username string
	Password string
}

func NewAAPClient(cfg config.Config) *AAPClient {
	client := resty.New().
		SetBaseURL(cfg.TowerHost).
		SetHeader("Content-Type", "application/json").
		SetTimeout(cfg.Timeout)

	if cfg.InsecureSkipVerify {
		client.SetTLSClientConfig(&tls.Config{
			InsecureSkipVerify: true,
		})
	}

	if cfg.AccessToken != "" {

		client.SetHeader("Authorization", fmt.Sprintf("Bearer %s", cfg.AccessToken))
	} else {

		client.SetBasicAuth(cfg.Username, cfg.Password)
	}

	return &AAPClient{client: client}
}

func (c *AAPClient) CreateInventory(ctx context.Context, orgID int) (int, error) {
	invBody := map[string]interface{}{
		"name":         fmt.Sprintf("packer-inv-%d", time.Now().Unix()),
		"description":  "Temporary inventory for packer provisioning",
		"organization": orgID,
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(invBody).
		Post("/api/controller/v2/inventories/")

	if err != nil {
		return 0, fmt.Errorf("failed to create inventory: %s", err)
	}

	if resp.IsError() {
		return 0, fmt.Errorf("failed to create inventory: %s (status: %d)", resp.String(), resp.StatusCode())
	}

	var result struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return 0, fmt.Errorf("failed to parse inventory response: %s", err)
	}

	if result.ID == 0 {
		return 0, fmt.Errorf("failed to create inventory. Response: %s", resp.String())
	}

	return result.ID, nil
}

func (c *AAPClient) CreateHost(ctx context.Context, invID int, details HostDetails, credentialType string) (int, error) {
	hostVars := map[string]interface{}{
		"ansible_host": details.Host,
		"ansible_port": details.Port,
		"ansible_user": details.Username,
	}

	// Set connection type based on port or credential type
	if details.Port == 5985 || details.Port == 5986 || credentialType == "winrm_password" {
		// Windows host - use WinRM
		hostVars["ansible_connection"] = "winrm"
		hostVars["ansible_winrm_server_cert_validation"] = "ignore"
		hostVars["ansible_winrm_transport"] = "basic"
		hostVars["ansible_become_method"] = "runas"
		hostVars["ansible_become"] = "yes"
		hostVars["ansible_become_user"] = "Administrator"

		switch details.Port {
		case 5985:
			hostVars["ansible_winrm_scheme"] = "http"
		case 5986:
			hostVars["ansible_winrm_scheme"] = "https"
		}

	} else {
		// Linux host - use SSH
		hostVars["ansible_connection"] = "ssh"
	}

	// Note: We don't set ansible_ssh_private_key_file here because we're using AAP credentials
	// The credential will handle the SSH authentication

	hostVarsJSON, err := json.Marshal(hostVars)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal host variables: %s", err)
	}

	hostBody := map[string]interface{}{
		"name":      details.Host,
		"inventory": invID,
		"variables": string(hostVarsJSON),
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(hostBody).
		Post("/api/controller/v2/hosts/")

	if err != nil {
		return 0, fmt.Errorf("failed to create host: %s", err)
	}
	if resp.IsError() {
		return 0, fmt.Errorf("failed to create host: %s (status: %d)", resp.String(), resp.StatusCode())
	}

	var result struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return 0, fmt.Errorf("failed to parse host response: %s", err)
	}
	return result.ID, nil
}

func (c *AAPClient) DeleteHost(ctx context.Context, hostID int) error {
	resp, err := c.client.R().
		SetContext(ctx).
		Delete(fmt.Sprintf("/api/controller/v2/hosts/%d/", hostID))

	if err != nil {
		return fmt.Errorf("failed to delete host: %s", err)
	}
	if resp.IsError() {
		return fmt.Errorf("failed to delete host: %s (status: %d)", resp.String(), resp.StatusCode())
	}
	return nil
}

func (c *AAPClient) DeleteInventory(ctx context.Context, invID int) error {
	resp, err := c.client.R().
		SetContext(ctx).
		Delete(fmt.Sprintf("/api/controller/v2/inventories/%d/", invID))

	if err != nil {
		return fmt.Errorf("failed to delete inventory: %s", err)
	}
	if resp.IsError() {
		return fmt.Errorf("failed to delete inventory: %s (status: %d)", resp.String(), resp.StatusCode())
	}
	return nil
}

func (c *AAPClient) GetCredentialTypes(ctx context.Context) ([]struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}, error) {
	// Fetch all pages of credential types
	var allResults []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	url := "/api/controller/v2/credential_types/?page_size=200"
	for url != "" {
		resp, err := c.client.R().
			SetContext(ctx).
			Get(url)

		if err != nil {
			return nil, fmt.Errorf("failed to fetch credential types: %s", err)
		}
		if resp.IsError() {
			return nil, fmt.Errorf("failed to fetch credential types: %s (status: %d)", resp.String(), resp.StatusCode())
		}

		var result struct {
			Results []struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"results"`
			Next *string `json:"next"`
		}
		if err := json.Unmarshal(resp.Body(), &result); err != nil {
			return nil, fmt.Errorf("failed to parse credential types response: %s", err)
		}

		allResults = append(allResults, result.Results...)

		if result.Next != nil && *result.Next != "" {
			// Extract just the path from the next URL since it might be a full URL
			if strings.HasPrefix(*result.Next, "http") {
				// Parse the URL to get just the path and query
				parts := strings.SplitN(*result.Next, "/api/", 2)
				if len(parts) == 2 {
					url = "/api/" + parts[1]
				} else {
					url = ""
				}
			} else {
				url = *result.Next
			}
		} else {
			url = ""
		}
	}

	return allResults, nil
}

func (c *AAPClient) GetCredentialTypeID(ctx context.Context, name string) (int, error) {
	credTypes, err := c.GetCredentialTypes(ctx)
	if err != nil {
		return 0, err
	}

	for _, ct := range credTypes {
		if ct.Name == name {
			return ct.ID, nil
		}
	}

	// Build available types list for error message
	var available []string
	for _, ct := range credTypes {
		available = append(available, fmt.Sprintf("%s (ID: %d)", ct.Name, ct.ID))
	}

	return 0, fmt.Errorf("credential type '%s' not found. Available types: %s", name, strings.Join(available, ", "))
}

func (c *AAPClient) CreateCredential(ctx context.Context, orgID int, username, privateKeyData string) (int, error) {
	credentialTypeID, err := c.GetCredentialTypeID(ctx, "Machine")
	if err != nil {
		return 0, fmt.Errorf("failed to get Machine credential type ID: %s", err)
	}

	credentialBody := map[string]interface{}{
		"name":            fmt.Sprintf("packer-ssh-cred-%d", time.Now().Unix()),
		"description":     "SSH credential for Packer builds",
		"credential_type": credentialTypeID,
		"organization":    orgID,
		"inputs": map[string]interface{}{
			"username":     username,
			"ssh_key_data": privateKeyData,
		},
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(credentialBody).
		Post("/api/controller/v2/credentials/")

	if err != nil {
		return 0, fmt.Errorf("failed to create credential: %s", err)
	}
	if resp.IsError() {
		return 0, fmt.Errorf("failed to create credential: %s (status: %d)", resp.String(), resp.StatusCode())
	}

	var result struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return 0, fmt.Errorf("failed to parse credential response: %s", err)
	}

	if result.ID == 0 {
		return 0, fmt.Errorf("failed to create credential. Response: %s", resp.String())
	}

	return result.ID, nil
}

func (c *AAPClient) DeleteCredential(ctx context.Context, credentialID int) error {
	resp, err := c.client.R().
		SetContext(ctx).
		Delete(fmt.Sprintf("/api/controller/v2/credentials/%d/", credentialID))

	if err != nil {
		return fmt.Errorf("failed to delete credential: %s", err)
	}
	if resp.IsError() {
		return fmt.Errorf("failed to delete credential: %s (status: %d)", resp.String(), resp.StatusCode())
	}
	return nil
}

func (c *AAPClient) CreatePasswordCredential(ctx context.Context, orgID int, username, password string) (int, error) {
	credentialTypeID, err := c.GetCredentialTypeID(ctx, "Machine")
	if err != nil {
		return 0, fmt.Errorf("failed to get Machine credential type ID: %s", err)
	}

	credentialBody := map[string]interface{}{
		"name":            fmt.Sprintf("packer-password-cred-%d", time.Now().Unix()),
		"description":     "SSH password credential for Packer builds",
		"credential_type": credentialTypeID,
		"organization":    orgID,
		"inputs": map[string]interface{}{
			"username": username,
			"password": password,
		},
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(credentialBody).
		Post("/api/controller/v2/credentials/")

	if err != nil {
		return 0, fmt.Errorf("failed to create password credential: %s", err)
	}
	if resp.IsError() {
		return 0, fmt.Errorf("failed to create password credential: %s (status: %d)", resp.String(), resp.StatusCode())
	}

	var result struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return 0, fmt.Errorf("failed to parse password credential response: %s", err)
	}

	if result.ID == 0 {
		return 0, fmt.Errorf("failed to create password credential. Response: %s", resp.String())
	}

	return result.ID, nil
}

func (c *AAPClient) CreateWinRMCredential(ctx context.Context, orgID int, username, password string) (int, error) {
	credentialTypeID, err := c.GetCredentialTypeID(ctx, "Machine")
	if err != nil {
		return 0, fmt.Errorf("failed to get Machine credential type ID: %s", err)
	}

	credentialBody := map[string]interface{}{
		"name":            fmt.Sprintf("packer-winrm-cred-%d", time.Now().Unix()),
		"description":     "WinRM credential for Packer builds",
		"credential_type": credentialTypeID,
		"organization":    orgID,
		"inputs": map[string]interface{}{
			"username": username,
			"password": password,
		},
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(credentialBody).
		Post("/api/controller/v2/credentials/")

	if err != nil {
		return 0, fmt.Errorf("failed to create WinRM credential: %s", err)
	}
	if resp.IsError() {
		return 0, fmt.Errorf("failed to create WinRM credential: %s (status: %d)", resp.String(), resp.StatusCode())
	}

	var result struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return 0, fmt.Errorf("failed to parse WinRM credential response: %s", err)
	}

	if result.ID == 0 {
		return 0, fmt.Errorf("failed to create WinRM credential. Response: %s", resp.String())
	}

	return result.ID, nil
}

func (c *AAPClient) LaunchJob(
	ctx context.Context,
	invID, jobTemplateID, workflowTemplateID, credentialID int,
	extraVars map[string]interface{},
) (int, error) {
	launch := map[string]interface{}{
		"inventory":  invID,
		"extra_vars": extraVars,
	}

	// Add credential if provided
	if credentialID != 0 {
		launch["credentials"] = []int{credentialID}
	}

	// pick the right endpoint
	var endpoint string
	if workflowTemplateID != 0 {
		launch["workflow_template"] = workflowTemplateID
		endpoint = fmt.Sprintf("/api/controller/v2/workflow_job_templates/%d/launch/", workflowTemplateID)
	} else {
		launch["job_template"] = jobTemplateID
		endpoint = fmt.Sprintf("/api/controller/v2/job_templates/%d/launch/", jobTemplateID)
	}

	// send the request
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(launch).
		Post(endpoint)
	if err != nil {
		return 0, fmt.Errorf("failed to launch job: %s", err)
	}
	if resp.IsError() {
		return 0, fmt.Errorf(
			"failed to launch job: %s (status: %d)",
			resp.String(),
			resp.StatusCode(),
		)
	}

	// parse the job ID
	var result struct {
		Job int `json:"job"`
	}
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
			}
			resp, err := c.client.R().SetContext(ctx).SetResult(&result).Get(fmt.Sprintf("/api/controller/v2/jobs/%d/", jobID))
			if err != nil || resp.IsError() {
				return fmt.Errorf("failed to poll job %d: %s", jobID, err)
			}
			if result.Status == "successful" {
				return nil
			}
			if result.Status == "failed" || result.Failed {
				return fmt.Errorf("job %d failed", jobID)
			}
		}
		if time.Since(start) > timeout {
			return errors.New("timeout waiting for job completion")
		}
	}
}

func (c *AAPClient) GetJobStdout(ctx context.Context, jobID int) (string, error) {
	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Accept", "text/plain").
		Get(fmt.Sprintf("/api/controller/v2/jobs/%d/stdout/?format=txt", jobID))

	if err != nil {
		return "", fmt.Errorf("failed to fetch job stdout: %s", err)
	}
	if resp.IsError() {
		return "", fmt.Errorf("failed to fetch job stdout: %s (status: %d)", resp.String(), resp.StatusCode())
	}

	return resp.String(), nil
}
