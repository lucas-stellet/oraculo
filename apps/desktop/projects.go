package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type KnownProject struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

func projectsFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".oraculo", "projects.json"), nil
}

func readProjects() ([]KnownProject, error) {
	path, err := projectsFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var projects []KnownProject
	return projects, json.Unmarshal(data, &projects)
}

func writeProjects(projects []KnownProject) error {
	path, err := projectsFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
