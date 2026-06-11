package code

import (
	"crypto/sha256"
	"encoding/hex"
)

// EntityID returns a deterministic identifier for a code entity.
func EntityID(repositoryID, entityType, qualifiedName string) string {
	sum := sha256.Sum256([]byte(repositoryID + ":" + entityType + ":" + qualifiedName))
	return hex.EncodeToString(sum[:])
}

// RelationshipID returns a deterministic identifier for a code relationship.
func RelationshipID(repositoryID, relationshipType, fromEntityID, toEntityID string) string {
	sum := sha256.Sum256([]byte(repositoryID + ":" + relationshipType + ":" + fromEntityID + ":" + toEntityID))
	return hex.EncodeToString(sum[:])
}

// RepositoryCodeLinkID returns a deterministic identifier for a repository-code link.
func RepositoryCodeLinkID(repositoryID, relationshipType, repositoryEntityID, codeEntityID string) string {
	sum := sha256.Sum256([]byte(repositoryID + ":" + relationshipType + ":" + repositoryEntityID + ":" + codeEntityID))
	return hex.EncodeToString(sum[:])
}
