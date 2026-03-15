package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
	regPath, err := registryDefaultPath()
	if err != nil {
		return nil, err
	}
	entries, _ := registryList(regPath)

	serverMap := make(map[string]registryEntry)
	var alive []registryEntry
	for _, e := range entries {
		if processAlive(e.PID) {
			serverMap[e.Path] = e
			alive = append(alive, e)
		}
	}

	// Clean orphaned entries
	if len(alive) != len(entries) {
		_ = registryWriteAll(regPath, alive)
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
	regPath, _ := registryDefaultPath()

	// If already registered and healthy, just connect the monitor.
	if port := s.healthyPort(regPath, projectPath); port != 0 {
		s.monitor.Connect(projectPath, port)
		return nil
	}

	// Open log file so the daemon writes to .oraculo/server.log.
	logPath := filepath.Join(projectPath, ".oraculo", "server.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer logFile.Close()

	var stderrBuf strings.Builder
	cmd := exec.Command(s.binaryPath, "start", "http")
	cmd.Dir = projectPath
	cmd.Stdout = logFile
	cmd.Stderr = io.MultiWriter(logFile, &stderrBuf)
	detachProcess(cmd)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	// Use a channel to detect fast exits (e.g. port already in use).
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	// Poll until ready (up to 5s) or process exits with error.
	for range 50 {
		time.Sleep(100 * time.Millisecond)

		select {
		case err := <-done:
			msg := strings.TrimSpace(stderrBuf.String())
			if msg != "" {
				return fmt.Errorf("server failed to start: %s", msg)
			}
			if err != nil {
				return fmt.Errorf("server failed to start: %w", err)
			}
			return fmt.Errorf("server exited immediately")
		default:
		}

		if port := s.healthyPort(regPath, projectPath); port != 0 {
			_ = cmd.Process.Release()
			s.monitor.Connect(projectPath, port)
			return nil
		}
	}

	_ = cmd.Process.Release()
	return fmt.Errorf("server did not become ready within 5s")
}

func (s *LauncherService) healthyPort(regPath, projectPath string) int {
	entries, _ := registryList(regPath)
	for _, e := range entries {
		if e.Path == projectPath {
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", e.Port))
			if err == nil {
				resp.Body.Close()
				return e.Port
			}
		}
	}
	return 0
}

func (s *LauncherService) StopServer(ctx context.Context, projectPath string) error {
	s.monitor.Disconnect(projectPath)
	cmd := exec.Command(s.binaryPath, "kill")
	cmd.Dir = projectPath
	return cmd.Run()
}

func findBinary() string {
	// 1. Bundled next to the app executable (Contents/MacOS/oraculo).
	if exe, err := os.Executable(); err == nil {
		bundled := filepath.Join(filepath.Dir(exe), "oraculo")
		if _, err := os.Stat(bundled); err == nil {
			return bundled
		}
	}

	// 2. Common user-local install locations not always in app $PATH.
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, ".local", "bin", "oraculo"),
		"/usr/local/bin/oraculo",
		"/opt/homebrew/bin/oraculo",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// 3. Fall back to PATH lookup.
	return "oraculo"
}
