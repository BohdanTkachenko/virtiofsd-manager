package virtiofsdmanager

import (
	"fmt"
	"path/filepath"
	"strings"
)

func getShareNameFromUnitName(unitName string, vmId int) string {
	return strings.TrimPrefix(unitName, fmt.Sprintf("virtiofsd-%d-", vmId))
}

func getVmShares(vmId int) ([]string, error) {
	services, err := ListServices("*", vmId)
	if err != nil {
		return nil, err
	}
	shares := []string{}
	for _, servicePath := range services {
		unitName := strings.TrimSuffix(filepath.Base(servicePath), ".service")
		shares = append(shares, getShareNameFromUnitName(unitName, vmId))
	}
	return shares, nil
}

func GetVfsArgs(vmId int) (string, error) {
	shares, err := getVmShares(vmId)
	if err != nil {
		return "", err
	}
	vfsArgs := []string{}
	for _, shareName := range shares {
		vfsArgs = append(vfsArgs, fmt.Sprintf("-chardev socket,id=%s,path=/run/virtiofsd/%d-%s.sock", shareName, vmId, shareName))
		vfsArgs = append(vfsArgs, fmt.Sprintf("-device vhost-user-fs-pci,chardev=%s,tag=%s", shareName, shareName))
	}
	return strings.Join(vfsArgs, " "), nil
}
