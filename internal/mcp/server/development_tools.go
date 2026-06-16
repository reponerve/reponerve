package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/reponerve/reponerve/internal/agent/development"
)

func isDevelopmentTool(name string) bool {
	switch name {
	case "ask", "explain", "explain_file", "explain_function", "explain_struct",
		"explain_interface", "explain_type", "plan", "review", "analyze_topic_impact", "onboard":
		return true
	default:
		return false
	}
}

func (s *Server) handleDevelopmentTool(
	ctx context.Context,
	id *json.RawMessage,
	toolName string,
	getArg func(string, bool) (string, error),
	resolveRepoID func() (string, error),
) bool {
	if !isDevelopmentTool(toolName) {
		return false
	}

	if s.service.DevelopmentService == nil {
		s.sendToolError(id, "development experience service is not configured")
		return true
	}

	repoID, err := resolveRepoID()
	if err != nil || repoID == "" {
		s.sendToolError(id, "failed to resolve repository ID")
		return true
	}

	dev := s.service.DevelopmentService

	switch toolName {
	case "ask":
		question, err := getArg("question", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return true
		}
		answer, err := dev.Ask(ctx, development.DevelopmentRequest{
			RepositoryID: repoID,
			Topic:        question,
		})
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("ask failed: %v", err))
			return true
		}
		s.sendDevelopmentResult(id, development.FormatAnswer(answer), answer)

	case "explain":
		topic, err := getArg("topic", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return true
		}
		out, err := dev.Explain(ctx, development.DevelopmentRequest{
			RepositoryID: repoID,
			Topic:        topic,
		})
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("explain failed: %v", err))
			return true
		}
		s.sendDevelopmentResult(id, development.FormatExplanation(out), out)

	case "explain_file":
		filePath, err := getArg("file_path", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return true
		}
		out, err := dev.ExplainFile(ctx, repoID, filePath)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("explain_file failed: %v", err))
			return true
		}
		s.sendDevelopmentResult(id, development.FormatExplanation(out), out)

	case "explain_function":
		symbol, err := getArg("symbol", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return true
		}
		packagePath, _ := getArg("package_path", false)
		out, err := dev.ExplainFunction(ctx, repoID, symbol, packagePath)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("explain_function failed: %v", err))
			return true
		}
		s.sendDevelopmentResult(id, development.FormatExplanation(out), out)

	case "explain_struct":
		symbol, err := getArg("symbol", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return true
		}
		packagePath, _ := getArg("package_path", false)
		out, err := dev.ExplainStruct(ctx, repoID, symbol, packagePath)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("explain_struct failed: %v", err))
			return true
		}
		s.sendDevelopmentResult(id, development.FormatExplanation(out), out)

	case "explain_interface":
		symbol, err := getArg("symbol", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return true
		}
		packagePath, _ := getArg("package_path", false)
		out, err := dev.ExplainInterface(ctx, repoID, symbol, packagePath)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("explain_interface failed: %v", err))
			return true
		}
		s.sendDevelopmentResult(id, development.FormatExplanation(out), out)

	case "explain_type":
		symbol, err := getArg("symbol", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return true
		}
		packagePath, _ := getArg("package_path", false)
		out, err := dev.ExplainType(ctx, repoID, symbol, packagePath)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("explain_type failed: %v", err))
			return true
		}
		s.sendDevelopmentResult(id, development.FormatExplanation(out), out)

	case "plan":
		task, err := getArg("task", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return true
		}
		out, err := dev.Plan(ctx, development.DevelopmentRequest{
			RepositoryID: repoID,
			Topic:        task,
		})
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("plan failed: %v", err))
			return true
		}
		s.sendDevelopmentResult(id, development.FormatPlan(out), out)

	case "review":
		topic, err := getArg("topic", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return true
		}
		out, err := dev.PrepareReview(ctx, development.DevelopmentRequest{
			RepositoryID: repoID,
			Topic:        topic,
		})
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("review failed: %v", err))
			return true
		}
		s.sendDevelopmentResult(id, development.FormatReviewGuide(out), out)

	case "analyze_topic_impact":
		subject, err := getArg("subject", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return true
		}
		out, err := dev.AnalyzeImpact(ctx, development.DevelopmentRequest{
			RepositoryID: repoID,
			Topic:        subject,
		})
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("analyze_topic_impact failed: %v", err))
			return true
		}
		s.sendDevelopmentResult(id, development.FormatImpactReport(out), out)

	case "onboard":
		assignment, _ := getArg("assignment", false)
		out, err := dev.Onboard(ctx, development.DevelopmentRequest{
			RepositoryID: repoID,
			Topic:        assignment,
		})
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("onboard failed: %v", err))
			return true
		}
		s.sendDevelopmentResult(id, development.FormatOnboarding(out), out)
	}

	return true
}

func (s *Server) sendDevelopmentResult(id *json.RawMessage, formatted string, structured any) {
	s.sendToolSuccess(id, development.NewMCPResult(formatted, structured))
}

func (s *Server) sendToolText(id *json.RawMessage, text string) {
	result := CallToolResult{
		Content: []ContentBlock{
			{
				Type: "text",
				Text: text,
			},
		},
		IsError: false,
	}
	s.sendResult(id, result)
}
