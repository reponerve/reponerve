package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/reponerve/reponerve/internal/agent/development"
)

func isDevelopmentTool(name string) bool {
	switch name {
	case "ask", "explain", "explain_file", "explain_function", "explain_struct",
		"explain_interface", "explain_type", "plan", "review", "analyze_topic_impact", "onboard",
		"list_features", "explain_feature", "reuse_check", "ship_check", "pr_context":
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
	outOpts := parseDevelopmentOutputOptions(getArg)

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
		s.sendDevelopmentResult(id, development.FormatAnswer(answer), answer, outOpts)

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
		s.sendDevelopmentResult(id, development.FormatExplanation(out), out, outOpts)

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
		s.sendDevelopmentResult(id, development.FormatExplanation(out), out, outOpts)

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
		s.sendDevelopmentResult(id, development.FormatExplanation(out), out, outOpts)

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
		s.sendDevelopmentResult(id, development.FormatExplanation(out), out, outOpts)

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
		s.sendDevelopmentResult(id, development.FormatExplanation(out), out, outOpts)

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
		s.sendDevelopmentResult(id, development.FormatExplanation(out), out, outOpts)

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
		s.sendDevelopmentResult(id, development.FormatPlan(out), out, outOpts)

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
		s.sendDevelopmentResult(id, development.FormatReviewGuide(out), out, outOpts)

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
		s.sendDevelopmentResult(id, development.FormatImpactReport(out), out, outOpts)

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
		s.sendDevelopmentResult(id, development.FormatOnboarding(out), out, outOpts)

	case "reuse_check":
		intent, err := getArg("intent", false)
		if strings.TrimSpace(intent) == "" {
			intent, err = getArg("topic", true)
		}
		if err != nil {
			s.sendToolError(id, err.Error())
			return true
		}
		out, err := dev.ReuseCheck(ctx, development.DevelopmentRequest{
			RepositoryID: repoID,
			Topic:        intent,
		})
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("reuse_check failed: %v", err))
			return true
		}
		s.sendDevelopmentResult(id, development.FormatReuseCheck(out), out, outOpts)

	case "ship_check":
		topic, err := getArg("topic", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return true
		}
		out, err := dev.ShipCheck(ctx, development.DevelopmentRequest{
			RepositoryID: repoID,
			Topic:        topic,
		})
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("ship_check failed: %v", err))
			return true
		}
		s.sendDevelopmentResult(id, development.FormatShipCheck(out), out, outOpts)

	case "pr_context":
		filesRaw, err := getArg("changed_files", true)
		if err != nil {
			s.sendToolError(id, err.Error())
			return true
		}
		topic, _ := getArg("topic", false)
		out, err := dev.PreparePRContext(ctx, development.PRContextRequest{
			RepositoryID: repoID,
			Topic:        topic,
			ChangedFiles: splitChangedFiles(filesRaw),
		})
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("pr_context failed: %v", err))
			return true
		}
		s.sendDevelopmentResult(id, development.FormatPRContext(out), out, outOpts)

	case "list_features":
		out, err := dev.ListFeatures(ctx, repoID)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("list_features failed: %v", err))
			return true
		}
		s.sendDevelopmentResult(id, development.FormatFeatureList(out), out, outOpts)

	case "explain_feature":
		name, _ := getArg("feature", false)
		if strings.TrimSpace(name) == "" {
			var err error
			name, err = getArg("name", true)
			if err != nil {
				s.sendToolError(id, err.Error())
				return true
			}
		}
		out, err := dev.ExplainFeature(ctx, repoID, name)
		if err != nil {
			s.sendToolError(id, fmt.Sprintf("explain_feature failed: %v", err))
			return true
		}
		s.sendDevelopmentResult(id, development.FormatExplanation(out), out, outOpts)
	}

	return true
}

func splitChangedFiles(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (s *Server) sendDevelopmentResult(id *json.RawMessage, formatted string, structured any, opts development.OutputOptions) {
	s.sendToolSuccess(id, development.NewMCPResultWithFormat(formatted, structured, opts))
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
