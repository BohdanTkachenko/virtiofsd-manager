package main

import (
	"github.com/BohdanTkachenko/virtiofsd-manager/pkg/virtiofsdmanager"
)

type EnableCmd struct {
	VmId int `long:"vm_id" short:"i" required:"true" description:"ID of VM to enable and start services for."`
}

func (cmd *EnableCmd) Execute(args []string) error {
	if _, err := virtiofsdmanager.EnableAndStart("*", cmd.VmId); err != nil {
		return err
	}
	return nil
}
