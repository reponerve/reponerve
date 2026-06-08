package query_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/reponerve/reponerve/internal/ownership/query"
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	models "github.com/reponerve/reponerve/pkg/models"
)

type mockContributorReader struct {
	getByIDFn          func(ctx context.Context, repositoryID string, id string) (*models.Contributor, error)
	listByRepositoryFn func(ctx context.Context, repositoryID string) ([]*models.Contributor, error)
}

func (m *mockContributorReader) GetByID(ctx context.Context, repositoryID string, id string) (*models.Contributor, error) {
	return m.getByIDFn(ctx, repositoryID, id)
}

func (m *mockContributorReader) ListByRepository(ctx context.Context, repositoryID string) ([]*models.Contributor, error) {
	return m.listByRepositoryFn(ctx, repositoryID)
}

type mockExpertiseReader struct {
	listByRepositoryFn  func(ctx context.Context, repositoryID string) ([]*models.Expertise, error)
	listByContributorFn func(ctx context.Context, repositoryID string, contributorID string) ([]*models.Expertise, error)
}

func (m *mockExpertiseReader) ListByRepository(ctx context.Context, repositoryID string) ([]*models.Expertise, error) {
	return m.listByRepositoryFn(ctx, repositoryID)
}

func (m *mockExpertiseReader) ListByContributor(ctx context.Context, repositoryID string, contributorID string) ([]*models.Expertise, error) {
	return m.listByContributorFn(ctx, repositoryID, contributorID)
}

type mockSourceReader struct {
	listByRepositoryFn func(ctx context.Context, repositoryID string) ([]*models.Source, error)
}

func (m *mockSourceReader) ListByRepository(ctx context.Context, repositoryID string) ([]*models.Source, error) {
	return m.listByRepositoryFn(ctx, repositoryID)
}

type mockDecisionReader struct {
	listByRepositoryFn func(ctx context.Context, repositoryID string) ([]*memorymodels.Decision, error)
}

func (m *mockDecisionReader) GetByID(ctx context.Context, id string) (*memorymodels.Decision, error) {
	return nil, nil
}
func (m *mockDecisionReader) ListByRepository(ctx context.Context, repositoryID string) ([]*memorymodels.Decision, error) {
	return m.listByRepositoryFn(ctx, repositoryID)
}
func (m *mockDecisionReader) ListAll(ctx context.Context) ([]*memorymodels.Decision, error) {
	return nil, nil
}

type mockFactReader struct {
	listByRepositoryFn func(ctx context.Context, repositoryID string) ([]*memorymodels.Fact, error)
}

func (m *mockFactReader) GetByID(ctx context.Context, id string) (*memorymodels.Fact, error) {
	return nil, nil
}
func (m *mockFactReader) ListByRepository(ctx context.Context, repositoryID string) ([]*memorymodels.Fact, error) {
	return m.listByRepositoryFn(ctx, repositoryID)
}
func (m *mockFactReader) ListAll(ctx context.Context) ([]*memorymodels.Fact, error) {
	return nil, nil
}

type mockEventReader struct {
	listByRepositoryFn func(ctx context.Context, repositoryID string) ([]*models.Event, error)
}

func (m *mockEventReader) GetByID(ctx context.Context, id string) (*models.Event, error) {
	return nil, nil
}
func (m *mockEventReader) ListByRepository(ctx context.Context, repositoryID string) ([]*models.Event, error) {
	return m.listByRepositoryFn(ctx, repositoryID)
}
func (m *mockEventReader) ListAll(ctx context.Context) ([]*models.Event, error) {
	return nil, nil
}

func TestReader_EmptyRepositories(t *testing.T) {
	mCr := &mockContributorReader{
		listByRepositoryFn: func(ctx context.Context, repositoryID string) ([]*models.Contributor, error) {
			return nil, nil
		},
	}
	mEr := &mockExpertiseReader{
		listByRepositoryFn: func(ctx context.Context, repositoryID string) ([]*models.Expertise, error) {
			return nil, nil
		},
	}

	r := query.NewReader(mCr, mEr, &mockSourceReader{}, &mockDecisionReader{}, &mockFactReader{}, &mockEventReader{})
	ctx := context.Background()

	contribs, err := r.ListContributors(ctx, "repo_1")
	if err != nil {
		t.Fatalf("unexpected error listing contributors: %v", err)
	}
	if len(contribs) != 0 {
		t.Errorf("expected 0 contributors, got %d", len(contribs))
	}

	exps, err := r.ListExpertise(ctx, "repo_1")
	if err != nil {
		t.Fatalf("unexpected error listing expertise: %v", err)
	}
	if len(exps) != 0 {
		t.Errorf("expected 0 expertise records, got %d", len(exps))
	}
}

func TestReader_MissingContributor(t *testing.T) {
	mCr := &mockContributorReader{
		getByIDFn: func(ctx context.Context, repositoryID string, id string) (*models.Contributor, error) {
			return nil, sql.ErrNoRows
		},
	}

	r := query.NewReader(mCr, &mockExpertiseReader{}, &mockSourceReader{}, &mockDecisionReader{}, &mockFactReader{}, &mockEventReader{})
	ctx := context.Background()

	_, err := r.GetContributor(ctx, "repo_1", "missing_c")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got: %v", err)
	}

	_, err = r.TraceContributor(ctx, "repo_1", "missing_c")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got: %v", err)
	}
}

func TestReader_ReaderFailures(t *testing.T) {
	expectedErr := errors.New("database breakdown")
	mCr := &mockContributorReader{
		listByRepositoryFn: func(ctx context.Context, repositoryID string) ([]*models.Contributor, error) {
			return nil, expectedErr
		},
	}

	r := query.NewReader(mCr, &mockExpertiseReader{}, &mockSourceReader{}, &mockDecisionReader{}, &mockFactReader{}, &mockEventReader{})
	ctx := context.Background()

	_, err := r.ListContributors(ctx, "repo_1")
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected wrapped database breakdown error, got: %v", err)
	}
}

func TestReader_DeterministicSorting(t *testing.T) {
	// 1. Contributors sorting: Name ascending, ID ascending fallback
	mCr := &mockContributorReader{
		listByRepositoryFn: func(ctx context.Context, repositoryID string) ([]*models.Contributor, error) {
			return []*models.Contributor{
				{ID: "c_3", Name: "Bob"},
				{ID: "c_1", Name: "Alice"},
				{ID: "c_2", Name: "Alice"},
			}, nil
		},
	}

	// 2. Expertise sorting: Domain ascending, ID ascending fallback
	mEr := &mockExpertiseReader{
		listByRepositoryFn: func(ctx context.Context, repositoryID string) ([]*models.Expertise, error) {
			return []*models.Expertise{
				{ID: "exp_3", Domain: "Storage"},
				{ID: "exp_1", Domain: "Authentication"},
				{ID: "exp_2", Domain: "Authentication"},
			}, nil
		},
	}

	r := query.NewReader(mCr, mEr, &mockSourceReader{}, &mockDecisionReader{}, &mockFactReader{}, &mockEventReader{})
	ctx := context.Background()

	contribs, _ := r.ListContributors(ctx, "repo_1")
	if len(contribs) != 3 || contribs[0].ID != "c_1" || contribs[1].ID != "c_2" || contribs[2].ID != "c_3" {
		t.Errorf("incorrect contributor sorting: %+v", contribs)
	}

	exps, _ := r.ListExpertise(ctx, "repo_1")
	if len(exps) != 3 || exps[0].ID != "exp_1" || exps[1].ID != "exp_2" || exps[2].ID != "exp_3" {
		t.Errorf("incorrect expertise sorting: %+v", exps)
	}
}

func TestReader_TraceContributor(t *testing.T) {
	repoID := "repo_1"
	
	testContributorID := func(repositoryID, name, email string) string {
		var input string
		if email != "" {
			input = repositoryID + ":" + email
		} else {
			input = repositoryID + ":" + name
		}
		hash := sha256.Sum256([]byte(input))
		return "ctr_" + hex.EncodeToString(hash[:])
	}
	cID := testContributorID(repoID, "Alice", "alice@example.com")

	mCr := &mockContributorReader{
		getByIDFn: func(ctx context.Context, repositoryID string, id string) (*models.Contributor, error) {
			return &models.Contributor{ID: cID, RepositoryID: repoID, Name: "Alice", Email: "alice@example.com"}, nil
		},
	}
	mEr := &mockExpertiseReader{
		listByContributorFn: func(ctx context.Context, repositoryID string, contributorID string) ([]*models.Expertise, error) {
			return []*models.Expertise{
				{ID: "exp_1", Domain: "Authentication", Score: 1.0},
			}, nil
		},
	}
	mSr := &mockSourceReader{
		listByRepositoryFn: func(ctx context.Context, repositoryID string) ([]*models.Source, error) {
			return []*models.Source{
				{ID: "src_alice_auth", RepositoryID: repoID, Author: "Alice <alice@example.com>"},
				{ID: "src_bob", RepositoryID: repoID, Author: "Bob <bob@example.com>"},
			}, nil
		},
	}

	t1 := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)

	mDr := &mockDecisionReader{
		listByRepositoryFn: func(ctx context.Context, repositoryID string) ([]*memorymodels.Decision, error) {
			return []*memorymodels.Decision{
				{ID: "dec_2", SourceID: "src_alice_auth", CreatedAt: t2},
				{ID: "dec_1", SourceID: "src_alice_auth", CreatedAt: t1},
				{ID: "dec_3", SourceID: "src_bob", CreatedAt: t2},
			}, nil
		},
	}

	mFr := &mockFactReader{
		listByRepositoryFn: func(ctx context.Context, repositoryID string) ([]*memorymodels.Fact, error) {
			return []*memorymodels.Fact{
				{ID: "fact_2", SourceID: "src_alice_auth", Subject: "token authentication"},
				{ID: "fact_1", SourceID: "src_alice_auth", Subject: "credential management"},
				{ID: "fact_3", SourceID: "src_bob", Subject: "database persistence"},
			}, nil
		},
	}

	mEvr := &mockEventReader{
		listByRepositoryFn: func(ctx context.Context, repositoryID string) ([]*models.Event, error) {
			return []*models.Event{
				{ID: "ev_1", SourceID: "src_alice_auth", Timestamp: t1},
				{ID: "ev_2", SourceID: "src_alice_auth", Timestamp: t2},
				{ID: "ev_3", SourceID: "src_bob", Timestamp: t2},
			}, nil
		},
	}

	r := query.NewReader(mCr, mEr, mSr, mDr, mFr, mEvr)
	ctx := context.Background()

	trace, err := r.TraceContributor(ctx, repoID, cID)
	if err != nil {
		t.Fatalf("TraceContributor failed: %v", err)
	}

	if trace.Contributor.Name != "Alice" {
		t.Errorf("incorrect contributor: %+v", trace.Contributor)
	}

	if len(trace.Expertise) != 1 || trace.Expertise[0].Domain != "Authentication" {
		t.Errorf("incorrect expertise: %+v", trace.Expertise)
	}

	// Verify decisions filtering and sorting: dec_2 (t2) should be first, dec_1 (t1) second. dec_3 excluded.
	if len(trace.Decisions) != 2 || trace.Decisions[0].ID != "dec_2" || trace.Decisions[1].ID != "dec_1" {
		t.Errorf("incorrect decisions trace: %+v", trace.Decisions)
	}

	// Verify facts filtering and sorting: fact_1 ("credential management") should be first, fact_2 ("token authentication") second. fact_3 excluded.
	if len(trace.Facts) != 2 || trace.Facts[0].ID != "fact_1" || trace.Facts[1].ID != "fact_2" {
		t.Errorf("incorrect facts trace: %+v", trace.Facts)
	}

	// Verify events filtering and sorting: ev_2 (t2) should be first, ev_1 (t1) second. ev_3 excluded.
	if len(trace.Events) != 2 || trace.Events[0].ID != "ev_2" || trace.Events[1].ID != "ev_1" {
		t.Errorf("incorrect events trace: %+v", trace.Events)
	}
}
