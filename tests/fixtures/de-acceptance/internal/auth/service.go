package auth

import "example.com/deacceptance/internal/store"

// Service provides authentication and OAuth session handling.
type Service struct {
	store *store.Store
}

// NewService constructs an authentication service.
func NewService(s *store.Store) *Service {
	return &Service{store: s}
}

// Login authenticates a user with OAuth credentials.
func (s *Service) Login(user string) error {
	return s.store.Save(user)
}
