package repository

import (
	"context"
	"reponerve/pkg/models"
)

type Discovery interface {
	Discover(ctx context.Context, path string) (*models.Repository, error)
}
