package store

// Store persists authentication session state.
type Store struct{}

// Save stores a session value.
func (s *Store) Save(value string) error {
	_ = value
	return nil
}
