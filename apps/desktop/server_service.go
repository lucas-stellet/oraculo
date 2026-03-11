package main

import (
	"context"
	"fmt"
	"sync"
)

type ServerService struct {
	selectedServer string
	mu             sync.Mutex
}

func NewServerService() *ServerService {
	return &ServerService{}
}

func (s *ServerService) SelectServer(ctx context.Context, port int) string {
	url := fmt.Sprintf("http://localhost:%d", port)
	s.mu.Lock()
	s.selectedServer = url
	s.mu.Unlock()
	return url
}

func (s *ServerService) GetCurrentServer(ctx context.Context) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.selectedServer
}
