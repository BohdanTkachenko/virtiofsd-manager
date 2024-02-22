package main

import (
	"fmt"

	"github.com/BohdanTkachenko/virtiofsd-manager/pkg/virtiofsdmanager"
)

type InstallCmd struct {
	verboseMixin

	SharePath      string `long:"path"       short:"p" required:"true" description:"Path to a directory that needs to be shared."`
	VmId           int    `long:"vm_id"      short:"i" required:"true" description:"VM ID to share the directory with."`
	ForceOverwrite bool   `long:"force"      short:"f"                 description:"Force overwrite existing service unit file."`
	LogLevel       string `long:"log_level"  short:"l" default:"debug" description:"Log level to start virtiofsd with."`
	ExtraArgs      string `long:"extra_args" short:"a"                 description:"Additoinal args to pass to virtiofsd."`
}

func (cmd *InstallCmd) Execute(args []string) error {
	s, err := virtiofsdmanager.CreateServiceManager(cmd.Verbose)
	if err != nil {
		return err
	}
	serviceName, err := s.Install(cmd.SharePath, cmd.VmId, cmd.LogLevel, cmd.ExtraArgs, cmd.ForceOverwrite)
	if err != nil {
		return err
	}
	fmt.Println(serviceName)
	return nil
}
