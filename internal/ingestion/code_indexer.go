package ingestion

import "context"

// CodeIndexer indexes Go source into code intelligence storage.
type CodeIndexer interface {
	Index(ctx context.Context, repositoryID, repositoryPath string) error
}
