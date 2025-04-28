package main

import (
	"context"

	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/provisioner"
	"github.com/hashicorp/packer-plugin-sdk/provisioner/sdk"
)

type Provisioner struct {
	config Config
	client *AAPClient
	packer.Ui
}

func main() {
	provisioner.Register(&Provisioner{})
}

func (p *Provisioner) Configure(raws ...interface{}) error {
	var cfg Config
	if err := sdk.DecodeConfig(&cfg, raws...); err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	p.config = cfg
	p.client = NewAAPClient(cfg)
	return nil
}

func (p *Provisioner) Provision(ctx context.Context, ui packer.Ui) (provisioner.Result, error) {
	p.Ui = ui
	var invID, hostID int

	if p.config.DynamicInventory {
		hostDetails, err := getHostDetails()
		if err != nil {
			ui.Error(err.Error())
			return provisioner.Result{}, err
		}

		invID, err = p.client.CreateInventory(ctx, p.config.OrganizationID)
		if err != nil {
			ui.Error(err.Error())
			return provisioner.Result{}, err
		}
		ui.Message(fmt.Sprintf("Created temporary inventory %d", invID))

		hostID, err = p.client.CreateHost(ctx, invID, hostDetails)
		if err != nil {
			ui.Error(err.Error())
			return provisioner.Result{}, err
		}
		ui.Message(fmt.Sprintf("Created temporary host %d", hostID))

		defer func() {
			if !p.config.KeepTempInventory {
				if err := p.client.DeleteHost(ctx, hostID); err != nil {
					ui.Error(fmt.Sprintf("Failed to delete host %d: %s", hostID, err))
				}
				if err := p.client.DeleteInventory(ctx, invID); err != nil {
					ui.Error(fmt.Sprintf("Failed to delete inventory %d: %s", invID, err))
				}
				ui.Message(fmt.Sprintf("Cleaned up inventory %d and host %d", invID, hostID))
			}
		}()
	} else {
		invID = p.config.InventoryID
	}

	jobID, err := p.client.LaunchJob(ctx, invID, p.config.JobTemplateID, p.config.WorkflowTemplateID, p.config.ExtraVars)
	if err != nil {
		ui.Error(err.Error())
		return provisioner.Result{}, err
	}
	ui.Message(fmt.Sprintf("Launched AAP job %d", jobID))

	if err := p.client.PollJob(ctx, jobID, p.config.Timeout, p.config.PollInterval); err != nil {
		ui.Error(err.Error())
		return provisioner.Result{}, err
	}
	ui.Message("AAP job completed successfully")

	return provisioner.Result{}, nil
}
