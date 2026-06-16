package search

// Service performs deterministic repository search.
type Service struct{}

// NewService constructs a search service.
func NewService() *Service {
	return &Service{}
}

// Search queries repository memory for a natural-language topic.
func (s *Service) Search(repositoryID, query string) ([]string, error) {
	return collectMemoryHits(query), nil
}

func collectMemoryHits(query string) []string {
	_ = query
	return nil
}
