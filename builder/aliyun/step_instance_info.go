package aliyun

import (
	"github.com/mitchellh/multistep"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/mitchellh/packer/packer"
	"fmt"
)

type stepInstanceInfo struct{}

func (s *stepInstanceInfo) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(Config)
	instanceId := state.Get("instance_id").(string)

	// Get instance information
	ui.Say("Getting instance info...")
	instanceIds := fmt.Sprintf("[\"%s\"]", instanceId)
	instances, _, err := client.DescribeInstances(&ecs.DescribeInstancesArgs{
		RegionId: c.RegionId,
		InstanceIds: instanceIds,
	})

	if err != nil {
		err := fmt.Errorf("Error getting instance info: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if instances == nil || len(instances) != 1 {
		err := fmt.Errorf("Fail to get instance info")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	invalid := instances[0].InnerIpAddress.IpAddress == nil ||
	           len(instances[0].InnerIpAddress.IpAddress) == 0 ||
	           instances[0].InnerIpAddress.IpAddress[0] == ""
	if invalid {
		err := fmt.Errorf("inner ip address not found for instance")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	innerIpAddress := instances[0].InnerIpAddress.IpAddress[0]

	ui.Message(fmt.Sprintf("Inner IP : %s", innerIpAddress))
	// for ssh later
	state.Put("inner_ip", innerIpAddress)

	// Get disk information
	ui.Say("Getting disk info...")
	disks, _, err := client.DescribeDisks(&ecs.DescribeDisksArgs{
		RegionId: c.RegionId,
		InstanceId: instanceId,
	})
	if err != nil {
		err := fmt.Errorf("Error getting disk info: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if disks == nil || len(disks) != 1 || disks[0].DiskId == "" {
		err := fmt.Errorf("Fail to get disk info")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("disk_id", disks[0].DiskId)

	return multistep.ActionContinue
}


func (s *stepInstanceInfo) Cleanup(state multistep.StateBag) {
	// no cleanup
}