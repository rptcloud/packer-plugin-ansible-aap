# Packer Provisioner Plugin for Ansible Automation Platform (AAP)

This Packer provisioner plugin integrates with Ansible Automation Platform (AAP) to dynamically create inventories, hosts, and credentials, launch job templates, and manage the provisioning process during Packer builds.

## Features
- **Dynamic Resource Management**: Automatically creates temporary inventories, hosts, and SSH credentials
- **Flexible Authentication**: Support for both username/password and access token authentication
- **Automatic Cleanup**: Removes temporary resources in dependency-safe order
- **Job Monitoring**: Polls job status with configurable intervals and timeouts
- **Stdout Retrieval**: Fetches and displays job output for debugging
- **Extra Variables**: Pass custom variables to job templates
- **Resource Retention**: Option to keep temporary inventories for debugging

## Local Development Installation

1. **Build the plugin**:
   ```bash
   packer-sdc mapstructure-to-hcl2 -type Config pkgs/config/config.go
   goreleaser release --clean --skip=validate,publish
   ```

2. **Install the plugin**:
   Install from the dist of the goreleaser for your given arch. For darwin arm64, the command would be:
   ```bash
   packer plugins install --path $PWD/dist/packer-provisioner-ansible-aap_darwin_arm64_v8.0/packer-plugin-ansible-aap github.com/rptcloud/ansible-aap  
   ```

## Usage
See [Examples Directory](./examples/).

## Configuration Options

### Required Configuration
- `tower_host`: AAP API endpoint (e.g., `https://aap.example.com`)
- `organization_id`: ID of the organization for dynamic inventories (required when `dynamic_inventory` is true)

### Job Template Configuration (Choose One)
- `job_template_id`: ID of the job template to run
- `workflow_template_id`: ID of the workflow template to run

### Authentication (Choose One)
- `username` + `password`: Basic authentication
- `access_token`: Bearer token authentication (preferred)

### Inventory Settings
- `inventory_id`: Use an existing inventory instead of creating a new one
- `dynamic_inventory`: Whether to create a temporary inventory (default: true)
- `keep_temp_inventory`: Whether to keep temporary inventories after the build (default: false)

### Credential Management
- `create_credential`: Whether to create temporary credentials (default: true)
- `keep_temp_credential`: Whether to keep temporary credentials after the build (default: false)

### Job Configuration
- `extra_vars`: Map of extra variables to pass to the job template
- `timeout`: Maximum time to wait for job completion (default: "15m")
- `poll_interval`: Interval for polling job status (default: "10s")

### Security Configuration
- `insecure_skip_verify`: Skip SSL certificate verification (default: false)

## Workflow

The provisioner follows this workflow:

1. **Create Inventory**: Creates a temporary inventory in the specified organization
2. **Add Host**: Adds the target host to the inventory with proper Ansible variables
3. **Create Credential**: Creates an SSH credential using the specified private key
4. **Launch Job**: Launches the job template with the inventory and credential
5. **Poll Status**: Monitors job status until completion or failure
6. **Fetch Output**: Retrieves and displays job stdout
7. **Cleanup**: Removes temporary resources (inventory, host, credential)

## Requirements
- Packer with SSH communicator enabled
- AAP with API access and a valid job template
- SSH private key file for target host authentication

## License
MIT

## Kudos
- [@David Joo](https://github.com/glimpsovstar) for the original work.
