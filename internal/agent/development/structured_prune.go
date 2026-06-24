package development

// PruneReport records fields truncated for agent consumers.
type PruneReport struct {
	Truncated      bool
	TruncatedFields []string
	OmittedCounts  map[string]int
}

// PruneStructured returns a copy of structured DE payloads with list caps applied.
func PruneStructured(structured any) (any, PruneReport) {
	report := PruneReport{OmittedCounts: make(map[string]int)}
	switch v := structured.(type) {
	case *DevelopmentAnswer:
		if v == nil {
			return structured, report
		}
		c := *v
		c.Related, report = pruneRelated(c.Related, report)
		c.Evidence, report = pruneEvidence(c.Evidence, report)
		if c.Plan != nil {
			p, pr := PruneStructured(c.Plan)
			c.Plan = p.(*DevelopmentPlan)
			report = mergePruneReport(report, pr)
		}
		return &c, report
	case DevelopmentAnswer:
		out, r := PruneStructured(&v)
		return *out.(*DevelopmentAnswer), r
	case *DevelopmentPlan:
		if v == nil {
			return structured, report
		}
		c := *v
		c.StartingPoints, report = pruneStartingPoints(c.StartingPoints, report)
		c.ImpactedAreas, report = pruneImpacted(c.ImpactedAreas, report)
		c.RelevantDecisions, report = pruneField(c.RelevantDecisions, "relevant_decisions", MaxImpactedAreas, report)
		c.Evidence, report = pruneEvidence(c.Evidence, report)
		c.RepositoryCodeLinks, report = pruneLinks(c.RepositoryCodeLinks, report)
		return &c, report
	case DevelopmentPlan:
		out, r := PruneStructured(&v)
		return *out.(*DevelopmentPlan), r
	case *DevelopmentExplanation:
		if v == nil {
			return structured, report
		}
		c := *v
		c.Evidence, report = pruneEvidence(c.Evidence, report)
		c.RepositoryCodeLinks, report = pruneLinks(c.RepositoryCodeLinks, report)
		return &c, report
	case DevelopmentExplanation:
		out, r := PruneStructured(&v)
		return *out.(*DevelopmentExplanation), r
	case *DevelopmentImpactReport:
		if v == nil {
			return structured, report
		}
		c := *v
		c.DependentAreas, report = pruneImpacted(c.DependentAreas, report)
		c.Evidence, report = pruneEvidence(c.Evidence, report)
		c.RepositoryCodeLinks, report = pruneLinks(c.RepositoryCodeLinks, report)
		return &c, report
	case DevelopmentImpactReport:
		out, r := PruneStructured(&v)
		return *out.(*DevelopmentImpactReport), r
	case *DevelopmentOnboardingGuide:
		if v == nil {
			return structured, report
		}
		c := *v
		if c.AssignmentPlan != nil {
			p, pr := PruneStructured(c.AssignmentPlan)
			c.AssignmentPlan = p.(*DevelopmentPlan)
			report = mergePruneReport(report, pr)
		}
		return &c, report
	case *ReuseCheckResult:
		if v == nil {
			return structured, report
		}
		c := *v
		c.ReuseCandidates, report = pruneReuseCandidates(c.ReuseCandidates, report)
		c.RelatedDecisions, report = pruneField(c.RelatedDecisions, "related_decisions", MaxImpactedAreas, report)
		c.Evidence, report = pruneEvidence(c.Evidence, report)
		return &c, report
	case ReuseCheckResult:
		out, r := PruneStructured(&v)
		return *out.(*ReuseCheckResult), r
	case *ShipCheckResult:
		if v == nil {
			return structured, report
		}
		c := *v
		c.ImpactedAreas, report = pruneImpacted(c.ImpactedAreas, report)
		c.ShipBlockers, report = pruneShipItems(c.ShipBlockers, "ship_blockers", report)
		c.Advisories, report = pruneShipItems(c.Advisories, "advisories", report)
		c.Evidence, report = pruneEvidence(c.Evidence, report)
		return &c, report
	case ShipCheckResult:
		out, r := PruneStructured(&v)
		return *out.(*ShipCheckResult), r
	default:
		return structured, report
	}
}

func pruneRelated(refs []EntityRef, report PruneReport) ([]EntityRef, PruneReport) {
	capped, omitted := capEntityRefs(refs, MaxRelatedRefs)
	if omitted > 0 {
		report.Truncated = true
		report.TruncatedFields = append(report.TruncatedFields, "related")
		report.OmittedCounts["related"] = omitted
	}
	return capped, report
}

func pruneEvidence(items []EvidenceItem, report PruneReport) ([]EvidenceItem, PruneReport) {
	capped, omitted := capEvidence(items, MaxEvidenceItems)
	if omitted > 0 {
		report.Truncated = true
		report.TruncatedFields = append(report.TruncatedFields, "evidence")
		report.OmittedCounts["evidence"] = omitted
	}
	return capped, report
}

func pruneStartingPoints(refs []EntityRef, report PruneReport) ([]EntityRef, PruneReport) {
	capped, omitted := capEntityRefs(refs, MaxPlanStartingPoints)
	if omitted > 0 {
		report.Truncated = true
		report.TruncatedFields = append(report.TruncatedFields, "starting_points")
		report.OmittedCounts["starting_points"] = omitted
	}
	return capped, report
}

func pruneImpacted(refs []EntityRef, report PruneReport) ([]EntityRef, PruneReport) {
	capped, omitted := capEntityRefs(refs, MaxImpactedAreas)
	if omitted > 0 {
		report.Truncated = true
		report.TruncatedFields = append(report.TruncatedFields, "impacted_areas")
		report.OmittedCounts["impacted_areas"] = omitted
	}
	return capped, report
}

func pruneLinks(links []RepositoryCodeLinkRef, report PruneReport) ([]RepositoryCodeLinkRef, PruneReport) {
	capped, omitted := capRepoCodeLinks(links, MaxRepositoryCodeLinks)
	if omitted > 0 {
		report.Truncated = true
		report.TruncatedFields = append(report.TruncatedFields, "repository_code_links")
		report.OmittedCounts["repository_code_links"] = omitted
	}
	return capped, report
}

func pruneField(refs []EntityRef, field string, limit int, report PruneReport) ([]EntityRef, PruneReport) {
	capped, omitted := capEntityRefs(refs, limit)
	if omitted > 0 {
		report.Truncated = true
		report.TruncatedFields = append(report.TruncatedFields, field)
		report.OmittedCounts[field] = omitted
	}
	return capped, report
}

func pruneReuseCandidates(candidates []ReuseCandidate, report PruneReport) ([]ReuseCandidate, PruneReport) {
	if len(candidates) <= MaxReuseCandidates {
		return candidates, report
	}
	report.Truncated = true
	report.TruncatedFields = append(report.TruncatedFields, "reuse_candidates")
	report.OmittedCounts["reuse_candidates"] = len(candidates) - MaxReuseCandidates
	return candidates[:MaxReuseCandidates], report
}

func pruneShipItems(items []ShipCheckItem, field string, report PruneReport) ([]ShipCheckItem, PruneReport) {
	if len(items) <= MaxShipCheckItems {
		return items, report
	}
	report.Truncated = true
	report.TruncatedFields = append(report.TruncatedFields, field)
	report.OmittedCounts[field] = len(items) - MaxShipCheckItems
	return items[:MaxShipCheckItems], report
}

func mergePruneReport(base, add PruneReport) PruneReport {
	if add.Truncated {
		base.Truncated = true
	}
	base.TruncatedFields = append(base.TruncatedFields, add.TruncatedFields...)
	if base.OmittedCounts == nil {
		base.OmittedCounts = make(map[string]int)
	}
	for k, v := range add.OmittedCounts {
		base.OmittedCounts[k] = v
	}
	return base
}

// ApplyPruneReport augments agent meta when structured fields were capped.
func ApplyPruneReport(meta *AgentContextMeta, report PruneReport) {
	if meta == nil || !report.Truncated {
		return
	}
	meta.Truncated = true
	meta.TruncatedFields = append(meta.TruncatedFields, report.TruncatedFields...)
	meta.PreferNarrowTools = []string{"explain_function", "explain_file", "explain_struct", "explain_feature"}
	if meta.Completeness == CompletenessFull {
		meta.Completeness = CompletenessPartial
	}
	meta.GuidanceForAgent = append(meta.GuidanceForAgent,
		"Response was truncated for token budget. Do not grep the repository — call prefer_narrow_tools for specifics.",
	)
}
