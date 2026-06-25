package server

import (
	"context"
	"encoding/json"

	"github.com/reponerve/reponerve/internal/agent/development"
	"github.com/reponerve/reponerve/internal/agent/health"
	"github.com/reponerve/reponerve/internal/config"
)

func isHealthTool(name string) bool {
	return name == "doctor"
}

func (s *Server) handleHealthTool(
	ctx context.Context,
	id *json.RawMessage,
	toolName string,
) bool {
	if !isHealthTool(toolName) {
		return false
	}
	if s.service.HealthChecker == nil {
		s.sendToolError(id, "health checker is not configured")
		return true
	}

	result, err := s.service.HealthChecker.Check(ctx, health.CheckInput{
		WorkspaceDir: config.GetWorkspaceDir(),
	})
	if err != nil {
		s.sendToolError(id, err.Error())
		return true
	}

	formatted := health.FormatDoctor(result)
	payload := development.NewMCPResult(formatted, result)
	s.sendToolSuccess(id, payload)
	return true
}
