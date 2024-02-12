package main

import (
	"github.com/BohdanTkachenko/virtiofsd-manager/pkg/virtiofsdmanager"
)

type DisableCmd struct {
	VmId int `long:"vm_id" short:"i" required:"true" description:"ID of VM to stop and disable services for."`
}

func (cmd *DisableCmd) Execute(args []string) error {
	if _, err := virtiofsdmanager.DisableAndStop("*", cmd.VmId); err != nil {
		return err
	}
	return nil
}
