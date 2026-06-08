package query

import (
	memorymodels "github.com/reponerve/reponerve/internal/memory/models"
	models "github.com/reponerve/reponerve/pkg/models"
)

// ContributorTrace provides a complete trace of a contributor's ownership/involvement.
type ContributorTrace struct {
	Contributor *models.Contributor      `json:"contributor"`
	Expertise   []*models.Expertise      `json:"expertise"`
	Decisions   []*memorymodels.Decision `json:"decisions"`
	Facts       []*memorymodels.Fact     `json:"facts"`
	Events      []*models.Event          `json:"events"`
}
