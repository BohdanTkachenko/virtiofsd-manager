package virtiofsdmanager

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/charmbracelet/log"
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
	logger *log.Logger
	conn   *dbus.Conn
}

func CreateServiceManager(verbose bool) (*ServiceManager, error) {
	conn, err := dbus.NewSystemdConnectionContext(context.TODO())
	if err != nil {
		return nil, err
	}

	logger := log.NewWithOptions(os.Stderr, log.Options{
		Level: log.InfoLevel,
	})
	if verbose {
		logger.SetLevel(log.DebugLevel)
	}

	return &ServiceManager{
		logger: logger,
		conn:   conn,
	}, nil
}

func (s *ServiceManager) ListServices(sharePath string, vmId int) ([]string, error) {
	serviceName := getServiceName(sharePath, vmId)
	s.logger.Info("Listing services",
		"serviceName", serviceName)

	services, err := s.conn.ListUnitFilesByPatternsContext(context.TODO(), []string{}, []string{serviceName})
	if err != nil {
		return nil, err
	}

	servicePaths := []string{}
	for _, service := range services {
		servicePaths = append(servicePaths, service.Path)
	}

	s.logger.Debug("Found matching services",
		"services", services,
		"servicePath", servicePaths)
	return servicePaths, nil
}

func (s *ServiceManager) Install(sharePath string, vmId int, logLevel string, extraArgs string, forceOverwrite bool) (string, error) {
	serviceName := getServiceName(sharePath, vmId)
	serviceFilePath := filepath.Join(SystemDDirectory, serviceName)
	s.logger.Info("Installing unit",
		"sharePath", sharePath,
		"vmId", vmId,
		"logLevel", logLevel,
		"extraArgs", extraArgs,
		"forceOverwrite", forceOverwrite,
		"serviceName", serviceName,
		"serviceFilePath", serviceFilePath)

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

	s.logger.Debug("Created a service file",
		"serviceFilePath", serviceFilePath)

	if err = s.conn.ReloadContext(context.TODO()); err != nil {
		return "", err
	}

	s.logger.Debug("Reloaded systemd context")

	return serviceName, nil
}

func (s *ServiceManager) Uninstall(sharePath string, vmId int) error {
	s.logger.Info("Uninstalling service",
		"sharePath", sharePath,
		"vmId", vmId)

	servicePaths, err := s.DisableAndStop(sharePath, vmId)
	if err != nil {
		return err
	}
	for _, servicePath := range servicePaths {
		if err := os.Remove(servicePath); err != nil {
			return err
		}
	}

	if err = s.conn.ReloadContext(context.TODO()); err != nil {
		return err
	}

	s.logger.Debug("Reloaded systemd context")

	return nil
}

func (s *ServiceManager) EnableAndStart(sharePath string, vmId int) ([]string, error) {
	s.logger.Info("Enable and start service",
		"sharePath", sharePath,
		"vmId", vmId)

	servicePaths, err := s.ListServices(sharePath, vmId)
	if err != nil {
		return nil, err
	}

	if len(servicePaths) == 0 {
		return nil, fmt.Errorf("no services found matching '%s'", getServiceName(sharePath, vmId))
	}

	s.logger.Debug("Enabling services",
		"servicePaths", servicePaths)

	if _, _, err := s.conn.EnableUnitFilesContext(context.TODO(), servicePaths, true, false); err != nil {
		return nil, err
	}

	for _, servicePath := range servicePaths {
		serviceName := filepath.Base(servicePath)

		s.logger.Debug("Starting service",
			"serviceName", serviceName)

		startChan := make(chan string)
		if _, err := s.conn.StartUnitContext(context.TODO(), serviceName, "replace", startChan); err != nil {
			return nil, err
		}
		msg := <-startChan
		if msg != "done" {
			return nil, fmt.Errorf("cannot start '%s' (status: %q)", serviceName, msg)
		}

		s.logger.Debug("Service started",
			"serviceName", serviceName)
	}

	return servicePaths, nil
}

func (s *ServiceManager) DisableAndStop(sharePath string, vmId int) ([]string, error) {
	s.logger.Info("Disable and stop service",
		"sharePath", sharePath,
		"vmId", vmId)

	servicePaths, err := s.ListServices(sharePath, vmId)
	if err != nil {
		return nil, err
	}

	if len(servicePaths) == 0 {
		return nil, fmt.Errorf("no services found matching '%s'", getServiceName(sharePath, vmId))
	}

	services := []string{}
	for _, servicePath := range servicePaths {
		serviceName := filepath.Base(servicePath)
		services = append(services, serviceName)

		s.logger.Debug("Stopping service",
			"serviceName", serviceName)

		stopChan := make(chan string)
		if _, err = s.conn.StopUnitContext(context.TODO(), serviceName, "replace", stopChan); err != nil {
			return nil, err
		}
		msg := <-stopChan
		if msg != "done" {
			return nil, fmt.Errorf("cannot stop '%s' (status: %q)", serviceName, msg)
		}

		s.logger.Debug("Service stopped",
			"serviceName", serviceName)
	}

	s.logger.Debug("Disabling services",
		"services", services)

	if _, err := s.conn.DisableUnitFilesContext(context.TODO(), services, true); err != nil {
		return nil, err
	}

	return servicePaths, nil
}
