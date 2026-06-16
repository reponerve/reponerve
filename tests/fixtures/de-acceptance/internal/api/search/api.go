package search

// Search provides HTTP search entry points (homonym fixture for acceptance tests).
func Search(query string) ([]string, error) {
	return []string{query}, nil
}
