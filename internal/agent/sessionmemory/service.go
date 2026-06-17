package sessionmemory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	"github.com/reponerve/reponerve/internal/memory/searchindex"
	memorystorage "github.com/reponerve/reponerve/internal/memory/storage"
	"github.com/reponerve/reponerve/internal/query/storage"
	storagedef "github.com/reponerve/reponerve/internal/storage"
	pkgmodels "github.com/reponerve/reponerve/pkg/models"
)

const sessionSourceType = "session"

// Service manages repository-scoped session memory with evidence provenance.
type Service struct {
	factStore      memorystorage.FactStore
	sourceStore    storagedef.SourceStore
	factReader     storage.FactReader
	eventReader    storage.EventReader
	decisionReader storage.DecisionReader
	searchStore    storagedef.MemorySearchStore
	sourceReader   storage.SourceReader
	accessPath     string
}

// NewService constructs a session memory service.
func NewService(
	factStore memorystorage.FactStore,
	sourceStore storagedef.SourceStore,
	factReader storage.FactReader,
	eventReader storage.EventReader,
	decisionReader storage.DecisionReader,
	searchStore storagedef.MemorySearchStore,
	sourceReader storage.SourceReader,
	workspaceDir string,
) *Service {
	return &Service{
		factStore:      factStore,
		sourceStore:    sourceStore,
		factReader:     factReader,
		eventReader:    eventReader,
		decisionReader: decisionReader,
		searchStore:    searchStore,
		sourceReader:   sourceReader,
		accessPath:     filepathJoin(workspaceDir, "session-access.json"),
	}
}

func filepathJoin(a, b string) string {
	return strings.TrimRight(a, "/") + "/" + b
}

// Remember persists session knowledge as a traceable fact.
func (s *Service) Remember(ctx context.Context, req RememberRequest) (*memorymodels.Fact, error) {
	req.Subject = strings.TrimSpace(req.Subject)
	req.Content = strings.TrimSpace(req.Content)
	if req.RepositoryID == "" {
		return nil, fmt.Errorf("repository ID is required")
	}
	if req.Subject == "" || req.Content == "" {
		return nil, fmt.Errorf("subject and content are required")
	}

	sourceID, err := s.ensureSessionSource(ctx, req.RepositoryID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	fact := &memorymodels.Fact{
		ID:           factID(sourceID, req.Subject, PredicateSessionRemembered, req.Content),
		RepositoryID: req.RepositoryID,
		Subject:      req.Subject,
		Predicate:    PredicateSessionRemembered,
		Object:       req.Content,
		SourceID:     sourceID,
		CreatedAt:    now,
	}
	if err := s.factStore.UpsertFact(ctx, fact); err != nil {
		return nil, err
	}
	if err := s.rebuildSearch(ctx, req.RepositoryID); err != nil {
		return nil, err
	}
	_ = s.recordAccess(req.RepositoryID, fact.ID)
	return fact, nil
}

// WritebackQA stores a question/answer pair as session memory.
func (s *Service) WritebackQA(ctx context.Context, req WritebackRequest) (*memorymodels.Fact, error) {
	req.Question = strings.TrimSpace(req.Question)
	req.Answer = strings.TrimSpace(req.Answer)
	if req.RepositoryID == "" {
		return nil, fmt.Errorf("repository ID is required")
	}
	if req.Question == "" || req.Answer == "" {
		return nil, fmt.Errorf("question and answer are required")
	}
	content := req.Question + " => " + req.Answer
	return s.Remember(ctx, RememberRequest{
		RepositoryID: req.RepositoryID,
		Subject:      req.Question,
		Content:      content,
	})
}

// Forget removes a session fact by ID.
func (s *Service) Forget(ctx context.Context, repositoryID, factID string) error {
	if repositoryID == "" || factID == "" {
		return fmt.Errorf("repository ID and fact ID are required")
	}
	fact, err := s.factReader.GetByID(ctx, factID)
	if err != nil {
		return fmt.Errorf("fact not found: %w", err)
	}
	if fact.RepositoryID != repositoryID {
		return fmt.Errorf("fact does not belong to repository")
	}
	if !isSessionFact(fact, sessionSourceID(repositoryID)) {
		return fmt.Errorf("only session memory facts can be forgotten")
	}
	if err := s.factStore.DeleteFact(ctx, factID); err != nil {
		return err
	}
	return s.rebuildSearch(ctx, repositoryID)
}

// ListSessionFacts returns session facts ranked by access recency.
func (s *Service) ListSessionFacts(ctx context.Context, repositoryID string) ([]*memorymodels.Fact, error) {
	facts, err := s.factReader.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	sourceID := sessionSourceID(repositoryID)
	sessionFacts := make([]*memorymodels.Fact, 0)
	for _, f := range facts {
		if isSessionFact(f, sourceID) {
			sessionFacts = append(sessionFacts, f)
		}
	}
	ids := make([]string, len(sessionFacts))
	for i, f := range sessionFacts {
		ids[i] = f.ID
	}
	ranked := s.accessRanking(repositoryID, ids)
	byID := make(map[string]*memorymodels.Fact, len(sessionFacts))
	for _, f := range sessionFacts {
		byID[f.ID] = f
	}
	out := make([]*memorymodels.Fact, 0, len(sessionFacts))
	for _, id := range ranked {
		if f, ok := byID[id]; ok {
			out = append(out, f)
		}
	}
	return out, nil
}

// ExportHandoff builds a deterministic handoff bundle for agent transfer.
func (s *Service) ExportHandoff(ctx context.Context, repositoryID string) (*HandoffBundle, error) {
	facts, err := s.ListSessionFacts(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(facts))
	for i, f := range facts {
		ids[i] = f.ID
	}
	return &HandoffBundle{
		Version:       HandoffVersion,
		RepositoryID:  repositoryID,
		SessionID:     handoffSessionID(repositoryID),
		ExportedAt:    time.Now().UTC(),
		Facts:         facts,
		AccessRanking: ids,
	}, nil
}

// ImportHandoff restores session facts from a handoff bundle.
func (s *Service) ImportHandoff(ctx context.Context, bundle *HandoffBundle) error {
	if bundle == nil {
		return fmt.Errorf("bundle is nil")
	}
	if bundle.RepositoryID == "" {
		return fmt.Errorf("missing repository ID")
	}
	if bundle.Version != HandoffVersion {
		return fmt.Errorf("unsupported handoff version %q", bundle.Version)
	}
	if _, err := s.ensureSessionSource(ctx, bundle.RepositoryID); err != nil {
		return err
	}
	for _, fact := range bundle.Facts {
		if fact == nil {
			continue
		}
		fact.RepositoryID = bundle.RepositoryID
		fact.SourceID = sessionSourceID(bundle.RepositoryID)
		if err := s.factStore.UpsertFact(ctx, fact); err != nil {
			return err
		}
	}
	if err := s.rebuildSearch(ctx, bundle.RepositoryID); err != nil {
		return err
	}
	return mergeAccessRanking(s.accessPath, bundle.RepositoryID, bundle.AccessRanking)
}

func (s *Service) ensureSessionSource(ctx context.Context, repositoryID string) (string, error) {
	sourceID := sessionSourceID(repositoryID)
	meta, _ := json.Marshal(map[string]string{
		"provenance": "agent_session",
		"kind":       "session_writeback",
	})
	now := time.Now().UTC()
	src := &pkgmodels.Source{
		ID:           sourceID,
		RepositoryID: repositoryID,
		SourceType:   sessionSourceType,
		Reference:    ".reponerve/session",
		Title:        "Agent session memory",
		Timestamp:    now,
		MetadataJSON: string(meta),
	}
	if err := s.sourceStore.UpsertSource(ctx, src); err != nil {
		return "", err
	}
	return sourceID, nil
}

func (s *Service) rebuildSearch(ctx context.Context, repositoryID string) error {
	return searchindex.RebuildFromRepository(
		ctx, repositoryID,
		s.eventReader, s.decisionReader, s.factReader, s.sourceReader, s.searchStore,
	)
}

func isSessionFact(f *memorymodels.Fact, sessionSourceID string) bool {
	if f == nil {
		return false
	}
	return f.SourceID == sessionSourceID ||
		f.Predicate == PredicateSessionRemembered ||
		f.Predicate == PredicateSessionQA
}

func sessionSourceID(repositoryID string) string {
	h := sha256.Sum256([]byte("session:" + repositoryID))
	return "src_session_" + hex.EncodeToString(h[:8])
}

func factID(sourceID, subject, predicate, object string) string {
	h := sha256.Sum256([]byte(sourceID + subject + predicate + object))
	return "fact_" + hex.EncodeToString(h[:])
}

func handoffSessionID(repositoryID string) string {
	h := sha256.Sum256([]byte("handoff:" + repositoryID))
	return "hof_" + hex.EncodeToString(h[:8])
}

// ExportHandoffFile writes a handoff bundle to disk.
func ExportHandoffFile(bundle *HandoffBundle, path string) error {
	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// ImportHandoffFile reads a handoff bundle from disk.
func ImportHandoffFile(path string) (*HandoffBundle, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var bundle HandoffBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return nil, err
	}
	sort.Slice(bundle.Facts, func(i, j int) bool {
		return bundle.Facts[i].ID < bundle.Facts[j].ID
	})
	return &bundle, nil
}
