package agentsession

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const (
	ArtifactTypeContextPackage = "context_package"
	ArtifactTypeSearchResult   = "search_result"

	SessionTypeRepository  = "repository"
	SessionTypeDomain      = "domain"
	SessionTypeContributor = "contributor"

	SourceContext    = "context"
	SourceSearch     = "search"
	SourceDiscovery  = "discovery"
	SourceLearning   = "learning"
	SourceReviewers  = "reviewers"
	SourceChangePlan = "changeplan"
)

const defaultRepositoryIdentifier = "default"

var validArtifactTypes = map[string]bool{
	ArtifactTypeContextPackage: true,
	ArtifactTypeSearchResult:   true,
}

var validSessionTypes = map[string]bool{
	SessionTypeRepository:  true,
	SessionTypeDomain:      true,
	SessionTypeContributor: true,
}

var validSources = map[string]bool{
	SourceContext:    true,
	SourceSearch:     true,
	SourceDiscovery:  true,
	SourceLearning:   true,
	SourceReviewers:  true,
	SourceChangePlan: true,
}

type SessionArtifact struct {
	ArtifactType string          `json:"artifact_type"`
	Source       string          `json:"source"`
	Data         json.RawMessage `json:"data"`
}

type AgentSession struct {
	SessionID    string             `json:"session_id"`
	SessionType  string             `json:"session_type"`
	RepositoryID string             `json:"repository_id"`
	Artifacts    []*SessionArtifact `json:"artifacts"`
}

func ValidateSession(session *AgentSession) error {
	if session == nil {
		return fmt.Errorf("session is nil")
	}
	if session.SessionID == "" {
		return fmt.Errorf("missing session ID")
	}
	if session.SessionType == "" {
		return fmt.Errorf("missing session type")
	}
	if !validSessionTypes[session.SessionType] {
		return fmt.Errorf("unsupported session type %q (must be one of: repository, domain, contributor)", session.SessionType)
	}
	if session.RepositoryID == "" {
		return fmt.Errorf("missing repository ID")
	}
	if len(session.Artifacts) == 0 {
		return fmt.Errorf("session has no artifacts")
	}

	for i, artifact := range session.Artifacts {
		if artifact == nil {
			return fmt.Errorf("artifact %d is nil", i)
		}
		if artifact.ArtifactType == "" {
			return fmt.Errorf("artifact %d: missing artifact type", i)
		}
		if !validArtifactTypes[artifact.ArtifactType] {
			return fmt.Errorf("artifact %d: unsupported artifact type %q (must be one of: context_package, search_result)", i, artifact.ArtifactType)
		}
		if artifact.Source == "" {
			return fmt.Errorf("artifact %d: missing source", i)
		}
		if !validSources[artifact.Source] {
			return fmt.Errorf("artifact %d: unsupported source %q (must be one of: context, search, discovery, learning, reviewers, changeplan)", i, artifact.Source)
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

func buildSessionID(repositoryID string, sessionType string, identifier string) string {
	sum := sha256.Sum256([]byte(repositoryID + ":" + sessionType + ":" + identifier))
	return "ses_" + hex.EncodeToString(sum[:])
}
