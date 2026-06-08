package workflow

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const (
	ArtifactTypeSession         = "session"
	ArtifactTypeContextPackage  = "context_package"
	ArtifactTypeSearchResult    = "search_result"
	ArtifactTypeDiscoveryReport = "discovery_report"
	ArtifactTypeLearningPath    = "learning_path"
	ArtifactTypeReviewerReport  = "reviewer_report"
	ArtifactTypeChangePlan      = "change_plan"

	WorkflowTypeOnboarding           = "onboarding"
	WorkflowTypeReviewPreparation    = "review_preparation"
	WorkflowTypeChangePreparation    = "change_preparation"
	WorkflowTypeKnowledgeExploration = "knowledge_exploration"

	SourceSession    = "session"
	SourceContext    = "context"
	SourceSearch     = "search"
	SourceDiscovery  = "discovery"
	SourceLearning   = "learning"
	SourceReviewers  = "reviewers"
	SourceChangePlan = "changeplan"

	VersionV1 = "v1"
)

const defaultWorkflowIdentifier = "default"

var validArtifactTypes = map[string]bool{
	ArtifactTypeSession:         true,
	ArtifactTypeContextPackage:  true,
	ArtifactTypeSearchResult:    true,
	ArtifactTypeDiscoveryReport: true,
	ArtifactTypeLearningPath:    true,
	ArtifactTypeReviewerReport:  true,
	ArtifactTypeChangePlan:      true,
}

var validWorkflowTypes = map[string]bool{
	WorkflowTypeOnboarding:           true,
	WorkflowTypeReviewPreparation:    true,
	WorkflowTypeChangePreparation:    true,
	WorkflowTypeKnowledgeExploration: true,
}

var validSources = map[string]bool{
	SourceSession:    true,
	SourceContext:    true,
	SourceSearch:     true,
	SourceDiscovery:  true,
	SourceLearning:   true,
	SourceReviewers:  true,
	SourceChangePlan: true,
}

type WorkflowArtifact struct {
	ArtifactType string          `json:"artifact_type"`
	Source       string          `json:"source"`
	Data         json.RawMessage `json:"data"`
}

type WorkflowPackage struct {
	WorkflowID   string              `json:"workflow_id"`
	WorkflowType string              `json:"workflow_type"`
	RepositoryID string              `json:"repository_id"`
	Version      string              `json:"version"`
	Artifacts    []*WorkflowArtifact `json:"artifacts"`
}

func ValidateWorkflow(workflow *WorkflowPackage) error {
	if workflow == nil {
		return fmt.Errorf("workflow is nil")
	}
	if workflow.WorkflowID == "" {
		return fmt.Errorf("missing workflow ID")
	}
	if workflow.WorkflowType == "" {
		return fmt.Errorf("missing workflow type")
	}
	if !validWorkflowTypes[workflow.WorkflowType] {
		return fmt.Errorf("unsupported workflow type %q (must be one of: onboarding, review_preparation, change_preparation, knowledge_exploration)", workflow.WorkflowType)
	}
	if workflow.RepositoryID == "" {
		return fmt.Errorf("missing repository ID")
	}
	if workflow.Version == "" {
		return fmt.Errorf("missing version")
	}
	if len(workflow.Artifacts) == 0 {
		return fmt.Errorf("workflow has no artifacts")
	}

	for i, artifact := range workflow.Artifacts {
		if artifact == nil {
			return fmt.Errorf("artifact %d is nil", i)
		}
		if artifact.ArtifactType == "" {
			return fmt.Errorf("artifact %d: missing artifact type", i)
		}
		if !validArtifactTypes[artifact.ArtifactType] {
			return fmt.Errorf("artifact %d: unsupported artifact type %q (must be one of: session, context_package, search_result, discovery_report, learning_path, reviewer_report, change_plan)", i, artifact.ArtifactType)
		}
		if artifact.Source == "" {
			return fmt.Errorf("artifact %d: missing source", i)
		}
		if !validSources[artifact.Source] {
			return fmt.Errorf("artifact %d: unsupported source %q (must be one of: session, context, search, discovery, learning, reviewers, changeplan)", i, artifact.Source)
		}
		if len(artifact.Data) == 0 {
			return fmt.Errorf("artifact %d: missing data", i)
		}
		if !json.Valid(artifact.Data) {
			return fmt.Errorf("artifact %d: data must be valid JSON", i)
		}
	}

	return nil
}

func buildWorkflowID(repositoryID string, workflowType string, identifier string) string {
	sum := sha256.Sum256([]byte(repositoryID + ":" + workflowType + ":" + identifier))
	return "wrk_" + hex.EncodeToString(sum[:])
}
