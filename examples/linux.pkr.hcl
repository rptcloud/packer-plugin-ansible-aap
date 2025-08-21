packer {
  required_plugins {
    amazon = {
      version = ">= 1.0.0"
      source  = "github.com/hashicorp/amazon"
    }
    ansible-aap = {
      source  = "github.com/rptcloud/ansible-aap"
      version = "1.0.0"
    }
  }
}
variable "job_id" {
  type        = string
  description = "The job ID to use for the Windows build"
}

locals {
  timestamp = formatdate("YYYYMMDDhhmmss", timestamp())
}

source "amazon-ebs" "linux" {
  region = "us-east-1"

  source_ami_filter {
    filters = {
      virtualization-type = "hvm"
      name                = "amzn2-ami-hvm-*-x86_64-gp2"
      root-device-type    = "ebs"
    }
    owners      = ["amazon"]
    most_recent = true
  }

  instance_type        = "t3.micro"
  ssh_username         = "ec2-user"
  ssh_keypair_name     = "packer-aap-key"
  ssh_private_key_file = "packer-aap-key.pem"
  ssh_timeout          = "10m"

  ami_name        = "packer-ansible-demo-${local.timestamp}"
  ami_description = "Packer built AMI with Ansible configuration"

  tags = {
    Name        = "packer-ansible-demo"
    Environment = "demo"
    BuiltBy     = "packer"
  }
}

build {
  sources = [
    "source.amazon-ebs.linux",
  ]

  provisioner "ansible-aap" {
    tower_host           = "https://aap.example.com"
    access_token         = vault("secret/data/aap", "access_token")
    job_template_id      = var.job_id
    organization_id      = 1
    dynamic_inventory    = true
    keep_temp_inventory  = true
    keep_temp_credential = true
    timeout              = "15m"
    poll_interval        = "10s"
  }
}

