package main

import (
	"github.com/BohdanTkachenko/virtiofsd-manager/pkg/virtiofsdmanager"
)

type UninstallCmd struct {
	SharePath string `long:"path"  short:"p" required:"true" description:"Path to a shared directory."`
	VmId      int    `long:"vm_id" short:"i" required:"true" description:"VM ID of the directory with."`
}

func (cmd *UninstallCmd) Execute(args []string) error {
	return virtiofsdmanager.Uninstall(cmd.SharePath, cmd.VmId)
}
