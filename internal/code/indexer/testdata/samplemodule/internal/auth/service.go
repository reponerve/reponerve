package auth

import "example.com/samplemodule/internal/store"

type Service struct {
	store *store.Store
}

func NewService(s *store.Store) *Service {
	return &Service{store: s}
}

func (s *Service) Login(user string) error {
	return s.store.Save(user)
}
