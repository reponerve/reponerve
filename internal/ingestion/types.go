package ingestion

import "time"

// ScanResult contains statistics and results of a repository scan run.
type ScanResult struct {
	RepositoryID   string
	CommitsIndexed int
	ADRsIndexed    int
	Duration       time.Duration
}
