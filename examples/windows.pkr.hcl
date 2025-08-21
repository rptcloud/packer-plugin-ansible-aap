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

source "amazon-ebs" "windows" {
  ami_name      = "packer-windows-demo-${local.timestamp}"
  communicator  = "winrm"
  instance_type = "t3.micro"
  region        = "us-east-1"

  source_ami_filter {
    filters = {
      name                = "Windows_Server-2022-English-Full-Base-*"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    most_recent = true
    owners      = ["amazon"]
  }

  user_data_file = "./examples/bootstrap_win.txt"
  winrm_password = "SuperS3cr3t!!!!"
  winrm_username = "Administrator"
}


build {
  sources = [
    "source.amazon-ebs.windows",
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
