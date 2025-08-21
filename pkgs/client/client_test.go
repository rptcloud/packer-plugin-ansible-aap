package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rptcloud/packer-provisioner-ansible-aap/pkgs/client"
	"github.com/rptcloud/packer-provisioner-ansible-aap/pkgs/config"
)

func TestNewAAPClient(t *testing.T) {
	client := client.NewAAPClient(config.Config{
		TowerHost:          "https://aap.example.com",
		Username:           "admin",
		Password:           "secret",
		Timeout:            30 * time.Second,
		InsecureSkipVerify: true,
	})
	if client == nil {
		t.Fatal("Expected client to be created")
	}
}

func TestNewAAPClient_WithToken(t *testing.T) {
	client := client.NewAAPClient(config.Config{
		TowerHost:   "https://aap.example.com",
		AccessToken: "token123",
		Timeout:     30 * time.Second,
	})
	if client == nil {
		t.Fatal("Expected client to be created")
	}
}

func TestAAPClient_CreateInventory(t *testing.T) {
	// Create test server with TLS
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/controller/v2/inventories/" {
			t.Errorf("Expected path /api/controller/v2/inventories/, got %s", r.URL.Path)
		}

		// Check content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Return mock response
		response := map[string]interface{}{"id": 123}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := client.NewAAPClient(config.Config{
		TowerHost:          server.URL,
		Username:           "admin",
		Password:           "secret",
		Timeout:            30 * time.Second,
		InsecureSkipVerify: true,
	})

	inventoryID, err := c.CreateInventory(t.Context(), 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if inventoryID != 123 {
		t.Errorf("Expected inventory ID 123, got %d", inventoryID)
	}
}

func TestAAPClient_CreateInventory_Error(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)

		_, err := w.Write([]byte(`{"error": "Bad request"}`))
		if err != nil {
			t.Fatalf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	c := client.NewAAPClient(config.Config{
		TowerHost:          server.URL,
		Username:           "admin",
		Password:           "secret",
		Timeout:            30 * time.Second,
		InsecureSkipVerify: true,
	})
	_, err := c.CreateInventory(t.Context(), 1)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestAAPClient_CreateHost(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/controller/v2/hosts/" {
			t.Errorf("Expected path /api/controller/v2/hosts/, got %s", r.URL.Path)
		}

		response := map[string]interface{}{"id": 456}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := client.NewAAPClient(config.Config{
		TowerHost:          server.URL,
		Username:           "admin",
		Password:           "secret",
		Timeout:            30 * time.Second,
		InsecureSkipVerify: true,
	})

	details := client.HostDetails{
		Host:     "192.168.1.100",
		Port:     22,
		Username: "ec2-user",
	}

	hostID, err := c.CreateHost(t.Context(), 123, details, "ssh_key")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if hostID != 456 {
		t.Errorf("Expected host ID 456, got %d", hostID)
	}
}

func TestAAPClient_LaunchJob(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/controller/v2/job_templates/42/launch/" {
			t.Errorf("Expected path /api/controller/v2/job_templates/42/launch/, got %s", r.URL.Path)
		}

		response := map[string]interface{}{"job": 999}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := client.NewAAPClient(config.Config{
		TowerHost:          server.URL,
		Username:           "admin",
		Password:           "secret",
		Timeout:            30 * time.Second,
		InsecureSkipVerify: true,
	})

	jobID, err := c.LaunchJob(t.Context(), 123, 42, 0, 0, map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if jobID != 999 {
		t.Errorf("Expected job ID 999, got %d", jobID)
	}
}

func TestAAPClient_LaunchWorkflowJob(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/controller/v2/workflow_job_templates/84/launch/" {
			t.Errorf("Expected path /api/controller/v2/workflow_job_templates/84/launch/, got %s", r.URL.Path)
		}

		response := map[string]interface{}{"job": 888}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := client.NewAAPClient(config.Config{
		TowerHost:          server.URL,
		Username:           "admin",
		Password:           "secret",
		Timeout:            30 * time.Second,
		InsecureSkipVerify: true,
	})

	jobID, err := c.LaunchJob(t.Context(), 123, 0, 84, 0, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if jobID != 888 {
		t.Errorf("Expected job ID 888, got %d", jobID)
	}
}

func TestAAPClient_PollJob_Failed(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status": "failed",
			"failed": true,
			"stdout": "Task failed",
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := client.NewAAPClient(config.Config{
		TowerHost:          server.URL,
		Username:           "admin",
		Password:           "secret",
		Timeout:            30 * time.Second,
		InsecureSkipVerify: true,
	})
	ctx := t.Context()

	err := c.PollJob(ctx, 999, 30*time.Second, 100*time.Millisecond)
	if err == nil {
		t.Fatal("Expected error for failed job, got nil")
	}
}

func TestAAPClient_PollJob_Timeout(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status": "running",
			"failed": false,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := client.NewAAPClient(config.Config{
		TowerHost:          server.URL,
		Username:           "admin",
		Password:           "secret",
		Timeout:            30 * time.Second,
		InsecureSkipVerify: true,
	})
	ctx := t.Context()

	err := c.PollJob(ctx, 999, 200*time.Millisecond, 100*time.Millisecond)
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
}

func TestAAPClient_DeleteOperations(t *testing.T) {
	deleteCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		deleteCount++
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := client.NewAAPClient(config.Config{
		TowerHost:          server.URL,
		Username:           "admin",
		Password:           "secret",
		Timeout:            30 * time.Second,
		InsecureSkipVerify: true,
	})

	// Test delete host
	err := c.DeleteHost(t.Context(), 456)
	if err != nil {
		t.Fatalf("Expected no error deleting host, got %v", err)
	}

	// Test delete inventory
	err = c.DeleteInventory(t.Context(), 123)
	if err != nil {
		t.Fatalf("Expected no error deleting inventory, got %v", err)
	}

	if deleteCount != 2 {
		t.Errorf("Expected 2 delete operations, got %d", deleteCount)
	}
}

func TestAAPClient_CreateCredential(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/api/controller/v2/credential_types/" {
			response := map[string]interface{}{
				"results": []map[string]interface{}{
					{"id": 4, "name": "Machine"},
				},
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Errorf("Failed to encode credential types response: %v", err)
			}
			return
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/controller/v2/credentials/" {
			t.Errorf("Expected path /api/controller/v2/credentials/, got %s", r.URL.Path)
		}

		response := map[string]interface{}{"id": 789}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := client.NewAAPClient(config.Config{
		TowerHost:          server.URL,
		Username:           "admin",
		Password:           "secret",
		Timeout:            30 * time.Second,
		InsecureSkipVerify: true,
	})

	credentialID, err := c.CreateCredential(t.Context(), 1, "testuser", "test-private-key")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if credentialID != 789 {
		t.Errorf("Expected credential ID 789, got %d", credentialID)
	}
}

func TestAAPClient_CreatePasswordCredential(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/api/controller/v2/credential_types/" {
			response := map[string]interface{}{
				"results": []map[string]interface{}{
					{"id": 4, "name": "Machine"},
				},
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Errorf("Failed to encode credential types response: %v", err)
			}
			return
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/controller/v2/credentials/" {
			t.Errorf("Expected path /api/controller/v2/credentials/, got %s", r.URL.Path)
		}

		response := map[string]interface{}{"id": 790}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := client.NewAAPClient(config.Config{
		TowerHost:          server.URL,
		Username:           "admin",
		Password:           "secret",
		Timeout:            30 * time.Second,
		InsecureSkipVerify: true,
	})

	credentialID, err := c.CreatePasswordCredential(t.Context(), 1, "testuser", "testpass")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if credentialID != 790 {
		t.Errorf("Expected credential ID 790, got %d", credentialID)
	}
}

func TestAAPClient_CreateWinRMCredential(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/api/controller/v2/credential_types/" {
			response := map[string]interface{}{
				"results": []map[string]interface{}{
					{"id": 4, "name": "Machine"},
				},
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Errorf("Failed to encode credential types response: %v", err)
			}
			return
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/controller/v2/credentials/" {
			t.Errorf("Expected path /api/controller/v2/credentials/, got %s", r.URL.Path)
		}

		response := map[string]interface{}{"id": 791}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := client.NewAAPClient(config.Config{
		TowerHost:          server.URL,
		Username:           "admin",
		Password:           "secret",
		Timeout:            30 * time.Second,
		InsecureSkipVerify: true,
	})

	credentialID, err := c.CreateWinRMCredential(t.Context(), 1, "testuser", "testpass")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if credentialID != 791 {
		t.Errorf("Expected credential ID 791, got %d", credentialID)
	}
}

func TestAAPClient_DeleteCredential(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if r.URL.Path != "/api/controller/v2/credentials/789/" {
			t.Errorf("Expected path /api/controller/v2/credentials/789/, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := client.NewAAPClient(config.Config{
		TowerHost:          server.URL,
		Username:           "admin",
		Password:           "secret",
		Timeout:            30 * time.Second,
		InsecureSkipVerify: true,
	})
	ctx := t.Context()

	err := c.DeleteCredential(ctx, 789)
	if err != nil {
		t.Fatalf("Expected no error deleting credential, got %v", err)
	}
}

func TestAAPClient_GetCredentialTypeID(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/controller/v2/credential_types/" {
			t.Errorf("Expected path /api/controller/v2/credential_types/, got %s", r.URL.Path)
		}

		response := map[string]interface{}{
			"results": []map[string]interface{}{
				{"id": 4, "name": "Machine"},
				{"id": 2, "name": "Source Control"},
				{"id": 3, "name": "Vault"},
			},
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	c := client.NewAAPClient(config.Config{
		TowerHost:          server.URL,
		Username:           "admin",
		Password:           "secret",
		Timeout:            30 * time.Second,
		InsecureSkipVerify: true,
	})

	// Test finding existing credential type
	credentialTypeID, err := c.GetCredentialTypeID(t.Context(), "Machine")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if credentialTypeID != 4 {
		t.Errorf("Expected credential type ID 4, got %d", credentialTypeID)
	}

	// Test finding non-existing credential type
	_, err = c.GetCredentialTypeID(t.Context(), "NonExistent")
	if err == nil {
		t.Fatal("Expected error for non-existent credential type, got nil")
	}
}
