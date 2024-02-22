package main

import (
	"github.com/BohdanTkachenko/virtiofsd-manager/pkg/virtiofsdmanager"
)

type EnableCmd struct {
	verboseMixin

	VmId int `long:"vm_id" short:"i" required:"true" description:"ID of VM to enable and start services for."`
}

func (cmd *EnableCmd) Execute(args []string) error {
	s, err := virtiofsdmanager.CreateServiceManager(cmd.Verbose)
	if err != nil {
		return err
	}
	if _, err := s.EnableAndStart("*", cmd.VmId); err != nil {
		return err
	}
	return nil
}
