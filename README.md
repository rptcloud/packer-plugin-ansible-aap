# Packer Provisioner Plugin for Ansible Automation Platform (AAP)

This Packer provisioner plugin integrates with Ansible Automation Platform (AAP) to dynamically create inventories and hosts, launch job templates, and manage the provisioning process during Packer builds.

## Features
- Dynamically creates AAP inventories and hosts based on Packer's SSH communicator.
- Launches AAP job templates with custom extra variables.
- Polls job status until completion.
- Optional cleanup of temporary inventories and hosts.
- Configurable timeout and polling interval.

## Installation

1. **Build the plugin**:
   ```bash
   go build -o packer-provisioner-ansible-aap
   ```

2. **Install the plugin**:
   Move the binary to Packer's plugin directory:
   ```bash
   mkdir -p ~/.packer.d/plugins
   mv packer-provisioner-ansible-aap ~/.packer.d/plugins/
   ```

## Usage

Add the provisioner to your Packer template:

```hcl
source "amazon-ebs" "example" {
  ami_name      = "ansible-aap-test-{{timestamp}}"
  instance_type = "t2.micro"
  region        = "us-east-1"
  ssh_username  = "ubuntu"
  source_ami    = "ami-12345678"
}

build {
  sources = ["source.amazon-ebs.example"]

  provisioner "ansible-aap" {
    tower_host          = "https://aap.example.com"
    username            = "admin"
    password            = "secret"
    job_template_id     = 42
    organization_id     = 1
    dynamic_inventory   = true
    keep_temp_inventory = false
    extra_vars = {
      key1 = "value1"
      key2 = "value2"
    }
    timeout       = "15m"
    poll_interval = "10s"
  }
}
```

## Configuration Options
- `tower_host`: AAP API endpoint (e.g., `https://aap.example.com`).
- `username`: AAP API username.
- `password`: AAP API password.
- `job_template_id`: ID of the job template to run.
- `inventory_id`: ID of an existing inventory (if `dynamic_inventory` is false).
- `organization_id`: ID of the organization for dynamic inventories.
- `dynamic_inventory`: Whether to create a temporary inventory (default: false).
- `keep_temp_inventory`: Whether to keep temporary inventories after the build (default: false).
- `extra_vars`: Extra variables to pass to the job template.
- `timeout`: Maximum time to wait for job completion (default: 30m).
- `poll_interval`: Interval for polling job status (default: 5s).

## Requirements
- Packer with SSH communicator enabled.
- AAP with API access and a valid job template.
- Environment variables set by Packer: `PACKER_SSH_HOST`, `PACKER_SSH_PORT`, `PACKER_SSH_USERNAME`, and either `PACKER_SSH_KEY_FILE` or `PACKER_SSH_PASSWORD`.

## Development
To contribute or modify the plugin:
1. Clone the repository.
2. Make changes to `provisioner.go`, `config.go`, or `client.go`.
3. Build and test:
   ```bash
   go test ./...
   go build -o packer-provisioner-ansible-aap
   ```

## License
MIT