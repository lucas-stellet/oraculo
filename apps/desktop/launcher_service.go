package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/lucas/oraculo/apps/backend/src/registry"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type ProjectWithStatus struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Online bool   `json:"online"`
	Port   int    `json:"port,omitempty"`
}

type LauncherService struct {
	binaryPath string
	monitor    *wsMonitor
}

func NewLauncherService() *LauncherService {
	return &LauncherService{
		monitor: newWSMonitor(),
	}
}

// ServiceStartup is called by Wails v3 during app initialization.
func (s *LauncherService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	s.binaryPath = findBinary()
	return nil
}

// ServiceShutdown is called by Wails v3 during app teardown.
func (s *LauncherService) ServiceShutdown() error {
	s.monitor.DisconnectAll()
	return nil
}

func (s *LauncherService) ListProjects(ctx context.Context) ([]ProjectWithStatus, error) {
	projects, _ := readProjects()
	regPath, err := registry.DefaultPath()
	if err != nil {
		return nil, err
	}
	entries, _ := registry.List(regPath)

	serverMap := make(map[string]registry.Entry)
	var alive []registry.Entry
	for _, e := range entries {
		if processAlive(e.PID) {
			serverMap[e.Path] = e
			alive = append(alive, e)
		}
	}

	// Clean orphaned entries
	if len(alive) != len(entries) {
		_ = registry.WriteAll(regPath, alive)
	}

	var result []ProjectWithStatus
	for _, p := range projects {
		ps := ProjectWithStatus{Name: p.Name, Path: p.Path}
		if e, ok := serverMap[p.Path]; ok {
			ps.Online = true
			ps.Port = e.Port
		}
		result = append(result, ps)
	}
	return result, nil
}

func (s *LauncherService) AddProject(ctx context.Context) (*KnownProject, error) {
	dir, err := application.Get().Dialog.OpenFile().
		SetTitle("Select Oraculo Project").
		CanChooseDirectories(true).
		CanChooseFiles(false).
		PromptForSingleSelection()
	if err != nil || dir == "" {
		return nil, err
	}

	if _, err := os.Stat(filepath.Join(dir, ".oraculo")); err != nil {
		return nil, fmt.Errorf("selected directory is not an Oraculo project (missing .oraculo/)")
	}

	name := filepath.Base(dir)
	project := KnownProject{Path: dir, Name: name}

	projects, _ := readProjects()
	for _, p := range projects {
		if p.Path == dir {
			return &project, nil
		}
	}
	projects = append(projects, project)
	if err := writeProjects(projects); err != nil {
		return nil, err
	}
	return &project, nil
}

func (s *LauncherService) RemoveProject(ctx context.Context, projectPath string) error {
	projects, err := readProjects()
	if err != nil {
		return err
	}
	for i, p := range projects {
		if p.Path == projectPath {
			projects = append(projects[:i], projects[i+1:]...)
			return writeProjects(projects)
		}
	}
	return nil
}

func (s *LauncherService) StartServer(ctx context.Context, projectPath string) error {
	cmd := exec.Command(s.binaryPath, "start", "http")
	cmd.Dir = projectPath
	detachProcess(cmd)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	_ = cmd.Process.Release()

	// Poll health endpoint until ready (up to 5s).
	regPath, _ := registry.DefaultPath()
	for range 50 {
		time.Sleep(100 * time.Millisecond)
		entries, _ := registry.List(regPath)
		for _, e := range entries {
			if e.Path == projectPath {
				resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", e.Port))
				if err == nil {
					resp.Body.Close()
					// Connect WS monitor for event notifications.
					s.monitor.Connect(projectPath, e.Port)
					return nil
				}
			}
		}
	}
	return fmt.Errorf("server did not become ready within 5s")
}

func (s *LauncherService) StopServer(ctx context.Context, projectPath string) error {
	s.monitor.Disconnect(projectPath)
	cmd := exec.Command(s.binaryPath, "kill")
	cmd.Dir = projectPath
	return cmd.Run()
}

func findBinary() string {
	exe, err := os.Executable()
	if err != nil {
		return "oraculo"
	}
	bundled := filepath.Join(filepath.Dir(exe), "oraculo")
	if _, err := os.Stat(bundled); err == nil {
		return bundled
	}
	return "oraculo"
}
