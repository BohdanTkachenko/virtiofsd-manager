package main

import (
	"github.com/BohdanTkachenko/virtiofsd-manager/pkg/virtiofsdmanager"
)

type UninstallCmd struct {
	SharePath string `long:"path"  short:"p"                 description:"If specified, only share with the provided path will be uninstalled."`
	VmId      int    `long:"vm_id" short:"i" required:"true" description:"VM ID of the directory with."`
}

func (cmd *UninstallCmd) Execute(args []string) error {
	s, err := virtiofsdmanager.CreateServiceManager()
	if err != nil {
		return err
	}
	sharePath := "*"
	if cmd.SharePath != "" {
		sharePath = cmd.SharePath
	}
	return s.Uninstall(sharePath, cmd.VmId)
}
