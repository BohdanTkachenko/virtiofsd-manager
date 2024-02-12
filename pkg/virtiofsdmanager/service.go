package virtiofsdmanager

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/coreos/go-systemd/v22/dbus"
)

//go:embed templates/virtiofsd.service
var templateFS embed.FS

type TemplateData struct {
	SharePath string
	ShareName string
	VmId      int
	LogLevel  string
	ExtraArgs string
}

func getShareNameFromPath(path string) string {
	return strings.ReplaceAll(strings.Trim(path, "/"), "/", "__")
}

func getServiceName(sharePath string, vmId int) string {
	return fmt.Sprintf("virtiofsd-%d-%s.service", vmId, getShareNameFromPath(sharePath))
}

func ListServices(shareName string, vmId int) ([]string, error) {
	systemd, err := dbus.NewSystemdConnectionContext(context.TODO())
	if err != nil {
		return nil, err
	}
	pattern := fmt.Sprintf("virtiofsd-%d-%s.service", vmId, shareName)
	units, err := systemd.ListUnitFilesByPatternsContext(context.TODO(), []string{}, []string{pattern})
	if err != nil {
		return nil, err
	}
	unitPaths := []string{}
	for _, unit := range units {
		unitPaths = append(unitPaths, unit.Path)
	}
	return unitPaths, nil
}

func Install(sharePath string, vmId int, logLevel string, extraArgs string, forceOverwrite bool) (string, error) {
	if _, err := os.Stat(sharePath); err != nil {
		return "", err
	}

	tmpl, err := template.ParseFS(templateFS, "templates/virtiofsd.service")
	if err != nil {
		return "", err
	}

	ShareName := getShareNameFromPath(sharePath)
	data := TemplateData{
		SharePath: sharePath,
		ShareName: ShareName,
		VmId:      vmId,
		LogLevel:  logLevel,
		ExtraArgs: extraArgs,
	}

	serviceName := getServiceName(sharePath, vmId)
	serviceFilePath := filepath.Join(SystemDDirectory, serviceName)
	if !forceOverwrite {
		if _, err := os.Stat(serviceFilePath); err == nil {
			return "", fmt.Errorf("service unit file '%s' alread exists", serviceFilePath)
		}
	}

	file, err := os.Create(serviceFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err = tmpl.Execute(file, data); err != nil {
		return "", err
	}

	systemd, err := dbus.NewSystemdConnectionContext(context.TODO())
	if err != nil {
		return "", err
	}

	if err = systemd.ReloadContext(context.TODO()); err != nil {
		return "", err
	}

	return serviceName, nil
}

func Uninstall(sharePath string, vmId int) error {
	systemd, err := dbus.NewSystemdConnectionContext(context.TODO())
	if err != nil {
		return err
	}
	unitPaths, err := DisableAndStop(sharePath, vmId)
	if err != nil {
		return err
	}
	for _, unitPath := range unitPaths {
		if err := os.Remove(unitPath); err != nil {
			return err
		}
	}
	if err = systemd.ReloadContext(context.TODO()); err != nil {
		return err
	}
	return nil
}

func EnableAndStart(sharePath string, vmId int) ([]string, error) {
	systemd, err := dbus.NewSystemdConnectionContext(context.TODO())
	if err != nil {
		return nil, err
	}
	unitPaths, err := ListServices(sharePath, vmId)
	if err != nil {
		return nil, err
	}
	if _, _, err := systemd.EnableUnitFilesContext(context.TODO(), unitPaths, true, false); err != nil {
		return nil, err
	}
	return unitPaths, nil
}

func DisableAndStop(sharePath string, vmId int) ([]string, error) {
	systemd, err := dbus.NewSystemdConnectionContext(context.TODO())
	if err != nil {
		return nil, err
	}
	unitPaths, err := ListServices(sharePath, vmId)
	if err != nil {
		return nil, err
	}
	units := []string{}
	for _, unitPath := range unitPaths {
		unitName := filepath.Base(unitPath)
		if _, err = systemd.StopUnitContext(context.TODO(), unitName, "replace", nil); err != nil {
			return nil, err
		}
		units = append(units, unitName)
	}
	if _, err := systemd.DisableUnitFilesContext(context.TODO(), units, true); err != nil {
		return nil, err
	}
	return unitPaths, nil
}
