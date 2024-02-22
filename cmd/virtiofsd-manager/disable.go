package main

import (
	"github.com/BohdanTkachenko/virtiofsd-manager/pkg/virtiofsdmanager"
)

type DisableCmd struct {
	verboseMixin

	VmId int `long:"vm_id" short:"i" required:"true" description:"ID of VM to stop and disable services for."`
}

func (cmd *DisableCmd) Execute(args []string) error {
	s, err := virtiofsdmanager.CreateServiceManager(cmd.Verbose)
	if err != nil {
		return err
	}
	if _, err := s.DisableAndStop("*", cmd.VmId); err != nil {
		return err
	}
	return nil
}
