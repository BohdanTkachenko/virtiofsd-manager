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

type ServiceManager struct {
	conn *dbus.Conn
}

func CreateServiceManager() (*ServiceManager, error) {
	conn, err := dbus.NewSystemdConnectionContext(context.TODO())
	if err != nil {
		return nil, err
	}

	return &ServiceManager{
		conn: conn,
	}, nil
}

func (s *ServiceManager) ListServices(sharePath string, vmId int) ([]string, error) {
	units, err := s.conn.ListUnitFilesByPatternsContext(context.TODO(), []string{}, []string{
		getServiceName(sharePath, vmId),
	})
	if err != nil {
		return nil, err
	}
	unitPaths := []string{}
	for _, unit := range units {
		unitPaths = append(unitPaths, unit.Path)
	}
	return unitPaths, nil
}

func (s *ServiceManager) Install(sharePath string, vmId int, logLevel string, extraArgs string, forceOverwrite bool) (string, error) {
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

	if err = s.conn.ReloadContext(context.TODO()); err != nil {
		return "", err
	}

	return serviceName, nil
}

func (s *ServiceManager) Uninstall(sharePath string, vmId int) error {
	unitPaths, err := s.DisableAndStop(sharePath, vmId)
	if err != nil {
		return err
	}
	for _, unitPath := range unitPaths {
		if err := os.Remove(unitPath); err != nil {
			return err
		}
	}
	if err = s.conn.ReloadContext(context.TODO()); err != nil {
		return err
	}
	return nil
}

func (s *ServiceManager) EnableAndStart(sharePath string, vmId int) ([]string, error) {
	unitPaths, err := s.ListServices(sharePath, vmId)
	if err != nil {
		return nil, err
	}

	if len(unitPaths) == 0 {
		return nil, fmt.Errorf("no services found matching '%s'", getServiceName(sharePath, vmId))
	}

	if _, _, err := s.conn.EnableUnitFilesContext(context.TODO(), unitPaths, true, false); err != nil {
		return nil, err
	}

	return unitPaths, nil
}

func (s *ServiceManager) DisableAndStop(sharePath string, vmId int) ([]string, error) {
	unitPaths, err := s.ListServices(sharePath, vmId)
	if err != nil {
		return nil, err
	}

	if len(unitPaths) == 0 {
		return nil, fmt.Errorf("no services found matching '%s'", getServiceName(sharePath, vmId))
	}

	units := []string{}
	for _, unitPath := range unitPaths {
		unitName := filepath.Base(unitPath)
		if _, err = s.conn.StopUnitContext(context.TODO(), unitName, "replace", nil); err != nil {
			return nil, err
		}
		units = append(units, unitName)
	}
	if _, err := s.conn.DisableUnitFilesContext(context.TODO(), units, true); err != nil {
		return nil, err
	}
	return unitPaths, nil
}
