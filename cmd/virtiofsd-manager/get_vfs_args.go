package main

import (
	"fmt"

	"github.com/BohdanTkachenko/virtiofsd-manager/pkg/virtiofsdmanager"
)

type GetVfsArgsCmd struct {
	verboseMixin

	VmId int `long:"vm_id" short:"i" required:"true" description:"VM ID to return args for."`
}

func (cmd *GetVfsArgsCmd) Execute(args []string) error {
	vfsArgs, err := virtiofsdmanager.GetVfsArgs(cmd.VmId, cmd.Verbose)
	if err != nil {
		return err
	}
	fmt.Println(vfsArgs)
	return nil
}
