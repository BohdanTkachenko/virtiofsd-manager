package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	Install    InstallCmd    `command:"install"      description:"Create a Systemd service file and install it."`
	Uninstall  UninstallCmd  `command:"uninstall"    description:"Uninstall Systemd service."`
	Enable     EnableCmd     `command:"enable"       description:"Enables and starts all virtiofsd Systemd services for VM"`
	Disable    DisableCmd    `command:"disable"      description:"Stops and disables all virtiofsd Systemd services for VM"`
	GetVfsArgs GetVfsArgsCmd `command:"get-vfs-args" description:"Generate VFS args string for QEMU."`
}

func main() {
	var cmd Options
	parser := flags.NewParser(&cmd, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			fmt.Println(err)
			os.Exit(0)
		} else {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
		}

		os.Exit(1)
	}
}
