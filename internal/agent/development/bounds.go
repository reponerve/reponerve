package development

// Response bounds keep Development Experience outputs agent-consumable on large repos.
const (
	DefaultTokenBudget      = 1500
	MaxRelatedRefs          = 15
	MaxEvidenceItems        = 20
	MaxPlanStartingPoints   = 8
	MaxImpactedAreas        = 15
	MaxRepositoryCodeLinks  = 12
	MaxReuseCandidates      = 15
	MaxShipCheckItems       = 10
	MaxDisciplineChecks     = 12
	MaxChangedFilesPR       = 30
)

// EffectiveTokenBudget returns the budget to apply (default when unset).
func (o OutputOptions) EffectiveTokenBudget() int {
	if o.TokenBudget > 0 {
		return o.TokenBudget
	}
	return DefaultTokenBudget
}

// WithDefaultBudget applies DefaultTokenBudget when budget is zero.
func (o OutputOptions) WithDefaultBudget() OutputOptions {
	if o.TokenBudget <= 0 {
		o.TokenBudget = DefaultTokenBudget
	}
	return o
}

func capEntityRefs(refs []EntityRef, limit int) ([]EntityRef, int) {
	if limit <= 0 || len(refs) <= limit {
		return refs, 0
	}
	out := make([]EntityRef, limit)
	copy(out, refs[:limit])
	return out, len(refs) - limit
}

func capEvidence(items []EvidenceItem, limit int) ([]EvidenceItem, int) {
	if limit <= 0 || len(items) <= limit {
		return items, 0
	}
	out := make([]EvidenceItem, limit)
	copy(out, items[:limit])
	return out, len(items) - limit
}

func capRepoCodeLinks(links []RepositoryCodeLinkRef, limit int) ([]RepositoryCodeLinkRef, int) {
	if limit <= 0 || len(links) <= limit {
		return links, 0
	}
	out := make([]RepositoryCodeLinkRef, limit)
	copy(out, links[:limit])
	return out, len(links) - limit
}
