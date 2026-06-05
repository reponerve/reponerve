package ingestion

import "reponerve/internal/scanner"

// RegisteredScanner pairs a scanner with its name for identification and error attribution.
type RegisteredScanner struct {
	Name    string
	Scanner scanner.SourceScanner
}

// Registry holds the registered scanners.
type Registry struct {
	scanners []RegisteredScanner
}

// NewRegistry creates a new Registry.
func NewRegistry() *Registry {
	return &Registry{
		scanners: make([]RegisteredScanner, 0),
	}
}

// Register registers a named scanner.
func (r *Registry) Register(name string, s scanner.SourceScanner) {
	r.scanners = append(r.scanners, RegisteredScanner{Name: name, Scanner: s})
}

// Scanners returns all registered scanners.
func (r *Registry) Scanners() []RegisteredScanner {
	return r.scanners
}
