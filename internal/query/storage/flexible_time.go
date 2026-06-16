package storage

import (
	"time"

	"github.com/reponerve/reponerve/internal/storage/sqlite"
)

func flexibleTime(t sqlite.FlexibleTime) time.Time {
	return t.Time
}
