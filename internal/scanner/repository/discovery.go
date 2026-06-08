package repository

import (
	"context"
	"github.com/reponerve/reponerve/pkg/models"
)

type Discovery interface {
	Discover(ctx context.Context, path string) (*models.Repository, error)
}
